package prosemirror

import "fmt"

type Slice struct {
	Content   Fragment `json:"content,omitempty"`
	OpenStart int      `json:"openStart,omitempty"`
	OpenEnd   int      `json:"openEnd,omitempty"`
}

func (s Slice) String() string {
	return fmt.Sprintf("Slice{Content: %v, OpenStart: %d, OpenEnd: %d}", s.Content, s.OpenStart, s.OpenEnd)
}

// insertAt(pos: number, fragment: Fragment) {
//     let content = insertInto(this.content, pos + this.openStart, fragment)
//     return content && new Slice(content, this.openStart, this.openEnd)
//   }

func (s Slice) InsertAt(pos int, frag Fragment) *Slice {
	content := insertInto(s.Content, pos+s.OpenStart, frag, nil)
	if content == nil {
		return nil
	}

	return &Slice{
		Content:   *content,
		OpenStart: s.OpenStart,
		OpenEnd:   s.OpenEnd,
	}
}

func insertInto(content Fragment, dist int, insert Fragment, parent *Node) *Fragment {
	index, offset := content.findIndex(dist)
	child := content.maybeChild(index)
	if offset == dist || (child != nil && child.IsText()) {
		if parent != nil && !parent.canReplace(index, index, insert, -1, -1) {
			return nil
		}

		c := content.cut(0, dist)
		c = c.append(insert)
		subc := content.cut(dist, -1)
		c = c.append(subc)
		return &c
	}

	inner := insertInto(child.Content, dist-offset-1, insert, nil)
	if inner != nil {
		r := content.replaceChild(index, child.copy(*inner))
		return &r
	}

	return nil
}
