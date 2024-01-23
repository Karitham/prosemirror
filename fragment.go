package prosemirror

import (
	"fmt"
	"slices"

	"github.com/go-json-experiment/json"
)

// Fragment represents a node's collection of child nodes
// its represented as an array of nodes in JSON, but we cache the size of the fragment
// for operational reasons.
type Fragment struct {
	Size    int
	Content []Node
}

func (f Fragment) String() string {
	return fmt.Sprintf("Fragment{Size: %d, Content: %v}", f.Size, f.Content)
}

func (f Fragment) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Content)
}

func (f Fragment) IsZero() bool {
	return len(f.Content) == 0
}

func (f *Fragment) UnmarshalJSON(data []byte) error {
	content := []Node{}
	err := json.Unmarshal(data, &content)
	if err != nil {
		return err
	}

	*f = NewFragment(content...)
	return nil
}

func NewFragment(nodes ...Node) Fragment {
	size := 0
	for _, node := range nodes {
		size += node.NodeSize()
	}

	return Fragment{
		Size:    size,
		Content: nodes,
	}
}

func (f Fragment) replaceChild(index int, n Node) Fragment {
	curr := f.Content[index]
	if curr.eq(n) {
		return f
	}

	content := slices.Clone(f.Content)
	content[index] = n
	return Fragment{
		Size:    f.Size - curr.NodeSize() + n.NodeSize(),
		Content: content,
	}
}

func (f Fragment) Child(index int) *Node {
	return &f.Content[index]
}

func (f Fragment) maybeChild(index int) *Node {
	if index < 0 || index >= len(f.Content) {
		return nil
	}

	return &f.Content[index]
}

func (f Fragment) ChildCount() int {
	return len(f.Content)
}

// cut removes a range of nodes from the fragment, or a range of text
//
// if `to` is -1, it uses the default size of the node
func (f Fragment) cut(from int, to int) Fragment {
	if to == -1 {
		to = f.Size
	}

	if from == 0 && to == f.Size {
		return f
	}

	if to <= from {
		return Fragment{}
	}

	result := make([]Node, 0)
	size := 0
	for i, pos := 0, 0; pos < to; i++ {
		child := f.Content[i]
		end := pos + child.NodeSize()

		if end <= from {
			pos = end
			continue
		}

		if pos < from || end > to {
			if child.IsText() {
				child = child.cut(max(0, from-pos), min(child.textLen(), to-pos))
			} else {
				child = child.cut(max(0, from-pos-1), min(child.Content.Size, to-pos-1))
			}
		}

		result = append(result, child)
		size += child.NodeSize()
		pos = end
	}

	return Fragment{
		Size:    size,
		Content: result,
	}
}

func (f Fragment) append(other Fragment) Fragment {
	if other.Size == 0 {
		return f
	}

	if f.Size == 0 {
		return other
	}

	last := *f.lastChild()
	first := *other.firstChild()
	content := slices.Clone(f.Content)
	i := 0

	if last.IsText() && last.sameMarkup(first) {
		content[len(content)-1] = last.withText(last.Text + first.Text)
		i = 1
	}

	for ; i < len(other.Content); i++ {
		content = append(content, other.Content[i])
	}

	return Fragment{
		Size:    f.Size + other.Size,
		Content: content,
	}
}

func (f Fragment) clone() Fragment {
	return Fragment{
		Size:    f.Size,
		Content: slices.Clone(f.Content),
	}
}

func (f Fragment) findIndex(pos int) (int, int) {
	if pos == 0 {
		return 0, pos
	}

	if pos == f.Size {
		return len(f.Content), pos
	}

	if pos > f.Size || pos < 0 {
		panic(fmt.Sprintf("position out of bounds: %d (of %d)", pos, f.Size))
	}

	for i, curPos := 0, 0; ; i++ {
		cur := f.Child(i)
		if cur == nil {
			panic("nil child")
		}
		end := curPos + cur.NodeSize()

		if end >= pos {
			if end == pos {
				return i + 1, end
			}

			return i, curPos
		}

		curPos = end
	}
}

func (f Fragment) firstChild() *Node {
	if len(f.Content) == 0 {
		return nil
	}

	return &f.Content[0]
}

func (f Fragment) lastChild() *Node {
	if len(f.Content) == 0 {
		return nil
	}

	return &f.Content[len(f.Content)-1]
}

func (f Fragment) nodesBetween(from, to int, fn func(node Node, start int, parent *Node, index int) bool, nodeStart int, parent *Node) {
	for i, pos := 0, 0; pos < to; i++ {
		child := f.Content[i]
		end := pos + child.NodeSize()

		if end < from {
			continue
		}

		if fn(child, nodeStart+pos, parent, i) && child.Content.Size != 0 {
			start := pos + 1
			child.nodesBetween(
				max(0, from-start),
				min(child.Content.Size, to-start),
				fn,
				nodeStart+start,
			)
		}

		pos = end
	}
}
