# goldmark-adf v2 overview

goldmark-adf converts CommonMark and GitHub Flavored Markdown (GFM) into
[Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
JSON. ADF is the document format used by Jira Cloud and Confluence Cloud.

## Requirements

goldmark-adf v2 requires Go 1.27 or later and depends on
[Goldmark v2](https://github.com/yuin/goldmark/releases/tag/v2.0.0). The
experimental `GOEXPERIMENT=jsonv2` setting is no longer required.

## Architecture

```
Markdown source -> Goldmark v2 parser -> Goldmark AST -> ADF renderer -> ADF JSON
```

The package exposes two levels of API:

- `New` creates a reusable CommonMark-to-ADF converter.
- `NewWithGFM` creates a reusable converter with GFM and the package's
  round-trip extension syntax.
- `NewParser`, `NewWithGFMParser`, and `NewRenderer` expose the native
  Goldmark v2 components for integrations that need to inspect or extend the
  AST.

`*Markdown` converters are safe for concurrent use. Each conversion gets a
fresh parser and renderer context; an application-supplied `ImageHandler` must
itself be safe for concurrent calls.

## ADF output

Every conversion produces an ADF document of this shape:

```json
{
  "version": 1,
  "type": "doc",
  "content": []
}
```

The renderer maps CommonMark blocks and inline content to their ADF
equivalents, including paragraphs, headings, blockquotes, lists, code blocks,
links, images, tables, and formatting marks. `NewWithGFM` additionally enables
tables, strikethrough, autolinks, task lists, and the ADF round-trip syntax.

The renderer prefers schema-correct output over lossy structural conversion.
For example, a standalone `[card:URL]` becomes a `blockCard`, while the same
token inside text becomes an `inlineCard`; a mixed inline `[embed:URL]` stays
literal text because ADF has no inline embed-card node.

## Configuration and validation

Options configure generated output. The table and image layouts are typed:

```go
md := adf.NewWithGFM(
	adf.WithTableLayout(adf.TableLayoutWide),
	adf.WithImageLayout(adf.ImageLayoutFullWidth),
)
```

`WithImageHandler` lets an application provide its own ADF image result. A
configured handler takes precedence over external-media output.

The renderer does not validate every document automatically. Use the optional
`adfschema` package at a boundary that requires validation:

```go
if err := adfschema.Validate(adfJSON); err != nil {
	return err
}
```

## Related documentation

- [Getting started](specs/getting-started.md)
- [Migrating to v2](migrating-to-v2.md)
- [Markdown extensions](extensions.md)
- [Goldmark to ADF node mapping](node-mapping.md)
- [ADF schema coverage](schema-coverage.md)
