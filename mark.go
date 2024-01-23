package prosemirror

import "maps"

type Mark struct {
	Type  MarkType       `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

func (m Mark) Eq(other Mark) bool {
	return m.Type.Eq(other.Type) && maps.Equal(m.Attrs, other.Attrs)
}
