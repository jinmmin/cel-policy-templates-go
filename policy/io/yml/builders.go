// Copyright 2020 Google LLC
//
// Liceed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yml

import (
	"fmt"

	"github.com/google/cel-policy-templates-go/policy/model"
)

// objRef defines a series of methods used to build an object model from the YAML decode step.
type objRef interface {
	// id assigns the relative source element identifier to the object.
	id(id int64)

	// assign a primitive value to the object.
	//
	// If the object is not a primitive value, return an error.
	assign(value interface{}) error

	// prop creates an objRef for the property with the given name for building nested objects.
	//
	// If the object does not have the property or is not a map-like type, the method will return
	// an error.
	prop(id int64, name string) (objRef, error)

	// propAt creates an objRef for the item at the given 'idx' ordinal for building list entries.
	//
	// If the object is not a list or the index is not between 0 and the length of the list, the
	// function will return an error.
	propAt(idx interface{}) (objRef, error)
}

// newBaseBuilder returns a base builder which implements the core methods of the objRef interface.
func newBaseBuilder(typeName string) *baseBuilder {
	return &baseBuilder{typeName: typeName}
}

type baseBuilder struct {
	typeName string
}

// id is an implementation of the objRef interface method.
func (b *baseBuilder) id(id int64) {}

// assign is an implementation of the objRef interface method.
func (b *baseBuilder) assign(val interface{}) error {
	return valueNotAssignableToType(b.typeName, val)
}

// prop is an implementation of the objRef interface method.
func (b *baseBuilder) prop(id int64, name string) (objRef, error) {
	return nil, typeNotAssignableToType(b.typeName, "struct")
}

// propAt is an implementation of the objRef interface method.
func (b *baseBuilder) propAt(idx interface{}) (objRef, error) {
	return nil, typeNotAssignableToType(b.typeName, "list")
}

// newInstanceBuilder produces a builder for a model.Instance object.
func newInstanceBuilder(inst *model.Instance) *instanceBuilder {
	return &instanceBuilder{
		baseBuilder: newBaseBuilder("instance"),
		instance:    inst,
	}
}

type instanceBuilder struct {
	*baseBuilder
	instance *model.Instance
}

// id is an implementation of the objRef interface method.
func (b *instanceBuilder) id(id int64) {
	b.instance.ID = id
}

// prop returns a builder for the model.Instance fields as appropriate.
func (b *instanceBuilder) prop(id int64, name string) (objRef, error) {
	switch name {
	case "version":
		field := &model.StructField{
			ID:  id,
			Ref: &model.DynValue{},
		}
		b.instance.Version = field
		return newDynValueBuilder(field.Ref), nil
	case "description":
		field := &model.StructField{
			ID:  id,
			Ref: &model.DynValue{},
		}
		b.instance.Description = field
		return newDynValueBuilder(field.Ref), nil
	case "kind":
		field := &model.StructField{
			ID:  id,
			Ref: &model.DynValue{},
		}
		b.instance.Kind = field
		return newDynValueBuilder(field.Ref), nil
	case "metadata":
		sv := &model.StructValue{
			Fields: []*model.StructField{},
		}
		field := &model.StructField{
			ID: id,
			Ref: &model.DynValue{
				Value: sv,
			},
		}
		b.instance.Metadata = field
		db := newDynValueBuilder(field.Ref)
		db.sb = newStructBuilder(sv)
		return db, nil
	case "rules":
		lv := &model.ListValue{
			Entries: []*model.DynValue{},
		}
		field := &model.StructField{
			ID: id,
			Ref: &model.DynValue{
				Value: lv,
			},
		}
		b.instance.Rules = field
		db := newDynValueBuilder(field.Ref)
		db.lb = newListBuilder(lv)
		return db, nil
	case "selector":
		b.instance.Selector = &model.Selector{ID: id}
		return newSelectorBuilder(b.instance.Selector), nil
	}
	return nil, noSuchProperty("instance", name)
}

// newSelectorBuilder returns a builder for model.Selector instances.
func newSelectorBuilder(sel *model.Selector) *selectorBuilder {
	return &selectorBuilder{
		baseBuilder: newBaseBuilder("selector"),
		sel:         sel,
	}
}

