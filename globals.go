package prosemirror

// TODO: possibly revisite this approach to avoid using a global store.
// not using a global store means that we can have multiple instance of different schemas
// but it also means we lost the convenience of using `json.Unmarshaler`s.
// This is fine because most of the time we have a single schema, but using global state still sucks.

import "sync"

var (
	// nodeTypeStore is a map of node types by name.
	// used for unmarshalling Nodes
	nodeTypeStore = map[NodeTypeName]NodeType{}

	// markTypeStore is a map of mark types by name.
	// used for unmarshalling Marks
	markTypeStore = map[MarkTypeName]MarkType{}

	registerMu = sync.Mutex{}
)

// registerNodeType registers a node type in the global store.
// It is used for unmarshalling Nodes.
// It must have been registered beforehand.
// The typ must be fully initialized, i.e it must come from a valid schema
func registerNodeType(name NodeTypeName, typ NodeType) {
	registerMu.Lock()
	defer registerMu.Unlock()

	nodeTypeStore[name] = typ
}

// registerMarkType registers a mark type in the global store.
// It is used for unmarshalling Marks.
// It must have been registered beforehand.
// The typ must be fully initialized, i.e it must come from a valid schema
func registerMarkType(name MarkTypeName, spec MarkType) {
	registerMu.Lock()
	defer registerMu.Unlock()

	markTypeStore[name] = spec
}

// RegisterSchema registers a schema in the global store.
// It is used for unmarshalling Nodes and Marks.
func RegisterSchema(s Schema) {
	for name, typ := range s.Nodes {
		registerNodeType(name, typ)
	}

	for name, spec := range s.Marks {
		registerMarkType(name, spec)
	}
}
