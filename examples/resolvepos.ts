import { Node, Schema } from "prosemirror-model";

const textSchema = new Schema({
  nodes: {
    text: {
      group: "inline",
    },
    doc: {
      content: "block+",
    },
    paragraph: {
      content: "inline*",
      group: "block",
    },
  },
  marks: {
    em: {
      parseDOM: [{ tag: "i" }, { tag: "em" }, { style: "font-style=bold" }],
      toDOM() {
        return ["em", 0];
      },
    },
  },
});

const text = [
  "rats",
  "rats",
  "we are the rats",
  "we prey at night",
  "we stalk at night",
  "we're the rats",
  "i'm the giant rat that makes all of the rules",
  "let's see what kind of trouble we can get ourselves into",
  "we're the rats",
  "we're the rats",
].map((text) => ({
  type: "paragraph",
  content: [{ type: "text", text }],
}));

// create a basic doc
const doc = Node.fromJSON(textSchema, {
  type: "doc",
  content: text,
});

let resolved = doc.resolve(64);
console.log({
  depth: resolved.depth,
  pos: resolved.pos,
  parentOffset: resolved.parentOffset,
  name: resolved.parent.type.name,
  indexPath: [resolved.index(0).toString(), resolved.index(1).toString()],
});

const doc2 = Node.fromJSON(textSchema, {
  type: "doc",
  content: [{ type: "paragraph", content: [{ type: "text", text: "Crazy" }] }],
});

const r = doc2.resolve(6);

console.log({ start: r.pos, end_offset: r.parentOffset, depth: r.depth });

const doc3 = Node.fromJSON(textSchema, {
  type: "doc",
  content: [
    { type: "paragraph", content: [{ type: "text", text: "Hello " }] },
    {
      type: "paragraph",
      content: [{ type: "text", text: "Man this is epic." }],
    },
    {
      type: "paragraph",
      content: [{ type: "text", text: "Does this still work." }],
    },
  ],
});

const doc3From = doc3.resolve(9);

console.log({
  depth: doc3From.depth,
  end: doc3From.end(1),
  sharedDepth: doc3From.sharedDepth(26),
  start: doc3From.start(doc3From.sharedDepth(26)),
  slice: doc3.slice(9, 26).content.toString(),
});
