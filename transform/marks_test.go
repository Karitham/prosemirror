package transform_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/karitham/prosemirror"
	"github.com/karitham/prosemirror/transform"
	"github.com/stretchr/testify/assert"

	// for side effects
	_ "github.com/karitham/prosemirror/schema"
)

func TestMark(t *testing.T) {
	type tt struct {
		name string
		doc  prosemirror.Node
		step transform.Step
		want prosemirror.Node
	}

	tests := []tt{
		{
			name: "add mark basic",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World!"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello "},{"type":"text","marks":[{"type":"em"}],"text":"World"},{"type":"text","text":"!"}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"addMark","mark":{"type":"em"},"from":7,"to":12}`),
		},
		{
			name: "remove mark",
			step: fromJSON[transform.Step](`{"stepType":"removeMark","mark":{"type":"em"},"from":7,"to":12}`),
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello "},{"type":"text","marks":[{"type":"em"}],"text":"World"},{"type":"text","text":"!"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World!"}]}]}`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldDoc := spew.Sdump(tc.doc)
			got, err := tc.step.Apply(tc.doc)
			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, spew.Sdump(tc.doc), oldDoc, "doc should not be mutated") {
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, tc.want, got) {
				return
			}
		})
	}
}
