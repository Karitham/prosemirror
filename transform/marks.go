package transform

import (
	"fmt"
	"slices"

	"github.com/go-json-experiment/json"

	"github.com/karitham/prosemirror"
)

var _ Applier = (*MarkStep)(nil)

func init() {
	RegisterTransformer("addMark", func() Applier {
		return &MarkStep{}
	})
	RegisterTransformer("removeMark", func() Applier {
		return &MarkStep{
			remove: true,
		}
	})
}

type MarkStep struct {
	BaseStep
	Mark prosemirror.Mark `json:"mark"`

	// internal
	// toggle whether it's an add or remove
	remove bool `json:"-"`
}

func (s *MarkStep) String() string {
	return fmt.Sprintf("MarkStep{Type: %s, From: %d, To: %d, Mark: %v}", s.Type, s.From, s.To, s.Mark)
}

func NewAddMarkStep(from, to int, mark prosemirror.Mark) *MarkStep {
	return &MarkStep{
		BaseStep: BaseStep{
			Type: "addMark",
			From: from,
			To:   to,
		},
		Mark: mark,
	}
}

func NewRemoveMarkStep(from, to int, mark prosemirror.Mark) *MarkStep {
	return &MarkStep{
		BaseStep: BaseStep{
			Type: "removeMark",
			From: from,
			To:   to,
		},
		Mark:   mark,
		remove: true,
	}
}

func (s *MarkStep) UnmarshalJSON(data []byte) error {
	type a MarkStep

	// important since the constructor sets defaults
	aux := a(*s)

	if err := json.Unmarshal(data, &aux, json.RejectUnknownMembers(true)); err != nil {
		return fmt.Errorf("failed to decode add mark step (%s): %w", string(data), err)
	}

	*s = MarkStep(aux)
	return nil
}

func (s *MarkStep) MarshalJSON() ([]byte, error) {
	type a MarkStep
	aux := a(*s)

	return json.Marshal(aux)
}

func mapFragments(f prosemirror.Fragment, fn func(node, parent prosemirror.Node, i int) prosemirror.Node, parent prosemirror.Node) prosemirror.Fragment {
	content := []prosemirror.Node{}
	for i := 0; i < f.ChildCount(); i++ {
		child := f.Child(i).Clone()

		if child.Content.Size > 0 {
			child.Content = mapFragments(child.Content, fn, child)
		}

		if child.IsInline() {
			child = fn(child, parent, i)
		}

		content = append(content, child)
	}

	return prosemirror.NewFragment(content...)
}

func (s *MarkStep) Apply(doc prosemirror.Node) (prosemirror.Node, error) {
	if s.remove {
		return s.removeMark(doc)
	}

	return s.addMark(doc)
}

func (s *MarkStep) addMark(doc prosemirror.Node) (prosemirror.Node, error) {
	oldSlice, err := doc.Slice(s.From, s.To, false)
	if err != nil {
		return doc, fmt.Errorf("failed to slice document: %w", err)
	}

	from, err := doc.Resolve(s.From)
	if err != nil {
		return doc, fmt.Errorf("failed to resolve document: %w", err)
	}

	parent := from.Node(from.SharedDepth(s.To))

	slice := mapFragments(oldSlice.Content, func(node prosemirror.Node, parent prosemirror.Node, i int) prosemirror.Node {
		if !node.IsAtom() || !parent.Type.AllowsMarkType(s.Mark.Type) {
			return node
		}

		return node.WithMarks(append(node.Marks, s.Mark))
	}, parent)

	return doc.Replace(s.From, s.To, prosemirror.Slice{
		Content:   slice,
		OpenStart: oldSlice.OpenStart,
		OpenEnd:   oldSlice.OpenEnd,
	})
}

func (s *MarkStep) removeMark(doc prosemirror.Node) (prosemirror.Node, error) {
	oldSlice, err := doc.Slice(s.From, s.To, false)
	if err != nil {
		return doc, fmt.Errorf("failed to slice document: %w", err)
	}

	slice := mapFragments(oldSlice.Content, func(node, parent prosemirror.Node, i int) prosemirror.Node {
		return node.WithMarks(slices.DeleteFunc(node.Marks, func(m prosemirror.Mark) bool {
			return m.Eq(s.Mark)
		}))
	}, doc)

	return doc.Replace(s.From, s.To, prosemirror.Slice{
		Content:   slice,
		OpenStart: oldSlice.OpenStart,
		OpenEnd:   oldSlice.OpenEnd,
	})
}
