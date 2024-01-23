package transform

import (
	"fmt"

	"github.com/go-json-experiment/json"

	"github.com/karitham/prosemirror"
)

// Applier is an interface that all step types implement.
//
// To register a new step type, you need to call RegisterTransformer with
// the step type and a function that returns a new instance of
// the step type.
type Applier interface {
	Apply(prosemirror.Node) (prosemirror.Node, error)
	json.UnmarshalerV1
	json.MarshalerV1
}

// Step is an abstract type that gets implemented by the various step types
// that we have. This is so that we can unmarshal the step without knowing
// what type it is, and then we can use the type to apply the step.
type Step struct {
	Impl Applier
}

var _ Applier = (*Step)(nil)

func (s *Step) UnmarshalJSON(data []byte) error {
	aux := struct {
		Type string `json:"stepType"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to decode step (%s): %w", string(data), err)
	}

	// now we can check what we registered
	f, ok := transformers[aux.Type]
	if !ok {
		return fmt.Errorf("unknown step type: %s", aux.Type)
	}

	impl := f()
	err := impl.UnmarshalJSON(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal step (%s): %w", string(data), err)
	}

	s.Impl = impl
	return nil
}

func (s *Step) MarshalJSON() ([]byte, error) {
	return s.Impl.MarshalJSON()
}

func (s *Step) Apply(n prosemirror.Node) (prosemirror.Node, error) {
	return s.Impl.Apply(n)
}

type BaseStep struct {
	Type string `json:"stepType"`
	From int    `json:"from"`
	To   int    `json:"to"`
}

// NewStep returns a new step from the given applier.
func NewStep(a Applier) Step {
	return Step{Impl: a}
}

// RegisterTransformer registers a new step type.
func RegisterTransformer(name string, f func() Applier) {
	if _, ok := transformers[name]; ok {
		panic(fmt.Sprintf("transformer %s already registered", name))
	}

	transformers[name] = f
}

// transformers is a map of step type names to functions that return a new
// instance of that step type.
var transformers = map[string]func() Applier{}
