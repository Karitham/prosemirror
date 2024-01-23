package prosemirror

import (
	"fmt"
)

// resolve resolves the position within the node's content to a position in the document.
func resolve(doc Node, pos int) (ResolvedPos, error) {
	if pos < 0 || pos > doc.Content.Size {
		return ResolvedPos{}, fmt.Errorf("position out of bounds: %d (of %d)", pos, doc.Content.Size)
	}

	r := ResolvedPos{
		Pos:          pos,
		ParentOffset: pos,
		NodePath:     []Node{},
		IndexPath:    []int{},
		OffsetPath:   []int{},
	}
	start := 0

	for node := doc; ; {
		index, offset := node.Content.findIndex(r.ParentOffset)

		r.NodePath = append(r.NodePath, node)
		r.IndexPath = append(r.IndexPath, index)
		r.OffsetPath = append(r.OffsetPath, start+offset)

		rem := r.ParentOffset - offset
		if rem == 0 {
			break
		}

		node = *node.Child(index)
		if node.IsText() {
			break
		}

		r.ParentOffset = rem - 1
		start += offset + 1
	}

	r.Depth = len(r.NodePath) - 1
	return r, nil
}

type ResolvedPos struct {
	Depth        int
	Pos          int
	ParentOffset int

	// in the original JS implementation, paths is a any[] type
	// they index it to retrieve the right one.
	// nodepath is *3
	NodePath []Node
	// indexpath is *3+1
	IndexPath []int
	// offsetpath is *3+2 or *3-1
	OffsetPath []int
}

func (r ResolvedPos) String() string {
	return fmt.Sprintf("ResolvedPos{Pos: %d, ParentOffset: %d, IndexPath: %v, OffsetPath: %v}", r.Pos, r.ParentOffset, r.IndexPath, r.OffsetPath)
}

func (r ResolvedPos) Node(depth int) Node {
	return r.NodePath[r.resolveDepth(depth)]
}

func (r ResolvedPos) Index(depth int) int {
	return r.IndexPath[r.resolveDepth(depth)]
}

// The index pointing after this position into the ancestor at the
// given level.
func (r ResolvedPos) IndexAfter(depth int) int {
	depth = r.resolveDepth(depth)
	i := r.Index(depth)
	if depth != r.Depth || r.TextOffset() {
		i += 1
	}

	return i
}

// The Parent node that the position points into. Note that even if
// a position points into a text node, that node is not considered
// the Parent—text nodes are ‘flat’ in this model, and have no content.
func (r ResolvedPos) Parent() Node {
	return r.Node(r.Depth)
}

// When this position points into a text node, this returns the
// distance between the position and the start of the text node.
// Will be zero for positions that point between nodes.
func (r ResolvedPos) TextOffset() bool {
	return r.Pos-r.OffsetPath[len(r.OffsetPath)-1] != 0
}

// Get the node directly after the position, if any. If the position
// points into a text node, only the part of that node after the
// position is returned.
func (r ResolvedPos) NodeAfter() *Node {
	parent := r.Parent()
	index := r.Index(r.Depth)
	if index == parent.ChildCount() {
		return nil
	}

	dOff := r.Pos - r.OffsetPath[len(r.OffsetPath)-1]
	if dOff != 0 {
		n := parent.Child(index).cut(dOff, -1)
		return &n
	}

	return parent.Child(index)
}

// Get the node directly before the position, if any. If the
// position points into a text node, only the part of that node
// before the position is returned.
func (r ResolvedPos) NodeBefore() *Node {
	index := r.Index(r.Depth)

	dOff := r.Pos - r.OffsetPath[len(r.OffsetPath)-1]
	if dOff != 0 {
		n := r.Parent().Child(index).cut(0, dOff)
		return &n
	}

	if index == 0 {
		return nil
	}

	return r.Parent().Child(index - 1)
}

func (r ResolvedPos) SharedDepth(pos int) int {
	for depth := r.Depth; depth > 0; depth-- {
		if r.Start(depth) <= pos && r.End(depth) >= pos {
			return depth
		}
	}

	return 0
}

func (r ResolvedPos) Start(depth int) int {
	depth = r.resolveDepth(depth)
	if depth == 0 {
		return 0
	}

	// Weirdly complex in the original impl. This actually queries the *parent* index path.
	// https://github.com/ProseMirror/prosemirror-model/blob/a37b6b3adeb548dc9822211b680ce9d31be65842/src/resolvedpos.ts#L65
	last := len(r.OffsetPath) - 1
	return r.OffsetPath[last-1] + 1
}

func (r ResolvedPos) End(depth int) int {
	depth = r.resolveDepth(depth)
	return r.Start(depth) + r.Node(depth).Content.Size
}

func (r ResolvedPos) resolveDepth(depth int) int {
	if depth < 0 {
		return r.Depth + depth
	}

	return depth
}

func joinable(before, after ResolvedPos, depth int) (*Node, error) {
	node := before.Node(depth)
	err := checkJoin(node, after.Node(depth))
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func checkJoin(main, sub Node) error {
	if !sub.Type.compatibleContent(main.Type) {
		return fmt.Errorf("can't join incompatible nodes (%s onto %s)", sub.Type.Name, main.Type.Name)
	}

	return nil
}
