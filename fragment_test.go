package prosemirror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFragment_findIndex(t *testing.T) {
	s := Must(NewSchema(SchemaSpec{
		Nodes: map[NodeTypeName]NodeSpec{
			"doc": {
				Content: "block+",
			},
			"paragraph": {
				Content: "inline*",
				Group:   "block",
			},
			"text": {
				Group: "inline",
			},
		},
		TopNode: "doc",
	}))

	newP := func(text string) Node {
		return s.Node("paragraph", nil, NewFragment(s.Text(text)))
	}

	tests := []struct {
		name string
		f    Fragment

		pos int

		index  int
		offset int
	}{
		{
			name:   "empty fragment",
			f:      NewFragment(),
			pos:    0,
			index:  0,
			offset: 0,
		},
		{
			name:   "text node",
			f:      NewFragment(s.Text("Hello World!")),
			pos:    5,
			index:  0,
			offset: 0,
		},
		{
			name: "basic doc",
			f: NewFragment(
				newP("Crazy?"),
				newP("I was crazy once."),
				newP("They put me in a room."),
			),
			pos:    14,
			index:  1,
			offset: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, offset := tt.f.findIndex(tt.pos)
			assert.Equal(t, tt.index, index)
			assert.Equal(t, tt.offset, offset)
		})
	}
}