type selectorBuilder struct {
	*baseBuilder
	sel *model.Selector
}

// prop returns builders for selector matcher fields.
func (b *selectorBuilder) prop(id int64, name string) (objRef, error) {
	switch name {
	case "matchLabels":
		b.sel.MatchLabels = &model.MatchLabels{
			ID:       id,
			Matchers: []*model.LabelMatcher{},
		}
		return newMatchLabelsBuilder(b.sel.MatchLabels), nil
	case "matchExpressions":
		b.sel.MatchExpressions = &model.MatchExpressions{
			ID:       id,
			Matchers: []*model.ExprMatcher{},
		}
		return newMatchExpressionsBuilder(b.sel.MatchExpressions), nil
	default:
		return nil, noSuchProperty("selector", name)
	}
}

// newMatchLabelsBuilder returns a builder for matchLabels.
func newMatchLabelsBuilder(labels *model.MatchLabels) *matchLabelsBuilder {
	return &matchLabelsBuilder{
		baseBuilder: newBaseBuilder("matchLabels"),
		labels:      labels,
	}
}

type matchLabelsBuilder struct {
	*baseBuilder
	labels *model.MatchLabels
}

// prop returns a builder for the key, value pairs expected by the matchLabels object.
func (b *matchLabelsBuilder) prop(id int64, name string) (objRef, error) {
	kv := &model.DynValue{ID: id, Value: model.StringValue(name)}
	val := &model.DynValue{}
	lbl := &model.LabelMatcher{Key: kv, Value: val}
	b.labels.Matchers = append(b.labels.Matchers, lbl)
	return newDynValueBuilder(val), nil
}

// newMatchExpressionsBuilder returns a builder for the list of match expressions which
// perform set-like tests on key values.
func newMatchExpressionsBuilder(exprs *model.MatchExpressions) *matchExpressionsBuilder {
	return &matchExpressionsBuilder{
		baseBuilder: newBaseBuilder("matchExpressions"),
		exprs:       exprs,
	}
}

type matchExpressionsBuilder struct {
	*baseBuilder
	exprs *model.MatchExpressions
}

// propAt returns a builder for a single expression matcher within the matchExpressions list.
func (b *matchExpressionsBuilder) propAt(idx interface{}) (objRef, error) {
	i, ok := idx.(int)
	if !ok {
		return nil, invalidIndexType("int", idx)
	}
	if i < 0 || i > len(b.exprs.Matchers) {
		return nil, indexOutOfRange(idx, len(b.exprs.Matchers))
	}
	m := &model.ExprMatcher{}
	b.exprs.Matchers = append(b.exprs.Matchers, m)
	return newExprMatcherBuilder(m), nil
}

// newExprMatcher returns a builder for a matchExpressions set-like operation.
func newExprMatcherBuilder(m *model.ExprMatcher) *exprMatcherBuilder {
	return &exprMatcherBuilder{
		baseBuilder: newBaseBuilder("exprMatcher"),
		match:       m,
	}
}

type exprMatcherBuilder struct {
	*baseBuilder
	match *model.ExprMatcher
}

// prop implements the objRef interface method and sets the values supported by the
// matchExpressions.
func (b *exprMatcherBuilder) prop(id int64, name string) (objRef, error) {
	switch name {
	case "key":
		b.match.Key = &model.DynValue{ID: id}
		return newDynValueBuilder(b.match.Key), nil
	case "operator":
		b.match.Operator = &model.DynValue{ID: id}
		return newDynValueBuilder(b.match.Operator), nil
	case "values":
		lv := &model.ListValue{Entries: []*model.DynValue{}}
		b.match.Values = &model.DynValue{ID: id, Value: lv}
		db := newDynValueBuilder(b.match.Values)
		db.lb = newListBuilder(lv)
		return db, nil
	default:
		return nil, noSuchProperty("exprMatcher", name)
	}
}

// newStructBuilder returns a builder for dynamic values of struct type.
func newStructBuilder(sv *model.StructValue) *structBuilder {
	return &structBuilder{
		baseBuilder: newBaseBuilder("struct"),
		structVal:   sv,
	}
}

type structBuilder struct {
	*baseBuilder
	structVal *model.StructValue
}

