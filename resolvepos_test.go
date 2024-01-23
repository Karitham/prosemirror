package prosemirror_test

import (
	"testing"

	"github.com/karitham/prosemirror"
	b "github.com/karitham/prosemirror/builder"
	"github.com/karitham/prosemirror/schema"
	"github.com/stretchr/testify/assert"
)

func TestNode_Resolve(t *testing.T) {
	b := b.New(prosemirror.Must(prosemirror.NewSchema(schema.DefaultSpec)))
	type test struct {
		name string
		doc  prosemirror.Node
		pos  int
		want prosemirror.ResolvedPos
	}

	bigTestDoc := b.Doc(
		b.PText("rats"),
		b.PText("rats"),
		b.PText("we are the rats"),
		b.PText("we prey at night"),
		b.PText("we stalk at night"),
		b.PText("we're the rats"),
		b.PText("i'm the giant rat that makes all of the rules"),
		b.PText("let's see what kind of trouble we can get ourselves into"),
		b.PText("we're the rats"),
		b.PText("we're the rats"),
	)

	tests := []test{
		{
			name: "resolve start",
			doc:  b.Doc(b.P(b.Em("cd"), b.Text("ef"))),
			pos:  1,
			want: prosemirror.ResolvedPos{
				NodePath: []prosemirror.Node{
					b.Doc(b.P(b.Em("cd"), b.Text("ef"))),
					b.P(b.Em("cd"), b.Text("ef")),
				},
				OffsetPath:   []int{0, 1},
				IndexPath:    []int{0, 0},
				Pos:          1,
				ParentOffset: 0,
			},
		},
		{
			name: "resolve end",
			doc:  bigTestDoc,
			pos:  bigTestDoc.Content.Size,
			want: prosemirror.ResolvedPos{
				Pos:          bigTestDoc.Content.Size,
				ParentOffset: bigTestDoc.Content.Size,
				IndexPath:    []int{10},
				OffsetPath:   []int{bigTestDoc.Content.Size},
			},
		},
		{
			name: "resolve depth 3",
			doc:  b.Doc(b.P(b.Em("cd"), b.Text("ef"))),
			pos:  6,
			want: prosemirror.ResolvedPos{
				OffsetPath:   []int{6},
				IndexPath:    []int{1},
				Pos:          6,
				ParentOffset: 6,
			},
		},
		{
			name: "resolve 64 big doc",
			doc:  bigTestDoc,
			pos:  64,
			want: prosemirror.ResolvedPos{
				Pos:          64,
				ParentOffset: 16,
				IndexPath:    []int{4, 0},
				OffsetPath:   []int{47, 48},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.doc.Resolve(tt.pos)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.IndexPath, got.IndexPath)
			assert.Equal(t, tt.want.OffsetPath, got.OffsetPath)
			assert.Equal(t, tt.want.ParentOffset, got.ParentOffset)
			assert.Equal(t, tt.want.Pos, got.Pos)
		})
	}
}

func TestNodeResolveExample(t *testing.T) {
	doc := fromJSON[prosemirror.Node](`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Crazy"}]}]}`)
	r, err := doc.Resolve(6)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 6, r.Pos)
	assert.Equal(t, 5, r.ParentOffset)
}

func TestResolvedPos(t *testing.T) {
	doc := fromJSON[prosemirror.Node](`{
		"type": "doc",
		"content": [
		  { "type": "paragraph", "content": [{ "type": "text", "text": "Hello " }]},
		  {
			"type": "paragraph",
			"content": [{ "type": "text", "text": "Man this is epic." }]
		  },
		  {
			"type": "paragraph",
			"content": [{ "type": "text", "text": "Does this still work." }]
		  }
		]
	  }`)
	from, err := doc.Resolve(9)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 1, from.Depth, "depth")
	assert.Equal(t, 26, from.End(1), "end")
	assert.Equal(t, 1, from.SharedDepth(26), "shared depth")
	assert.Equal(t, 9, from.Start(from.SharedDepth(26)), "start")
}
