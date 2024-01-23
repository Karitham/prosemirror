package prosemirror

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/go-json-experiment/json"
)

// A node is a nested object which represents a fragment of a document. It roughly corresponds to an HTML element.
// Each node type has an associated string, known as its name.
// It may have any number of attributes, represented as a map from strings to arbitrary values.
//   - If the node type is text, it may have a string content property.
//   - If the node type is a non-leaf node, it may have a content property containing an array of child nodes.
//   - If the node type has inline content, it may have a marks property containing an array of marks.
//   - If the node type is a block node, it may have a attrs property containing a map of additional HTML attributes for the element.
//
// The size of a node is either n.textLen(), 1, or the sum of the sizes of its children.
type Node struct {
	// The type of this node.
	Type NodeType `json:"type,omitempty"`

	// A map of attribute names to values.
	// The kind of attributes depends on the node type.
	Attrs map[string]any `json:"attrs,omitempty"`

	// The marks (things like whether it is emphasized or part of a link) applied to this node.
	Marks []Mark `json:"marks,omitempty"`

	// For text nodes, this contains the node's text content.
	Text string `json:"text,omitempty"`

	// The child node at the given index.
	Content Fragment `json:"content,omitempty,omitzero"`
}

func (n Node) Replace(from, to int, s Slice) (Node, error) {
	fromNode, err := resolve(n, from)
	if err != nil {
		return Node{}, fmt.Errorf("error resolving from: %w", err)
	}

	toNode, err := resolve(n, to)
	if err != nil {
		return Node{}, fmt.Errorf("error resolving to: %w", err)
	}

	return replace(fromNode, toNode, s)
}

func (n Node) Resolve(pos int) (ResolvedPos, error) {
	return resolve(n, pos)
}

func (n Node) eq(other Node) bool {
	if n.Type.isText() {
		return n.sameMarkup(other) && n.Text == other.Text
	}

	return n.Type.Eq(other.Type) &&
		maps.Equal(n.Attrs, other.Attrs) &&
		slices.EqualFunc(n.Marks, other.Marks, func(a, b Mark) bool {
			return a.Eq(b)
		}) &&
		n.Text == other.Text &&
		slices.EqualFunc(n.Content.Content, other.Content.Content, func(a, b Node) bool {
			return a.eq(b)
		})
}

func (n Node) String() string {
	b := &strings.Builder{}

	b.WriteString("Node{")
	fmt.Fprintf(b, "Type: %s", n.Type)

	if len(n.Attrs) > 0 {
		fmt.Fprintf(b, ", Attrs: %v", n.Attrs)
	}
	if len(n.Marks) > 0 {
		fmt.Fprintf(b, ", Marks: %v", n.Marks)
	}
	if n.Text != "" {
		fmt.Fprintf(b, ", Text: %q", n.Text)
	}
	if n.Content.Size > 0 {
		fmt.Fprintf(b, ", Content: %v", n.Content)
	}

	b.WriteString("}")
	return b.String()
}

func (n Node) NodeSize() int {
	if n.IsText() {
		return n.textLen()
	}

	if n.IsLeaf() {
		return 1
	}

	return 2 + n.Content.Size
}

func (n Node) Child(index int) *Node {
	return n.Content.Child(index)
}

func (n Node) MaybeChild(index int) *Node {
	return n.Content.maybeChild(index)
}

func (n Node) FirstChild() *Node {
	return n.Content.firstChild()
}

func (n Node) IsText() bool {
	return n.Type.isText()
}

func (n Node) IsLeaf() bool {
	return n.Type.isLeaf()
}

func (n Node) IsAtom() bool {
	return n.Type.isAtom()
}

func (n Node) IsInline() bool {
	return n.Type.isInline()
}

func (n Node) WithMarks(m []Mark) Node {
	n = n.Clone()
	n.Marks = m
	return n
}

// Cut out the part of the document between the given positions, and
// return it as a `Slice` object.
func (n Node) Slice(from, to int, includeParents bool) (Slice, error) {
	if to == -1 {
		to = n.Content.Size
	}

	if from == to {
		return Slice{}, nil
	}

	fromNode, err := n.Resolve(from)
	if err != nil {
		return Slice{}, fmt.Errorf("error resolving from: %w", err)
	}

	toNode, err := n.Resolve(to)
	if err != nil {
		return Slice{}, fmt.Errorf("error resolving to: %w", err)
	}

	depth := 0
	if !includeParents {
		depth = fromNode.SharedDepth(to)
	}

	start := fromNode.Start(depth)
	node := fromNode.Node(depth)

	content := node.Content.cut(fromNode.Pos-start, toNode.Pos-start)

	return Slice{
		Content:   content,
		OpenStart: fromNode.Depth - depth,
		OpenEnd:   toNode.Depth - depth,
	}, nil
}

