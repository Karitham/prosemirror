package prosemirror

import (
	"fmt"
	"strings"
)

type SchemaSpec struct {
	// The node types in this schema. Maps names to
	// NodeSpec objects that describe the node type
	// associated with that name.
	// The order in which they occur in the list is significant.
	Nodes map[NodeTypeName]NodeSpec

	// The mark types that exist in this schema.
	Marks map[MarkTypeName]MarkSpec

	// The name of the default top-level node for the schema.
	TopNode NodeTypeName

	// Don't register the schema in the global schema store.
	DontRegister bool
}

type Schema struct {
	Spec SchemaSpec

	TopNodeType *NodeType
	Nodes       map[NodeTypeName]NodeType
	Marks       map[MarkTypeName]MarkType
}

func (s Schema) Node(typ NodeTypeName, attrs map[string]any, content Fragment, marks ...Mark) Node {
	return Node{
		Type:    s.Nodes[typ],
		Attrs:   attrs,
		Content: content,
		Marks:   marks,
	}
}

func (s Schema) Text(text string, marks ...Mark) Node {
	return Node{
		Type:  s.Nodes["text"],
		Marks: marks,
		Text:  text,
	}
}

func (s Schema) Mark(typ MarkTypeName, attrs map[string]any) Mark {
	return s.Marks[typ].Create(attrs)
}

func compileContentMatch(typ *NodeType, schema Schema, contentExprCache map[string]ContentMatch) error {
	switch {
	default:
		typ.Marks = nil
	case typ.Spec.Marks == nil || *typ.Spec.Marks == "_":
		typ.Marks = nil
	case *typ.Spec.Marks != "":
		typ.Marks = []MarkType{}
		for _, markName := range strings.Split(*typ.Spec.Marks, " ") {
			if markName == "" {
				continue
			}

			typ.Marks = append(typ.Marks, schema.Marks[MarkTypeName(markName)])
		}
	case *typ.Spec.Marks == "" || !typ.InlineContent:
		typ.Marks = []MarkType{}
	}

	if ce, ok := contentExprCache[typ.Spec.Content]; ok {
		typ.ContentMatch = ce
		return nil
	}

	if typ.Spec.Content == "" {
		return nil
	}

	cm, err := parseNodespecContent(typ.Spec.Content, schema.Nodes)
	if err != nil {
		return fmt.Errorf("error parsing content for node %q: %w", typ, err)
	}

	ce := dfa(nfa(cm))

	contentExprCache[typ.Spec.Content] = *ce
	typ.ContentMatch = *ce
	typ.InlineContent = ce.InlineContent()
	return nil
}

func NewSchema(spec SchemaSpec) (Schema, error) {
	s := Schema{Spec: spec}

	nodes, err := compileNodeTypeSet(s, spec.Nodes)
	if err != nil {
		return Schema{}, fmt.Errorf("error compiling nodes: %w", err)
	}

	marks, err := compileMarkTypeSet(s, spec.Marks)
	if err != nil {
		return Schema{}, fmt.Errorf("error compiling marks: %w", err)
	}

	s.Nodes = nodes
	s.Marks = marks

	contentExprCache := map[string]ContentMatch{}
	for k := range nodes {
		node := nodes[k]
		err := compileContentMatch(&node, s, contentExprCache)
		if err != nil {
			return Schema{}, err
		}

		nodes[k] = node
	}

	topnodeT := s.Nodes[spec.TopNode]
	s.TopNodeType = &topnodeT

	if !spec.DontRegister {
		RegisterSchema(s)
	}

	return s, nil
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
