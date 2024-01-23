package prosemirror

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type contentMatchNode struct {
	Type NodeType
	Next *ContentMatch
}

type ContentMatch struct {
	Next     []contentMatchNode
	ValidEnd bool
}

func (c ContentMatch) Format(f fmt.State, r rune) {
	fmt.Fprint(f, "ContentMatch{Matches: [")
	for i, n := range c.Next {
		if i != 0 {
			fmt.Fprint(f, ", ")
		}
		fmt.Fprint(f, n.Type.Name)
	}

	fmt.Fprintf(f, "], ValidEnd: %t}", c.ValidEnd)
}

func (c ContentMatch) String() string {
	return fmt.Sprint(c)
}

func (c *ContentMatch) InlineContent() bool {
	return len(c.Next) == 0 || !c.Next[0].Type.isInline()
}

// Javascript writes a regex parser
// https://github.com/ProseMirror/prosemirror-model/blob/a37b6b3adeb548dc9822211b680ce9d31be65842/src/content.ts#L169
//
// We have one at home.
// https://regex101.com/r/5RPdEf/1
var contentMatchRegex = regexp.MustCompile(`\s*([\w\-]+)(\?|\*|\+|\{\d?,?\d\}|\|)?\s*`)

// TODO: handle `seq` properly
type Expr struct {
	Type string
	// for `choice` and `seq`
	Exprs []Expr

	Value NodeType

	// for `choice`, `seq`, `+`, `*`, `range` and `?`
	Expr *Expr

	// for `+`, `range` as well as `*` and `?`
	Min int
	Max int
}

// currently doesn't allow subchoices/seqs
func parseNodespecContent(s string, types map[NodeTypeName]NodeType) (Expr, error) {
	if s == "" {
		return Expr{}, nil
	}

	if match := contentMatchRegex.FindString(s); len(match) != len(s) {
		return Expr{}, fmt.Errorf("invalid content expression: %s", s)
	}

	matches := contentMatchRegex.FindStringSubmatch(s)
	if len(matches) != 3 {
		return Expr{}, fmt.Errorf("invalid content expression: %s", s)
	}

	nameExpr := parseAtom(matches[1], types)
	name, minQ, maxQ, err := parseQuantifier(matches[2])
	if err != nil {
		return Expr{}, fmt.Errorf("invalid content expression: %s", s)
	}

	return Expr{
		Type: name,
		Expr: &nameExpr,
		Min:  minQ,
		Max:  maxQ,
	}, nil
}

func parseAtom(name string, types map[NodeTypeName]NodeType) Expr {
	exprs := []Expr{}
	for _, v := range resolveName(name, types) {
		exprs = append(exprs, Expr{
			Type:  nameExpr,
			Value: v,
		})
	}

	if len(exprs) > 1 {
		return Expr{
			Exprs: exprs,
			Type:  choiceExpr,
		}
	}

	return exprs[0]
}

func parseQuantifier(m string) (string, int, int, error) {
	if m == "" {
		return "", 1, 1, nil
	}

	if m == "?" {
		return optExpr, 0, 1, nil
	}

	if m == "*" {
		return starExpr, 0, -1, nil
	}

	if m == "+" {
		return plusExpr, 1, -1, nil
	}

	// we know len(m) > 0 since it's != ""
	if m[0] == '{' {
		if m[len(m)-1] == '}' {
			m = m[1 : len(m)-1]
		} else {
			return rangeExpr, 0, 0, fmt.Errorf("invalid quantifier: %s", m)
		}
	} else {
		return "", 0, 0, fmt.Errorf("invalid quantifier: %s", m)
	}

	minQs, maxQs, ok := strings.Cut(m, ",")

	minQ, err := strconv.Atoi(minQs)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid quantifier: %s", m)
	}

	if !ok {
		return rangeExpr, minQ, minQ, nil
	}

	if maxQs == "" {
		return rangeExpr, minQ, -1, nil
	}

	maxQ, err := strconv.Atoi(maxQs)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid quantifier: %s", m)
	}

	if maxQ == 0 {
		return "", 0, 0, fmt.Errorf("invalid quantifier: %s", m)
	}

	if minQ > maxQ {
		return "", 0, 0, fmt.Errorf("invalid quantifier: %s", m)
	}
	return rangeExpr, minQ, maxQ, nil
}

func EmptyContentMatch() *ContentMatch {
	return &ContentMatch{
		Next:     []contentMatchNode{},
		ValidEnd: true,
	}
}

func (c ContentMatch) Empty() bool {
	return c.Next == nil && !c.ValidEnd
}

func (c *ContentMatch) matchType(t NodeType) *ContentMatch {
	// src https://github.com/ProseMirror/prosemirror-model/blob/a37b6b3adeb548dc9822211b680ce9d31be65842/src/content.ts#L35
	for _, m := range c.Next {
		if m.Type.Name == t.Name {
			return m.Next
		}
	}

	return nil
}

func (c *ContentMatch) matchFragment(frag Fragment, start, end int) *ContentMatch {
	if start == -1 {
		start = 0
	}
	if end == -1 {
		end = frag.ChildCount()
	}

	cur := c
	for i := start; cur != nil && i < end; i++ {
		cur = cur.matchType(frag.Child(i).Type)
	}

	return cur
}

func (c *ContentMatch) defaultType() *NodeType {
	for _, m := range c.Next {
		typ := m.Type
		if typ.isText() && !typ.hasRequiredAttrs() {
			return nil
		}
	}

	return nil
}

