# goldmark-adf

A [goldmark](https://github.com/yuin/goldmark) renderer that outputs Atlassian Document Format (ADF) JSON instead of HTML.

ADF is the native document format used by Atlassian products like Jira Cloud and Confluence Cloud.

## Requirements

- Go 1.25+
- `GOEXPERIMENT=jsonv2` environment variable (uses experimental `encoding/json/v2`)

## Installation

```bash
go get github.com/ajbeck/goldmark-adf
```

## Usage

### Basic Conversion

```go
package main

import (
    "bytes"
    "fmt"
    "log"

    "github.com/ajbeck/goldmark-adf"
)

func main() {
    // Using convenience function
    output, err := adf.Convert([]byte("# Hello World"))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(output))

    // Using reusable instance
    md := adf.New()
    var buf bytes.Buffer
    if err := md.Convert([]byte("**Bold** text"), &buf); err != nil {
        log.Fatal(err)
    }
    fmt.Println(buf.String())
}
```

### With GFM Extensions

```go
// Enable tables, strikethrough, autolinks, and task lists
md := adf.NewWithGFM()

markdown := []byte(`| Name | Age |
| ---- | --- |
| Alice | 30 |

This has ~~strikethrough~~ text.`)

var buf bytes.Buffer
md.Convert(markdown, &buf)
```

### With External Media Images

By default, images are converted to linked text. To render images as actual media in Atlassian products, enable external media:

```go
md := adf.New(adf.WithExternalMedia(true))

markdown := []byte(`Check out this diagram:

![Architecture](https://example.com/diagram.png)

Pretty cool, right?`)

var buf bytes.Buffer
md.Convert(markdown, &buf)
```

This produces `mediaSingle` nodes that display images inline in Jira and Confluence.

You can also control image layout:

```go
// Options: "center" (default), "wide", "full-width",
//          "wrap-left", "wrap-right", "align-start", "align-end"
md := adf.New(
    adf.WithExternalMedia(true),
    adf.WithImageLayout("wide"),
)
```

## Building and Testing

```bash
# Build
GOEXPERIMENT=jsonv2 go build ./...

# Test
GOEXPERIMENT=jsonv2 go test ./...
```

## Round-Trip with adf-to-markdown

This library is designed to work with [adf-to-markdown](https://github.com/ajbeck/adf-to-markdown) for lossless ADF round-tripping:

```
ADF JSON --> adf-to-markdown --> Markdown --> goldmark-adf --> ADF JSON
```

**adf-to-markdown** converts ADF JSON to Markdown using custom syntax extensions for ADF-specific nodes (status badges, mentions, panels, etc.). **goldmark-adf** parses that Markdown — including the custom syntax — back into ADF JSON.

Use `NewWithGFM()` for round-trip workflows. It enables GFM parsing (tables, strikethrough, task lists) and registers the custom extension parsers:

```go
md := adf.NewWithGFM(
    adf.WithExternalMedia(true), // ![alt](url) -> mediaSingle
)

var buf bytes.Buffer
if err := md.Convert(markdown, &buf); err != nil {
    log.Fatal(err)
}
// buf contains ADF JSON
```

### Custom Extension Syntax

These extensions are parsed by `NewWithGFM()` and correspond to the output of adf-to-markdown:

| Markdown Syntax | ADF Node |
|---|---|
| `[status:text\|color]` | `status` |
| `@[name](id)` | `mention` |
| `[date:1234567890]` | `date` |
| `{{placeholder text}}` | `placeholder` |
| `[card:url]` | `inlineCard` / `blockCard` |
| `[embed:url]` | `embedCard` |
| `:shortcode:` | `emoji` |
| `> [!WARNING]` | `panel` (GitHub alert syntax) |
| `- [!] text` / `- [?] text` | `decisionList` / `decisionItem` |
| `- [x]` / `- [ ]` | `taskList` / `taskItem` |

For the full syntax specification and escaping rules, see the [roundtrip-extensions.md](https://github.com/ajbeck/adf-to-markdown/blob/main/docs/roundtrip-extensions.md) document in adf-to-markdown.

## Supported Markdown Features

### Block Elements
- Headings (1-6)
- Paragraphs
- Blockquotes
- Panels (GitHub alert syntax: `> [!NOTE]`, `> [!WARNING]`, etc.)
- Code blocks (fenced and indented)
- Unordered lists
- Ordered lists
- Decision lists (`- [!]` / `- [?]`)
- Horizontal rules

### Inline Elements
- Bold (`**text**`)
- Italic (`*text*`)
- Inline code (`` `code` ``)
- Links (`[text](url)`)
- Images (converted to links by default, or external media with `WithExternalMedia(true)`)
- Hard breaks
- Status badges (`[status:text|color]`)
- Mentions (`@[name](id)`)
- Dates (`[date:timestamp]`)
- Placeholders (`{{text}}`)
- Cards (`[card:url]`, `[embed:url]`)
- Emoji (`:shortcode:`)

### GFM Extensions (with `NewWithGFM`)
- Tables
- Strikethrough (`~~text~~`)
- Autolinks
- Task lists (`- [x]` / `- [ ]` as `taskList` / `taskItem`)

## Schema Validation

The `adfschema` subpackage provides validation against the official Atlassian ADF JSON Schema:

```go
import "github.com/ajbeck/goldmark-adf/adfschema"

if err := adfschema.Validate(jsonBytes); err != nil {
    log.Printf("Invalid ADF: %v", err)
}
```

## Output Examples

### Basic Markdown

Input:
```markdown
# Hello World

This is **bold** text with a [link](https://example.com).
```

Output:
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "heading",
      "attrs": { "level": 1 },
      "content": [
        { "type": "text", "text": "Hello World" }
      ]
    },
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "This is " },
        {
          "type": "text",
          "marks": [{ "type": "strong" }],
          "text": "bold"
        },
        { "type": "text", "text": " text with a " },
        {
          "type": "text",
          "marks": [{ "type": "link", "attrs": { "href": "https://example.com" } }],
          "text": "link"
        },
        { "type": "text", "text": "." }
      ]
    }
  ]
}
```

### External Media Images

Input (with `WithExternalMedia(true)`):
```markdown
Check this out:

![Diagram](https://example.com/diagram.png)
```

Output:
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Check this out:" }
      ]
    },
    {
      "type": "mediaSingle",
      "attrs": { "layout": "center" },
      "content": [
        {
          "type": "media",
          "attrs": {
            "type": "external",
            "url": "https://example.com/diagram.png",
            "alt": "Diagram"
          }
        }
      ]
    }
  ]
}
```

## Documentation

- [Implementation Plan](docs/specs/getting-started.md)
- [Goldmark to ADF Node Mapping](docs/node-mapping.md)
- [HTML Renderer Patterns](docs/html-renderer-patterns.md)
- [Atlassian Image Handling Research](docs/research/atlassian-image-handling.md)
- [Round-Trip Extension Syntax](https://github.com/ajbeck/adf-to-markdown/blob/main/docs/roundtrip-extensions.md) (in adf-to-markdown)
- [Library Integration Guide](https://github.com/ajbeck/adf-to-markdown/blob/main/docs/library-integration.md) (in adf-to-markdown)

## ADF Resources

- [ADF Structure Documentation](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [ADF JSON Schema](https://unpkg.com/@atlaskit/adf-schema@51.5.6/dist/json-schema/v1/full.json)

## License

MIT
