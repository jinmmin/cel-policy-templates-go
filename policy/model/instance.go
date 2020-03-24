// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

// Package model contains abstract representations of policy template and instance objects.
package model

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// Instance declares the properties common to all policy instances.
//
// Each instance must specify its version and kind, as well as its name in the metadata field.
// An instance's kind value indicates the name of the policy template used to validate and evaluate
// the instance.
//
// Each instance may contain zero or more rule values within the Rules block. The format of the
// Rule is defined within the template rule schema and validated on a per-rule basis by the logic
// contained within the template indicated the instance kind.
//
// As a general reference the Instance object fields are patterned after those commonly found in
// Kubernetes resources; however, the Instance is not intended to be Kubernetes specific:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#resources
type Instance struct {
	// Version indicates the version of the Policy Template that backs the instance.
	// Type: string
	Version *StructField

	// Kind identifies the Policy Template used to validate and evaluate this instance.
	// Type: string
	Kind *StructField

	// Metadata contains information which uniquely identifies the instance.
	// The metadata struct must contain a 'name' or 'uid' property. A 'namespace' property may
	// optionally be set, otherwise the namespace is 'default'.
	// Type: struct
	Metadata *StructField

	// Description is a human-readable string describing the behavior of the Policy Instance.
	// Type: string
	Description *StructField

	// Selector is a Kubernetes-style selector which can be used to match simple key-value equality
	// or key-value set relations to determine whether to evaluate the rules contained within the
	// Instance. Optional.
	Selector *Selector

	// Rules contains a list of structured objects that describe inputs to policy evaluation logic
	// declared in the template. The rule schema is also defined within the Template. Optional.
	// Type: list
	Rules *StructField

	// ID specifying the source element id fo the instance within a model.Source.
	ID int64

	// SourceInfo contains metadata about source positions, line offsets, and comments.
	// The object is useful for debug purposes, but is not required to be available at evaluation
	// time.
	SourceInfo *SourceInfo
}

// Selector describes a set of matching rules that determine whether the policy instance should be
// evaluated.
type Selector struct {
	ID               int64
	MatchLabels      *MatchLabels
	MatchExpressions *MatchExpressions
}

// MatchLabels describes a key-value equality matching rule applied to a conceptual label set
// associated with the policy evaluation context.
type MatchLabels struct {
	ID       int64
	Matchers []*LabelMatcher
}

// LabelMatcher describes a single key-value equality match rule.
type LabelMatcher struct {
	Key   *DynValue
	Value *DynValue
}

// MatchExpressions describes a key-values set-relation rule applied to a conceptual label set
// associated with the policy evaluation context.
type MatchExpressions struct {
	ID       int64
	Matchers []*ExprMatcher
}

// ExprMatcher describes the key-values set relation according to an operator.
//
// Note, the supported operators are: In, NotIn, Exists
type ExprMatcher struct {
	Key      *DynValue
	Operator *DynValue
	Values   *DynValue
}

// DynValue is a dynamically typed value used to describe unstructured content.
// Whether the value has the desired type is determined by where it is used within the Instance or
// Template, and whether there are schemas which might enforce a more rigid type definition.
type DynValue struct {
	ID    int64
	Value ValueNode
}

// ValueNode is a marker interface used to indicate which value types may populate a DynValue's
// Value field.
type ValueNode interface {
	isValueNode()
}

// StructValue declares an object with a set of named fields whose values are dynamically typed.
type StructValue struct {
	Fields []*StructField
}

func (*StructValue) isValueNode() {}

// StructField specifies a field name and a reference to a dynamic value.
type StructField struct {
	ID   int64
	Name string
	Ref  *DynValue
}

// ListValue contains a list of dynamically typed entries.
type ListValue struct {
	Entries []*DynValue
}

func (*ListValue) isValueNode() {}

// BoolValue is a boolean value suitable for use within DynValue objects.
type BoolValue bool

func (BoolValue) isValueNode() {}

// BytesValue is a []byte value suitable for use within DynValue objects.
type BytesValue []byte

func (BytesValue) isValueNode() {}

// DoubleValue is a float64 value suitable for use within DynValue objects.
type DoubleValue float64

func (DoubleValue) isValueNode() {}

// IntValue is an int64 value suitable for use within DynValue objects.
type IntValue int64

func (IntValue) isValueNode() {}

// NullValue is a protobuf.Struct concrete null value suitable for use within DynValue objects.
type NullValue structpb.NullValue

func (NullValue) isValueNode() {}

// StringValue is a string value suitable for use within DynValue objects.
type StringValue string

func (StringValue) isValueNode() {}

// UintValue is a uint64 value suitable for use within DynValue objects.
type UintValue uint64

func (UintValue) isValueNode() {}

// Null is a singleton NullValue instance.
var Null NullValue
