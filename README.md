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

## Supported Markdown Features

### Block Elements
- Headings (1-6)
- Paragraphs
- Blockquotes
- Code blocks (fenced and indented)
- Unordered lists
- Ordered lists
- Horizontal rules

### Inline Elements
- Bold (`**text**`)
- Italic (`*text*`)
- Inline code (`` `code` ``)
- Links (`[text](url)`)
- Images (converted to links by default, or external media with `WithExternalMedia(true)`)
- Hard breaks

### GFM Extensions (with `NewWithGFM`)
- Tables
- Strikethrough (`~~text~~`)
- Autolinks
- Task lists

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

## ADF Resources

- [ADF Structure Documentation](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [ADF JSON Schema](https://unpkg.com/@atlaskit/adf-schema@51.5.6/dist/json-schema/v1/full.json)

## License

MIT
