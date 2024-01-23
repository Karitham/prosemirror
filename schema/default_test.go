package schema_test

import (
	"testing"

	p "github.com/karitham/prosemirror"
	"github.com/karitham/prosemirror/schema"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSchema(t *testing.T) {
	s := p.Must(p.NewSchema(schema.DefaultSpec))

	// test that text nodes can be created and are valid inside the top node
	if err := s.Nodes["doc"].CheckContent(p.NewFragment(s.Node("paragraph", nil, p.NewFragment(s.Text("hello"))))); err != nil {
		assert.NoError(t, err, "error creating text node")
	}
}
