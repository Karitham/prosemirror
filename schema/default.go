package schema

import (
	p "github.com/karitham/prosemirror"
)

var (
	DefaultSpec = p.SchemaSpec{
		Nodes:   DefaultNodes,
		TopNode: "doc",
		Marks:   DefaultMarks,
	}

	DefaultMarks = map[p.MarkTypeName]p.MarkSpec{
		"link":   {},
		"em":     {},
		"strong": {},
		"code":   {},
	}

	DefaultNodes = map[p.NodeTypeName]p.NodeSpec{
		"doc": {
			Content: "block+",
		},
		"paragraph": {
			Content: "inline*",
			Group:   "block",
		},
		"blockquote": {
			Content: "block+",
			Group:   "block",
		},
		"horizontal_rule": {
			Group: "block",
		},
		"heading": {
			Content: "inline*",
			Group:   "block",
			Attrs: map[string]p.Attribute{
				"level": {
					Default: 1,
				},
			},
		},
		"code_block": {
			Content: "text*",
			Group:   "block",
			Marks:   opt(""),
			Attrs: map[string]p.Attribute{
				"language": {
					Default: nil,
				},
			},
		},
		"text": {
			Group: "inline",
		},
		"image": {
			Inline: true,
			Attrs: map[string]p.Attribute{
				"src":   {},
				"alt":   {},
				"title": {},
			},
			Group: "inline",
		},
		"hard_break": {
			Inline: true,
			Group:  "inline",
		},
	}
)

func init() {
	_, err := p.NewSchema(DefaultSpec)
	if err != nil {
		panic(err)
	}

	DefaultSpec.DontRegister = true
}

func opt[T any](optVal T) *T {
	return &optVal
}