func (n Node) close(f Fragment) (Node, error) {
	if err := n.Type.CheckContent(f); err != nil {
		return Node{}, err
	}

	return n.copy(f), nil
}

func (n Node) ChildCount() int {
	return n.Content.ChildCount()
}

func (n Node) MarshalJSON() ([]byte, error) {
	type StringNode struct {
		Type  NodeType       `json:"type"`
		Attrs map[string]any `json:"attrs,omitempty"`
		Marks []Mark         `json:"marks,omitempty"`
		Text  string         `json:"text"`
	}

	type FragmentNode struct {
		Type    NodeType       `json:"type"`
		Attrs   map[string]any `json:"attrs,omitempty"`
		Marks   []Mark         `json:"marks,omitempty"`
		Content Fragment       `json:"content,omitempty,omitzero"`
	}

	if n.Type.isText() {
		return json.Marshal(StringNode{
			Type:  n.Type,
			Attrs: n.Attrs,
			Marks: n.Marks,
			Text:  n.Text,
		})
	}

	return json.Marshal(FragmentNode{
		Type:    n.Type,
		Attrs:   n.Attrs,
		Marks:   n.Marks,
		Content: n.Content,
	})
}

// Test whether replacing the range between `from` and `to` (by
// child index) with the given replacement fragment (which defaults
// to the empty fragment) would leave the node's content valid. You
// can optionally pass `start` and `end` indices into the
// replacement fragment.
func (n Node) canReplace(from, to int, replacement Fragment, start, end int) bool {
	if start == -1 {
		start = 0
	}
	if end == -1 {
		end = replacement.ChildCount()
	}

	one := n.contentMatchAt(from).matchFragment(replacement, start, end)
	if one == nil {
		return false
	}

	two := one.matchFragment(n.Content, to, -1)

	if two == nil || !two.ValidEnd {
		return false
	}

	for i := start; i < end; i++ {
		if err := n.Type.validMarks(replacement.Child(i).Marks); err != nil {
			return false
		}
	}

	return true
}

func (n Node) contentMatchAt(index int) *ContentMatch {
	return n.Type.ContentMatch.matchFragment(n.Content, 0, index)
}

func (n Node) cut(from int, to int) Node {
	if n.IsText() {
		if to == -1 {
			to = n.textLen()
		}

		if from == 0 && to == n.textLen() {
			return n
		}

		return n.withText(n.Text[from:to])
	}

	if to == -1 {
		to = n.Content.Size
	}

	if from == 0 && to == n.Content.Size {
		return n
	}

	return n.copy(n.Content.cut(from, to))
}

func (n Node) withText(s string) Node {
	if n.Text == s {
		return n
	}

	return Node{
		Type:    n.Type,
		Text:    s,
		Attrs:   maps.Clone(n.Attrs),
		Marks:   slices.Clone(n.Marks),
		Content: n.Content.clone(),
	}
}

func (n Node) copy(f Fragment) Node {
	return Node{
		Type:    n.Type,
		Text:    n.Text,
		Attrs:   maps.Clone(n.Attrs),
		Marks:   slices.Clone(n.Marks),
		Content: f.clone(),
	}
}

func (n Node) Clone() Node {
	return Node{
		Type:    n.Type,
		Attrs:   maps.Clone(n.Attrs),
		Marks:   slices.Clone(n.Marks),
		Text:    n.Text,
		Content: n.Content.clone(),
	}
}

func (n Node) hasMarkup(t NodeType, attrs map[string]any, marks []Mark) bool {
	return n.Type.Eq(t) &&
		maps.Equal(n.Attrs, attrs) &&
		slices.EqualFunc(n.Marks, marks, func(a, b Mark) bool {
			return a.Eq(b)
		})
}

func (n Node) sameMarkup(other Node) bool {
	return n.hasMarkup(other.Type, other.Attrs, other.Marks)
}

// Call the given callback for every descendant node. Doesn't
// descend into a node when the callback returns `false`.
func (n Node) Descendants(f func(n Node, pos int, parent *Node, index int) bool) {
	n.nodesBetween(0, n.Content.Size, f, 0)
}

func (n Node) nodesBetween(from int, to int, f func(node Node, start int, parent *Node, index int) bool, startPos int) {
	n.Content.nodesBetween(from, to, f, startPos, &n)
}

// Javascript uses UTF-16 for code points, so I gotta find the length of a string in UTF-16 bytes, not runes.
func (n Node) textLen() int {
	return utf16Len(n.Text)
}
