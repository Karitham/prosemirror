import { Fragment, Node, Slice, Schema } from "prosemirror-model";
import { schema } from "prosemirror-schema-basic";
import { addListNodes } from "prosemirror-schema-list";

const s = new Schema({
  nodes: addListNodes(schema.spec.nodes, "paragraph block*", "block"),
  marks: schema.spec.marks,
});

// create a basic doc

const replacedText = Node.fromJSON(s, {
  type: "doc",
  content: ["rats", "rats"].map((text) => ({
    type: "paragraph",
    content: [{ type: "text", text }],
  })),
}).replace(1, 5, new Slice(Fragment.fromArray([s.text("farts")]), 0, 0));

console.log(JSON.stringify(replacedText));

const deletedText = Node.fromJSON(s, {
  type: "doc",
  content: [
    { type: "paragraph", content: [{ type: "text", text: "Crazy?" }] },
    {
      type: "paragraph",
      content: [{ type: "text", text: "I was crazy once." }],
    },
  ],
}).replace(6, 7, Slice.empty);

console.log(JSON.stringify(deletedText));

const splitText = Node.fromJSON(s, {
  type: "doc",
  content: [{ type: "paragraph", content: [{ type: "text", text: "Crazy?" }] }],
}).replace(
  4,
  4,
  Slice.fromJSON(s, {
    content: [{ type: "paragraph" }, { type: "paragraph" }],
    openStart: 1,
    openEnd: 1,
  })
);

console.log(JSON.stringify(splitText));
