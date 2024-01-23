package prosemirror_test

import (
	"reflect"
	"testing"

	"github.com/go-json-experiment/json"

	"github.com/karitham/prosemirror"
	b "github.com/karitham/prosemirror/builder"
	"github.com/karitham/prosemirror/schema"
)

func TestNodeCalcSize(t *testing.T) {
	b := b.New(prosemirror.Must(prosemirror.NewSchema(schema.DefaultSpec)))
	n := b.Doc(
		b.PText("Crazy?"),
		b.PText("I was crazy once."),
		b.PText("They put me in a room."),
		b.PText("A rubber room."),
		b.PText("A rubber room with rats."),
		b.PText("Rubber rats."),
		b.PText("I hate rats."),
	)

	if n.Content.Size != 121 {
		t.Errorf("NodeSize() = %v, want %v", n.Content.Size, 121)
	}
}

func TestNodeIndexing(t *testing.T) {
	b := b.New(prosemirror.Must(prosemirror.NewSchema(schema.DefaultSpec)))
	type test struct {
		name string
		got  prosemirror.Node
		want int
	}

	tests := []test{{
		name: "default doc",
		want: 121,
		got: b.Doc(
			b.PText("Crazy?"),
			b.PText("I was crazy once."),
			b.PText("They put me in a room."),
			b.PText("A rubber room."),
			b.PText("A rubber room with rats."),
			b.PText("Rubber rats."),
			b.PText("I hate rats."),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got.Content.Size, tt.want) {
				t.Errorf("DocSize() = %v, want %v", tt.got.Content.Size, tt.want)
			}
		})
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func fromJSON[T any](s string) T {
	var t T
	err := json.Unmarshal([]byte(s), &t)
	if err != nil {
		panic(err)
	}

	return t
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
