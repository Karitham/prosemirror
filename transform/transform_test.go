package transform_test

import (
	"testing"

	"github.com/go-json-experiment/json"

	"github.com/davecgh/go-spew/spew"
	"github.com/karitham/prosemirror"
	_ "github.com/karitham/prosemirror/schema"
	"github.com/karitham/prosemirror/transform"
	"github.com/stretchr/testify/assert"
)

func TestTransforms(t *testing.T) {
	type test struct {
		name string
		step string
		want string
		doc  string
	}

	tests := []test{
		{
			name: "transform paragraph to code_block",
			doc:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello "}]},{"type":"paragraph","content":[{"type":"text","text":"Man this is epic."}]},{"type":"paragraph","content":[{"type":"text","text":"Does this still work."}]}]}`,
			want: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello "}]},{"type":"code_block","content":[{"type":"text","text":"Man this is epic."}]},{"type":"paragraph","content":[{"type":"text","text":"Does this still work."}]}]}`,
			step: `{"stepType":"replaceAround","from":8,"to":27,"gapFrom":9,"gapTo":26,"insert":1,"slice":{"content":[{"type":"code_block"}]},"structure":true}`,
		},
		{
			name: "append to paragraph",
			doc:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World!"}]}]}`,
			want: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World!?"}]}]}`,
			step: `{"stepType":"replace","from":13,"to":13,"slice":{"content":[{"type":"text","text":"?"}]}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := fromJSON[prosemirror.Node](tc.doc)
			want := fromJSON[prosemirror.Node](tc.want)
			step := fromJSON[transform.Step](tc.step)

			oldDoc := spew.Sdump(tc.doc)
			got, err := step.Apply(doc)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, spew.Sdump(tc.doc), oldDoc, "doc should not be mutated") {
				return
			}

			assert.Equal(t, spew.Sdump(want), spew.Sdump(got))
		})
	}
}

func fromJSON[T any](s string) T {
	var v T
	_ = json.Unmarshal([]byte(s), &v)
	return v
}
