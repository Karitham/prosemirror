package prosemirror

import (
	"fmt"
	"maps"

	"github.com/go-json-experiment/json"
)

type MarkTypeName string

// MarkType represents a mark on a node.
// Marks are used to represent things like whether a node is emphasized or part of a link.
type MarkType struct {
	// The Name of the mark type.
	Name MarkTypeName
	// The rank of the mark type.
	// Rank int

	// The schema this mark type is part of.
	Schema Schema
	// The spec for this mark type.
	Spec MarkSpec

	// The defined attributes for this mark type.
	Attrs map[string]Attribute

	// Marks excluded by this mark type.
	Excluded []MarkType

	// A mark instance with default attributes.
	Instance *Mark
}

func (mt MarkType) String() string {
	return fmt.Sprintf("%s<%s>", mt.Name, mt.Spec.Group)
}

func (mt MarkType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.Name)
}

func (mt *MarkType) UnmarshalJSON(b []byte) error {
	var name MarkTypeName
	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}

	n2, ok := markTypeStore[name]
	if !ok {
		return fmt.Errorf("unknown node type %q", name)
	}

	*mt = n2
	return nil
}

// Create a new mark of this type.
func (mt MarkType) Create(attrs map[string]any) Mark {
	if attrs == nil && mt.Instance != nil {
		return *mt.Instance
	}

	return Mark{
		Type:  mt,
		Attrs: attrs,
	}
}

// Remove this mark type from a mark set, if present.
func (mt MarkType) RemoveFromSet(set []MarkType) []MarkType {
	for i, m := range set {
		if m.Name == mt.Name {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// Check if this mark type is in a mark set.
func (mt MarkType) IsInSet(set []MarkType) MarkType {
	for _, m := range set {
		if m.Name == mt.Name {
			return m
		}
	}
	return MarkType{}
}

// Check if this mark type excludes another type.
func (mt MarkType) Excludes(other MarkType) bool {
	for _, t := range mt.Excluded {
		if t.Eq(other) {
			return true
		}
	}
	return false
}

func (m MarkType) Eq(other MarkType) bool {
	return m.Name == other.Name &&
		maps.Equal(m.Attrs, other.Attrs)
}

type MarkSpec struct {
	// The attributes this mark can have.
	Attrs map[string]Attribute

	// Whether this mark should be active at its end.
	Inclusive bool

	// Determines which other marks this can coexist with.
	Excludes string

	// The group or groups this mark belongs to.
	Group string

	// Whether this mark can span multiple nodes.
	Spanning bool

	// Additional spec properties.
	Extra map[string]any
}

func NewMarkType(s Schema, name MarkTypeName, spec MarkSpec) MarkType {
	return MarkType{
		Name:   name,
		Schema: s,
		Spec:   spec,
		Attrs:  initAttrs(spec.Attrs),
		// TODO: https://github.com/ProseMirror/prosemirror-model/blob/master/src/schema.ts#L272-L273
	}
}

func compileMarkTypeSet(s Schema, spec map[MarkTypeName]MarkSpec) (map[MarkTypeName]MarkType, error) {
	result := map[MarkTypeName]MarkType{}
	for name, spec := range spec {
		result[name] = NewMarkType(s, name, spec)
	}

	return result, nil
}
