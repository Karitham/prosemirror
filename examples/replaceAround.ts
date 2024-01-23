import { Node, Schema } from "prosemirror-model";
import { schema } from "prosemirror-schema-basic";
import { ReplaceAroundStep } from "prosemirror-transform";
import { addListNodes } from "prosemirror-schema-list";

const s = new Schema({
  nodes: addListNodes(schema.spec.nodes, "paragraph block*", "block"),
  marks: schema.spec.marks,
});

const doc = Node.fromJSON(s, {
  type: "doc",
  content: [
    { type: "paragraph", content: [{ type: "text", text: "Hello " }] },
    {
      type: "paragraph",
      content: [{ type: "text", text: "Man this is epic." }],
    },
  ],
});

const step = ReplaceAroundStep.fromJSON(s, {
  stepType: "replaceAround",
  from: 8,
  to: 27,
  gapFrom: 9,
  gapTo: 26,
  insert: 1,
  slice: { content: [{ type: "code_block" }] },
  structure: true,
});

console.log(JSON.stringify(step.apply(doc).doc));
