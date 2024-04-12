# Prosemirror

This is a go implementation of the prosemirror document model.

It tries to map closely to the original source code.

The `examples` directory is a set of examples providing ways to get easy test data out of the original prosemirror.

We want to be 100% compliant at least on document and transformation aspects.

## Usage

```sh
go get -u github.com/karitham/prosemirror
```

Then create a `Schema` and use it to build your document.

Due to how the original prosemirror implementation handles transforms and how go deals with unmarshalling, you need to globally register your marshallers for schemas. This means this library cannot handle using schemas with the same objects but different unmarshalling policies.

This might be possible since we use `github.com/go-json-experiment/json` to handle the marshalling.
