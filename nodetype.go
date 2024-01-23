package prosemirror

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
)

type NodeTypeName string

type Attrs map[string]Attribute

type Attribute struct {
	Default any
}

func (a Attribute) isRequired() bool {
	return a.Default == nil
}

type NodeSpec struct {
	// The content expression for this node.
	Content string

	// The marks that are allowed inside this node.
	Marks *string

	// The group or groups this node belongs to.
	Group string

	// Should be true for inline nodes.
	Inline bool

	// Can be set to true for non-leaf nodes.
	Atom bool

	// The attributes this node can have.
	Attrs map[string]Attribute

	// Determines if this is an important parent node.
	// DefiningAsContext bool

	// Preserve defining parents when possible.
	// DefiningForContent bool

	// Blocks regular editing operations from crossing sides.
	// Isolating bool

	// Arbitrary additional properties.
	Extra map[string]any
}

// NodeType is the type of a slice content element
type NodeType struct {
	Name          NodeTypeName
	Spec          NodeSpec
	Schema        Schema
	Block         bool
	Text          bool
	Groups        []string
	Marks         []MarkType
	Attrs         Attrs
	DefaultAttrs  map[string]any
	ContentMatch  ContentMatch
	InlineContent bool
}

func (n NodeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Name)
}

func (n *NodeType) UnmarshalJSON(b []byte) error {
	var name NodeTypeName
	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}

	n2, ok := nodeTypeStore[name]
	if !ok {
		return fmt.Errorf("unknown node type %q", name)
	}

	*n = n2
	return nil
}

func (n NodeType) Format(s fmt.State, verb rune) {
	fmt.Fprint(s, "<NodeType")
	fmt.Fprint(s, " name=", n.Name)
	fmt.Fprintf(s, " spec=%#v", n.Spec)
	fmt.Fprint(s, " block=", n.Block)
	fmt.Fprint(s, " text=", n.Text)
	fmt.Fprint(s, " groups=", n.Groups)
	fmt.Fprint(s, " marks=", n.Marks)
	fmt.Fprint(s, " attrs=", n.Attrs)
	fmt.Fprint(s, " defaultAttrs=", n.DefaultAttrs)
	fmt.Fprint(s, " contentMatch=", n.ContentMatch)
	fmt.Fprint(s, " inlineContent=", n.InlineContent)
	fmt.Fprint(s, ">")
}

// CreateNodeType creates a node of the given type with the given attributes and content.
func (n NodeType) Create(attrs map[string]any, marks []Mark, content ...Node) (Node, error) {
	if n.isText() {
		return Node{}, fmt.Errorf("cannot create text node through NodeType")
	}

	f := NewFragment(content...)

	if err := n.CheckContent(f); err != nil {
		return Node{}, err
	}

	return Node{
		Type:    n,
		Attrs:   attrs,
		Marks:   marks,
		Content: f,
	}, nil
}

func (n NodeType) isInline() bool {
	return !n.Block
}

func (n NodeType) isTextBlock() bool {
	return n.Block && n.InlineContent
}

func (n NodeType) isLeaf() bool {
	return n.ContentMatch.Empty()
}

func (n NodeType) isText() bool {
	return n.Text
}

func (n NodeType) isAtom() bool {
	return n.isLeaf() || n.Spec.Atom
}

func (t NodeType) hasRequiredAttrs() bool {
	for _, attr := range t.Attrs {
		if attr.isRequired() {
			return true
		}
	}

	return false
}

func (n NodeType) AllowsMarkType(markType MarkType) bool {
	return n.Marks == nil || slices.ContainsFunc(n.Marks, func(other MarkType) bool {
		return other.Name == markType.Name
	})
}

func (t NodeType) compatibleContent(other NodeType) bool {
	return t.Eq(other) || t.ContentMatch.compatible(other.ContentMatch)
}

func (t NodeType) computeAttrs(attrs map[string]any) map[string]any {
	if attrs == nil && t.DefaultAttrs != nil {
		return t.DefaultAttrs
	}
	return computeAttrs(t.Attrs, attrs)
}

func (n NodeType) Eq(other NodeType) bool {
	return n.Name == other.Name
}

func (n NodeType) CheckContent(f Fragment) error {
	result := n.ContentMatch.matchFragment(f, -1, -1)
	// be as descriptive as possible
	if result == nil {
		return fmt.Errorf("content does not match node type %s, no content match found", n.Name)
	}

	if !result.ValidEnd {
		return fmt.Errorf("content does not match node type %s, invalid end position for content %v", n.Name, f)
	}

	for _, child := range f.Content {
		if err := n.validMarks(child.Marks); err != nil {
			return err
		}
	}

	return nil
}

func (n NodeType) validMarks(marks []Mark) error {
	for _, mark := range marks {
		if !n.AllowsMarkType(mark.Type) {
			return fmt.Errorf("mark %s not allowed in node type %s (allowed %v)", mark.Type.Name, n.Name, n.Marks)
		}
	}

	return nil
}

func NewNodeType(name NodeTypeName) (NodeType, error) {
	n, ok := nodeTypeStore[name]
	if !ok {
		return NodeType{}, fmt.Errorf("unknown node type %q", name)
	}

	return n, nil
}

func newNodeType(name NodeTypeName, schema Schema, spec NodeSpec) (NodeType, error) {
	n := NodeType{
		Name:          name,
		Spec:          spec,
		Schema:        schema,
		Groups:        strings.Split(spec.Group, " "),
		Attrs:         initAttrs(spec.Attrs), // TODO: What is that
		DefaultAttrs:  defaultAttrs(spec.Attrs),
		Block:         !spec.Inline && name != "text",
		Text:          name == "text",
		ContentMatch:  ContentMatch{}, // Filled later
		InlineContent: false,          // Filled later
		Marks:         nil,            // Filled later, by default all marks allowed
	}

	return n, nil
}

func compileNodeTypeSet(schema Schema, nodeSet map[NodeTypeName]NodeSpec) (map[NodeTypeName]NodeType, error) {
	out := map[NodeTypeName]NodeType{}
	for name, spec := range nodeSet {
		nodeType, err := newNodeType(name, schema, spec)
		if err != nil {
			return nil, err
		}

		out[name] = nodeType
	}

	topNode := cmp.Or(schema.Spec.TopNode, "doc")
	if out[topNode].Eq(NodeType{}) {
		return nil, fmt.Errorf("no top level node %q defined in node set", topNode)
	}

	if out["text"].Eq(NodeType{}) {
		return nil, fmt.Errorf("no text node type defined in node set")
	}

	if len(out["text"].Attrs) > 0 {
		return nil, fmt.Errorf("text node type should not have attributes")
	}

	return out, nil
}

func computeAttrs(attrs Attrs, value map[string]any) map[string]any {
	built := map[string]any{}

	for name := range attrs {
		given := value[name]
		if given != nil {
			built[name] = given
			continue
		}

		attr := attrs[name]
		if attr.Default == nil {
			// Skill issue
			panic("No value supplied for attribute " + name)
		}

		built[name] = attr.Default
	}

	return built
}

// TODO: either of these is wrong
func initAttrs(attrs Attrs) Attrs {
	out := Attrs{}
	for k, v := range attrs {
		out[k] = v
	}

	return out
}

func defaultAttrs(attrs Attrs) map[string]any {
	out := map[string]any{}
	for k, v := range attrs {
		out[k] = v.Default
	}

	return out
}
