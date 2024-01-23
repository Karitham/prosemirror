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

func TestTextNodeReplace(t *testing.T) {
	type tt struct {
		name string
		doc  prosemirror.Node
		step transform.Step
		want prosemirror.Node
	}

	tests := []tt{
		{
			name: "replace full text",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"rats"}]},{"type":"paragraph","content":[{"type":"text","text":"rats"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"farts"}]},{"type":"paragraph","content":[{"type":"text","text":"rats"}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":1,"to":5, "slice":{"content":[{"type":"text","text":"farts"}]}}`),
		},
		{
			name: "insert text",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":6,"to":6,"slice":{"content":[{"type":"text","text":"?"}]}}`),
		},
		{
			name: "delete text",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]},{"type":"paragraph","content":[{"type":"text","text":"I was crazy once."}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy"}]},{"type":"paragraph","content":[{"type":"text","text":"I was crazy once."}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":6,"to":7}`),
		},
		{
			name: "insert paragraph",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]},{"type":"paragraph","content":[{"type":"text","text":"I hate rats."}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]},{"type":"paragraph"},{"type":"paragraph","content":[{"type":"text","text":"I hate rats."}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":7,"to":7,"slice":{"content":[{"type":"paragraph"},{"type":"paragraph"}],"openStart":1,"openEnd":1},"structure":true}`),
		},
		{
			name: "split paragraph",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Cra"}]},{"type":"paragraph","content":[{"type":"text","text":"zy?"}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":4,"to":4,"slice":{"content":[{"type":"paragraph"},{"type":"paragraph"}],"openStart":1,"openEnd":1},"structure":true}`),
		},
		{
			name: "split paragraph",
			doc:  fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy?"}]}]}`),
			want: fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Cra"}]},{"type":"paragraph","content":[{"type":"text","text":"zy?"}]}]}`),
			step: fromJSON[transform.Step](`{"stepType":"replace","from":4,"to":4,"slice":{"content":[{"type":"paragraph"},{"type":"paragraph"}],"openStart":1,"openEnd":1},"structure":true}`),
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

			if !assert.Equal(t, spew.Sdump(tc.want), spew.Sdump(got)) {
				return
			}
		})
	}
}