// prop returns a builder for a struct property.
func (b *structBuilder) prop(id int64, name string) (objRef, error) {
	field := &model.StructField{
		ID:   id,
		Name: name,
		Ref:  &model.DynValue{},
	}
	b.structVal.Fields = append(b.structVal.Fields, field)
	return newDynValueBuilder(field.Ref), nil
}

// newListBuilder returns a builder for a dynamic value of list type.
func newListBuilder(lv *model.ListValue) *listBuilder {
	return &listBuilder{
		baseBuilder: newBaseBuilder("list"),
		listVal:     lv,
	}
}

type listBuilder struct {
	*baseBuilder
	listVal *model.ListValue
}

// propAt returns a builder for a list element at the given index.
func (b *listBuilder) propAt(idx interface{}) (objRef, error) {
	dyn := &model.DynValue{}
	b.listVal.Entries = append(b.listVal.Entries, dyn)
	return newDynValueBuilder(dyn), nil
}

// newDynValueBuilder returns a builder for a model.DynValue.
func newDynValueBuilder(dyn *model.DynValue) *dynValueBuilder {
	return &dynValueBuilder{
		dyn: dyn,
	}
}

type dynValueBuilder struct {
	dyn *model.DynValue
	lb  *listBuilder
	sb  *structBuilder
}

// id sets the source element id of the dyn literal.
func (b *dynValueBuilder) id(id int64) {
	b.dyn.ID = id
}

// assign will set the value of the model.DynValue.
//
// If the builder had previously been configured to produce list or struct values, the function
// returns an error.
func (b *dynValueBuilder) assign(val interface{}) error {
	if b.sb != nil {
		return valueNotAssignableToType("struct", val)
	}
	if b.lb != nil {
		return valueNotAssignableToType("list", val)
	}
	var vn model.ValueNode
	switch v := val.(type) {
	case bool:
		vn = model.BoolValue(v)
	case []byte:
		vn = model.BytesValue(v)
	case float64:
		vn = model.DoubleValue(v)
	case int64:
		vn = model.IntValue(v)
	case string:
		vn = model.StringValue(v)
	case uint64:
		vn = model.UintValue(v)
	case model.NullValue:
		vn = v
	default:
		return valueNotAssignableToType("dyn", v)
	}
	b.dyn.Value = vn
	return nil
}

// prop returns a builder for a struct field.
//
// If the dyn builder was previously configured as a list builder, the function will error.
func (b *dynValueBuilder) prop(id int64, name string) (objRef, error) {
	if b.lb != nil {
		return nil, typeNotAssignableToType("list", "struct")
	}
	if b.sb == nil {
		sv := &model.StructValue{
			Fields: []*model.StructField{},
		}
		b.dyn.Value = sv
		b.sb = newStructBuilder(sv)
	}
	return b.sb.prop(id, name)
}

// propAt returns a builder for an element within a list value.
//
// If the dyn builder was previously configured as a struct, this function will error.
func (b *dynValueBuilder) propAt(idx interface{}) (objRef, error) {
	if b.sb != nil {
		return nil, typeNotAssignableToType("struct", "list")
	}
	if b.lb == nil {
		lv := &model.ListValue{
			Entries: []*model.DynValue{},
		}
		b.dyn.Value = lv
		b.lb = newListBuilder(lv)
	}
	return b.lb.propAt(idx)
}

// helper methods for formatting builder-related error messages.

func typeNotAssignableToType(typeName, valType string) error {
	return fmt.Errorf("type not assignable to target: target=%v, type=%s", typeName, valType)
}

func valueNotAssignableToType(typeName string, val interface{}) error {
	return fmt.Errorf("type not assignable to target: target=%s, type=%T", typeName, val)
}

func noSuchProperty(typeName, prop string) error {
	return fmt.Errorf("no such property: type=%s, property=%s", typeName, prop)
}

func indexOutOfRange(idx interface{}, len int) error {
	return fmt.Errorf("index out of range: index=%v, len=%d", idx, len)
}

func invalidIndexType(typeName string, idx interface{}) error {
	return fmt.Errorf("invalid index type: index-type:%s, argument=%T", typeName, idx)
}
