package transform

import (
	"fmt"

	"github.com/go-json-experiment/json"

	"github.com/karitham/prosemirror"
)

func init() {
	RegisterTransformer("replaceAround", func() Applier {
		return new(ReplaceAroundStep)
	})
}

type ReplaceAroundStep struct {
	BaseStep
	GapFrom   int               `json:"gapFrom"`
	GapTo     int               `json:"gapTo"`
	Insert    int               `json:"insert"`
	Slice     prosemirror.Slice `json:"slice"`
	Structure bool              `json:"structure"`
}

func (s *ReplaceAroundStep) UnmarshalJSON(data []byte) error {
	type a ReplaceAroundStep
	aux := a{}

	if err := json.Unmarshal(data, &aux, json.RejectUnknownMembers(true)); err != nil {
		return fmt.Errorf("failed to decode add mark step (%s): %w", string(data), err)
	}

	*s = ReplaceAroundStep(aux)
	return nil
}

func (s *ReplaceAroundStep) MarshalJSON() ([]byte, error) {
	type a ReplaceAroundStep
	aux := a(*s)

	return json.Marshal(aux)
}

func (s *ReplaceAroundStep) Apply(doc prosemirror.Node) (prosemirror.Node, error) {
	if s.Structure && (contentBetween(doc, s.From, s.GapFrom) || contentBetween(doc, s.GapTo, s.To)) {
		return prosemirror.Node{}, fmt.Errorf("structure gap-replace would overwrite content")
	}

	gap, err := doc.Slice(s.GapFrom, s.GapTo, false)
	if err != nil {
		return prosemirror.Node{}, fmt.Errorf("failed to slice gap: %w", err)
	}

	if gap.OpenStart != 0 || gap.OpenEnd != 0 {
		return prosemirror.Node{}, fmt.Errorf("gap is not a flat range")
	}

	inserted := s.Slice.InsertAt(s.Insert, gap.Content)
	if inserted == nil {
		return prosemirror.Node{}, fmt.Errorf("failed to insert slice")
	}

	return doc.Replace(s.From, s.To, *inserted)
}

func contentBetween(doc prosemirror.Node, from, to int) bool {
	fromNode, _ := doc.Resolve(from)
	dist := to - from
	depth := fromNode.Depth

	for dist > 0 && depth > 0 && fromNode.IndexAfter(depth) == fromNode.Node(depth).ChildCount() {
		depth--
		dist--
	}

	if dist > 0 {
		next := fromNode.Node(depth).MaybeChild(fromNode.IndexAfter(depth))
		for dist > 0 {
			if next == nil || next.IsLeaf() {
				return true
			}

			next = next.FirstChild()
			dist--
		}

	}

	return false
}
