# Getting started with goldmark-adf v2

goldmark-adf v2 renders Markdown as
[Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
JSON. It requires Go 1.27 or later.

## Install

```bash
go get github.com/ajbeck/goldmark-adf/v2@v2.0.0
```

The `GOEXPERIMENT=jsonv2` environment variable used by v1 is not required in
v2.

## Convert CommonMark

```go
package main

import (
	"bytes"
	"fmt"

	adf "github.com/ajbeck/goldmark-adf/v2"
)

func main() {
	var output bytes.Buffer
	if err := adf.New().Convert([]byte("# Hello\\n\\nADF output."), &output); err != nil {
		panic(err)
	}
	fmt.Println(output.String())
}
```

`New` converts CommonMark. The returned converter can be reused, including
from concurrent callers.

## Enable GFM and round-trip syntax

Use `NewWithGFM` for GFM tables, strikethrough, autolinks, task lists, and
goldmark-adf's ADF round-trip extensions:

```go
md := adf.NewWithGFM(adf.WithTableLayout(adf.TableLayoutWide))
err := md.Convert(markdown, output)
```

Only unordered task lists that can be represented without structural loss are
emitted as ADF task lists. Ordered or mixed lists retain their checkbox text.
For the complete custom syntax, see [Markdown extensions](../extensions.md).

## Configure images and layouts

Image and table layouts use typed constants:

```go
md := adf.New(
	adf.WithExternalMedia(true),
	adf.WithImageLayout(adf.ImageLayoutWide),
	adf.WithTableLayout(adf.TableLayoutFullWidth),
)
```

An `ImageHandler` can override image conversion when an application needs to
create a specific ADF inline or block result. It takes precedence over
`WithExternalMedia`; see the package documentation for its exact contract.

## Work with native Goldmark components

Use the facade unless an integration needs AST access. Goldmark v2 separates
the parser and renderer, so direct composition is explicit:

```go
p := adf.NewWithGFMParser()
r := adf.NewRenderer()

doc := p.Parse(markdown)
if err := r.Render(output, markdown, doc); err != nil {
	return err
}
```

A parser instance is for one parse at a time. Use `*adf.Markdown` for
concurrent conversion, or create a fresh parser for each parse.

## Validate output when needed

Schema validation is opt-in, so ordinary conversion does not incur its cost:

```go
import "github.com/ajbeck/goldmark-adf/v2/adfschema"

if err := adfschema.Validate(output.Bytes()); err != nil {
	return err
}
```

## Verify a checkout

```bash
go build ./...
go test ./...
go test -race ./...
```

For breaking API and semantic changes from v1, read the
[v2 migration guide](../migrating-to-v2.md).