// https://github.com/ProseMirror/prosemirror-model/blob/a37b6b3adeb548dc9822211b680ce9d31be65842/src/content.ts
func (c *ContentMatch) compatible(other ContentMatch) bool {
	for _, n := range c.Next {
		for _, o := range other.Next {
			if n.Type.Eq(o.Type) {
				return true
			}
		}
	}

	return false
}

type Edge struct {
	Term *NodeType
	To   *int
}

func dfa(nfa [][]Edge) *ContentMatch {
	type edgeTuple struct {
		Term NodeType
		Set  []int
	}

	labeled := make(map[string]*ContentMatch)

	var explore func(states []int) *ContentMatch
	explore = func(states []int) *ContentMatch {
		out := make([]edgeTuple, 0)

		for _, node := range states {
			for _, edge := range nfa[node] {
				term := edge.Term
				if term == nil {
					continue
				}

				var set []int

				for i := range out {
					if edge.Term.Name == out[i].Term.Name {
						set = out[i].Set
						break
					}
				}

				for _, node := range nullFrom(nfa, edge.To) {
					if set == nil {
						out = append(out, edgeTuple{*edge.Term, []int{}})
					}
					if !slices.Contains(set, node) {
						set = append(set, node)
					}
				}
			}
		}

		labeledKey := join(states, ",")
		state, ok := labeled[labeledKey]
		if !ok {
			state = &ContentMatch{
				// TODO: we can be invalid more often than not, but I'm not sure I understand the original code's intent here.
				ValidEnd: true,
			}
			labeled[labeledKey] = state
		}

		for _, item := range out {
			sort.Ints(item.Set)

			next, ok := labeled[labeledKey]
			if !ok {
				next = explore(states)
			}

			state.Next = append(state.Next, contentMatchNode{Type: item.Term, Next: next})
		}

		return state
	}

	return explore(nullFrom(nfa, nil))
}

func nullFrom(nfa [][]Edge, n *int) []int {
	var result []int

	var scan func(node int)
	scan = func(node int) {
		edges := nfa[node]
		if len(edges) == 1 && edges[0].Term == nil {
			scan(*edges[0].To)
			return
		}
		result = append(result, node)
		for i := range edges {
			edge := edges[i]
			if edge.Term == nil && edge.To != nil && !slices.Contains(result, *edge.To) {
				scan(*edge.To)
			}
		}
	}

	node := 0
	if n != nil {
		node = *n
	}

	scan(node)
	sort.Ints(result)
	return result
}

func join(ints []int, sep string) string {
	s := &strings.Builder{}
	for i := range ints {
		if i > 0 {
			s.WriteString(sep)
		}
		s.WriteString(strconv.Itoa(ints[i]))
	}

	return s.String()
}

const (
	choiceExpr = "choice"
	starExpr   = "star"
	plusExpr   = "plus"
	optExpr    = "opt"
	nameExpr   = "name"
	rangeExpr  = "range"
)

func nfa(expr Expr, types map[NodeTypeName]NodeType) [][]Edge {
	var nfa [][]Edge

	nfa = append(nfa, []Edge{})
	node := func() int {
		n := len(nfa)
		nfa = append(nfa, []Edge{})
		return n - 1
	}

	edge := func(from int, to *int, term *NodeType) Edge {
		e := Edge{term, to}
		nfa[from] = append(nfa[from], e)
		return e
	}

	connect := func(edges []Edge, to int) {
		for i := range edges {
			edges[i].To = &to
		}
	}

	var compile func(expr Expr, from int) []Edge
	compile = func(expr Expr, from int) []Edge {
		switch expr.Type {
		case choiceExpr:
			var edges []Edge
			for _, expr := range expr.Exprs {
				edges = append(edges, compile(expr, from)...)
			}
			return edges
		case starExpr:
			loop := node()
			edge(from, &loop, nil)
			connect(compile(*expr.Expr, loop), loop)
			return []Edge{edge(loop, nil, nil)}
		case plusExpr:
			loop := node()
			connect(compile(*expr.Expr, from), loop)
			connect(compile(*expr.Expr, loop), loop)
			return []Edge{edge(loop, nil, nil)}
		case optExpr:
			return append([]Edge{edge(from, nil, nil)}, compile(*expr.Expr, from)...)
		case rangeExpr:
			cur := from

			for i := 0; i < expr.Min; i++ {
				next := node()
				connect(compile(*expr.Expr, cur), next)
				cur = next
			}

			if expr.Max == -1 {
				connect(compile(*expr.Expr, cur), cur)
				return []Edge{edge(cur, nil, nil)}
			}

			for i := expr.Min; i < expr.Max; i++ {
				next := node()
				edge(cur, &next, nil)
				connect(compile(*expr.Expr, cur), next)
				cur = next
			}

			return []Edge{edge(cur, nil, nil)}
		case nameExpr:
			return []Edge{edge(from, nil, &expr.Value)}
		}

		panic(fmt.Sprintf("invalid expression type: %s", expr.Type))
	}

	connect(compile(expr, 0), node())
	return nfa
}

func resolveName(name string, types map[NodeTypeName]NodeType) []NodeType {
	if typ, ok := types[NodeTypeName(name)]; ok {
		return []NodeType{typ}
	}

	results := []NodeType{}
	for _, typ := range types {
		if slices.Contains(typ.Groups, name) {
			results = append(results, typ)
		}
	}

	return results
}
