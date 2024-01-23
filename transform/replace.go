package transform

import (
	"fmt"

	"github.com/go-json-experiment/json"

	"github.com/karitham/prosemirror"
)

func init() {
	RegisterTransformer("replace", func() Applier {
		return new(ReplaceStep)
	})
}

type ReplaceStep struct {
	BaseStep
	Slice     prosemirror.Slice `json:"slice"`
	Structure bool              `json:"structure,omitempty"`
}

func (s *ReplaceStep) UnmarshalJSON(data []byte) error {
	type a ReplaceStep
	aux := a{}

	if err := json.Unmarshal(data, &aux, json.RejectUnknownMembers(true)); err != nil {
		return fmt.Errorf("failed to decode replace step (%s): %w", string(data), err)
	}

	*s = ReplaceStep(aux)
	return nil
}

func (s *ReplaceStep) MarshalJSON() ([]byte, error) {
	type a ReplaceStep
	aux := a(*s)

	return json.Marshal(aux)
}

func (s *ReplaceStep) Apply(doc prosemirror.Node) (prosemirror.Node, error) {
	doc, err := doc.Replace(s.From, s.To, s.Slice)
	if err != nil {
		return prosemirror.Node{}, fmt.Errorf("failed to replace: %w", err)
	}

	return doc, nil
}
