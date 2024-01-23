package prosemirror

import (
	"fmt"
)

func replace(from, to ResolvedPos, slice Slice) (Node, error) {
	if slice.OpenStart > from.Depth {
		return Node{}, fmt.Errorf("inserted content deeper than insertion position (%d)", from.Depth)
	}

	if from.Depth-slice.OpenStart != to.Depth-slice.OpenEnd {
		return Node{}, fmt.Errorf("inconsistent open depths (%d and %d)", from.Depth-slice.OpenStart, to.Depth-slice.OpenEnd)
	}

	return replaceOuter(from, to, slice, 0)
}

func replaceOuter(from, to ResolvedPos, slice Slice, depth int) (Node, error) {
	index := from.Index(depth)
	node := from.Node(depth)

	switch {
	case index == to.Index(depth) && depth < from.Depth-slice.OpenStart:
		inner, err := replaceOuter(from, to, slice, depth+1)
		if err != nil {
			return Node{}, err
		}

		return node.copy(node.Content.replaceChild(index, inner)), nil

	case slice.Content.Size == 0:
		r, err := replaceTwoWay(from, to, depth)
		if err != nil {
			return Node{}, err
		}

		closed, err := node.close(r)
		if err != nil {
			return Node{}, err
		}

		return closed, nil

	case slice.OpenStart == 0 && slice.OpenEnd == 0 && from.Depth == depth && to.Depth == depth:
		parent := from.Parent()
		content := parent.Content

		cut1 := content.cut(0, from.ParentOffset)
		append1 := cut1.append(slice.Content)
		cut2 := content.cut(to.ParentOffset, -1)
		append2 := append1.append(cut2)

		closed, err := parent.close(append2)
		if err != nil {
			return Node{}, err
		}

		return closed, nil
	}

	start, end := prepareSliceForReplace(slice, from)

	r, err := replaceThreeWay(from, start, end, to, depth)
	if err != nil {
		return Node{}, err
	}

	closed, err := node.close(r)
	if err != nil {
		return Node{}, err
	}
	return closed, nil
}

func replaceThreeWay(from, start, end, to ResolvedPos, depth int) (Fragment, error) {
	var openStart, openEnd *Node = nil, nil
	var err error
	if from.Depth > depth {
		openStart, err = joinable(from, start, depth+1)
		if err != nil {
			return Fragment{}, err
		}
	}
	if to.Depth > depth {
		openEnd, err = joinable(end, to, depth+1)
		if err != nil {
			return Fragment{}, err
		}
	}

	content := appendRange([]Node{}, nil, &from, depth)
	if openStart != nil && openEnd != nil && start.Index(depth) == end.Index(depth) {
		err := checkJoin(*openStart, *openEnd)
		if err != nil {
			return Fragment{}, err
		}

		r, err := replaceThreeWay(from, start, end, to, depth+1)
		if err != nil {
			return Fragment{}, err
		}

		node, err := openStart.close(r)
		if err != nil {
			return Fragment{}, err
		}

		content = appendNode(content, node)
		content = appendRange(content, &to, nil, depth)
		return NewFragment(content...), nil
	}

	if openStart != nil {
		r, err := replaceTwoWay(from, start, depth+1)
		if err != nil {
			return Fragment{}, err
		}

		node, err := openStart.close(r)
		if err != nil {
			return Fragment{}, err
		}
		content = appendNode(content, node)
	}

	content = appendRange(content, &start, &end, depth)

	if openEnd != nil {
		r, err := replaceTwoWay(end, to, depth+1)
		if err != nil {
			return Fragment{}, err
		}

		node, err := openEnd.close(r)
		if err != nil {
			return Fragment{}, err
		}
		content = appendNode(content, node)
	}

	content = appendRange(content, &to, nil, depth)
	return NewFragment(content...), nil
}

func replaceTwoWay(from, to ResolvedPos, depth int) (Fragment, error) {
	content := appendRange([]Node{}, nil, &from, depth)

	if from.Depth > depth {
		nodeType, err := joinable(from, to, depth+1)
		if nodeType == nil {
			return Fragment{}, err
		}

		r, err := replaceTwoWay(from, to, depth+1)
		if err != nil {
			return Fragment{}, err
		}

		node, err := nodeType.close(r)
		if err != nil {
			return Fragment{}, err
		}
		content = appendNode(content, node)
	}

	content = appendRange(content, &to, nil, depth)
	return NewFragment(content...), nil
}

// prepareSliceForReplace prepares a slice for replacement by taking the open
// depth into account.
//
// returns {start, end}
func prepareSliceForReplace(slice Slice, along ResolvedPos) (ResolvedPos, ResolvedPos) {
	extra := along.Depth - slice.OpenStart
	parent := along.Node(extra)
	node := parent.copy(slice.Content)
	for i := extra - 1; i >= 0; i-- {
		node = along.Node(i).copy(NewFragment(node))
	}

	start, err := resolve(node, slice.OpenStart+extra)
	if err != nil {
		panic(err)
	}

	end, err := resolve(node, node.Content.Size-slice.OpenEnd-extra)
	if err != nil {
		panic(err)
	}

	return start, end
}

func appendNode(content []Node, child Node) []Node {
	last := len(content) - 1
	if last >= 0 && child.IsText() && child.sameMarkup(content[last]) {
		content[last] = child.withText(content[last].Text + child.Text)
		return content
	}

	return append(content, child)
}

func appendRange(content []Node, start *ResolvedPos, end *ResolvedPos, depth int) []Node {
	var node Node
	switch {
	case start != nil:
		node = start.Node(depth)
	case end != nil:
		node = end.Node(depth)
	}

	endIndex := node.ChildCount()
	if end != nil {
		endIndex = end.Index(depth)
	}

	startIndex := 0
	if start != nil {
		startIndex = start.Index(depth)
		if start.Depth > depth {
			startIndex++
		} else if start.TextOffset() {
			content = appendNode(content, *start.NodeAfter())
			startIndex++
		}
	}

	for i := startIndex; i < endIndex; i++ {
		content = appendNode(content, *node.Child(i))
	}

	if end != nil && end.Depth == depth && end.TextOffset() {
		content = appendNode(content, *end.NodeBefore())
	}

	return content
}
