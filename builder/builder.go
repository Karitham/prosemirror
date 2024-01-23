package builder

import p "github.com/karitham/prosemirror"

type Builder struct {
	schema p.Schema
}

func New(schema p.Schema) *Builder {
	return &Builder{schema: schema}
}

func (b *Builder) Doc(content ...p.Node) p.Node {
	return b.schema.Node("doc", nil, p.NewFragment(content...))
}

func (b *Builder) P(content ...p.Node) p.Node {
	return b.schema.Node("paragraph", nil, p.NewFragment(content...))
}

func (b *Builder) PText(text string) p.Node {
	return b.schema.Node("paragraph", nil, p.NewFragment(b.Text(text)))
}

func (b *Builder) Text(text string) p.Node {
	return b.schema.Text(text)
}

func (b *Builder) Em(content string) p.Node {
	return b.schema.Text(
		content,
		b.schema.Mark("em", nil),
	)
}
