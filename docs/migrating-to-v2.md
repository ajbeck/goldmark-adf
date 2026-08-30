# Migrating to goldmark-adf v2

goldmark-adf v2 moves to Goldmark v2 and Go 1.27. It is a new Go module
major version; v1 remains available for consumers that cannot migrate yet.

## Update the module path

Change imports from:

```go
"github.com/ajbeck/goldmark-adf"
```

to:

```go
"github.com/ajbeck/goldmark-adf/v2"
```

Then update the dependency:

```bash
go get github.com/ajbeck/goldmark-adf/v2@v2.0.0
go mod tidy
```

## Upgrade to Go 1.27

v2 requires Go 1.27 or later. `encoding/json/v2` is standard in Go 1.27, so
remove `GOEXPERIMENT=jsonv2` from local scripts and CI configuration.

## Update converter construction

`New` and `NewWithGFM` still provide the convenient reusable converter API.
They return goldmark-adf's `*Markdown` facade rather than Goldmark's removed
top-level `goldmark.Markdown` type:

```go
md := adf.NewWithGFM()
if err := md.Convert(markdown, output); err != nil {
	return err
}
```

Use `New` for CommonMark-only conversion. Use `NewWithGFM` when the input
needs GFM or goldmark-adf's round-trip syntax.

## Compose Goldmark directly when needed

Goldmark v2 separates parsers and renderers. v2 exposes the native components
for integrations that need to inspect or augment the AST:

```go
p := adf.NewWithGFMParser()
r := adf.NewRenderer()

doc := p.Parse(markdown)
if err := r.Render(output, markdown, doc); err != nil {
	return err
}
```

For a custom parser profile, combine `extension.GFMParser` and
`adf.RoundTripParser` with Goldmark v2's `parser.WithExtensions`. A parser is
for one parse at a time; use the `*Markdown` facade when conversions must run
concurrently. The custom nodes are a supported API in
`github.com/ajbeck/goldmark-adf/v2/astext`.

## Update configuration

`WithTableLayout` and `WithImageLayout` use exported layout types and
constants in v2. Use the named constants instead of arbitrary strings:

```go
adf.WithTableLayout(adf.TableLayoutWide)
adf.WithImageLayout(adf.ImageLayoutFullWidth)
```

Invalid values created by an explicit type conversion are reported by
conversion rather than producing schema-invalid ADF.

`WithTableLayout` now applies to generated tables; it had no effect in v1.

## Update custom image handling

v1 advertised `WithImageHandler` but did not invoke it. v2 makes it a working
override with this shape:

```go
type ImageHandler func(Image) (ImageResult, error)
```

`ImageResult` declares whether the returned ADF node is `ImageInline` or
`ImageBlock`. A block result uses paragraph splitting where required. A
configured handler takes precedence over `WithExternalMedia`.

Reusable v2 converters are safe for concurrent conversion. An image handler is
application code and must itself be safe for concurrent calls. The returned
node becomes renderer-owned after the handler returns.

## Review output changes

v2 keeps two-space-indented JSON, but consumers should compare decoded ADF
semantics rather than JSON bytes.

Notable correctness changes include:

- A `[card:URL]` token that is the only unmarked content of a paragraph emits
  `blockCard`; elsewhere it emits `inlineCard`.
- `[embed:URL]` emits a schema-valid `embedCard` with `layout: "center"` only
  when it is the sole unmarked content of a paragraph. In mixed inline text it
  remains literal text because ADF does not permit inline embed cards.
- Task and decision lists are converted only when the Markdown structure can
  be represented without loss in ADF. Ordered task lists remain ordered lists
  with textual checkbox markers.
- Task and decision `localId` values remain empty and deterministic because
  Markdown provides no persistent identity.
- Malformed custom round-trip markers remain ordinary Markdown/text instead of
  producing partial ADF nodes or conversion errors.

## Known compatibility boundaries

The shared GitHub-alert syntax maps `[!NOTE]` to `panelType: "info"`.
Consequently, ADF `note` and `custom` panels normalize to `info` through that
syntax. v2 keeps this established interchange format; preserving those exact
panel types requires a coordinated future change to both round-trip libraries.

## Validate when required

The renderer does not run JSON Schema validation for every conversion. Use the
opt-in `adfschema` package at application boundaries where validation is
required:

```go
if err := adfschema.Validate(output.Bytes()); err != nil {
	return err
}
```
