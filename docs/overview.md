# go-markdown-adf: Goldmark ADF Renderer

## Overview

This project implements a custom [goldmark](https://github.com/yuin/goldmark) renderer that outputs Atlassian Document Format (ADF) instead of HTML. This enables converting Markdown documents into a format consumable by Atlassian products like Jira Cloud and Confluence Cloud.

## What is Goldmark?

Goldmark is a CommonMark-compliant Markdown parser and renderer written in Go. It provides:

- **Parser**: Converts Markdown text into an Abstract Syntax Tree (AST)
- **Renderer**: Walks the AST and produces output (HTML by default)
- **Extension System**: Allows adding new syntax (GFM tables, strikethrough, etc.)

### Key Architecture Concepts

```
Markdown Text → Parser → AST → Renderer → Output (HTML/ADF/etc.)
```

**Core Interfaces:**

1. `renderer.Renderer` - Main interface with `Render(w io.Writer, source []byte, n ast.Node) error`
2. `renderer.NodeRenderer` - Provides rendering functions via `RegisterFuncs(NodeRendererFuncRegisterer)`
3. `renderer.NodeRendererFunc` - Function signature: `func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error)`

The `entering` parameter is key: it's `true` when first visiting a node (open tag) and `false` when leaving (close tag).

## What is ADF?

Atlassian Document Format is a JSON-based format for representing rich text documents. Every document has:

```json
{
  "version": 1,
  "type": "doc",
  "content": [...]
}
```

### ADF Node Categories

1. **Top-level Block Nodes**: Can appear directly under `doc`
   - `blockquote`, `bulletList`, `codeBlock`, `expand`, `heading`
   - `mediaGroup`, `mediaSingle`, `orderedList`, `panel`, `paragraph`
   - `rule`, `table`

2. **Child Block Nodes**: Must be nested in other nodes
   - `listItem`, `media`, `nestedExpand`
   - `tableCell`, `tableHeader`, `tableRow`

3. **Inline Nodes**: Contain document content
   - `date`, `emoji`, `hardBreak`, `inlineCard`
   - `mention`, `status`, `text`, `mediaInline`

4. **Marks**: Apply formatting to text nodes
   - `code`, `em`, `link`, `strike`, `strong`
   - `subsup`, `textColor`, `underline`

## Project Goals

Build a goldmark renderer that:

1. Walks the goldmark AST nodes
2. Produces valid ADF JSON output
3. Maps Markdown constructs to equivalent ADF nodes
4. Supports core Markdown features (paragraphs, headings, lists, code, emphasis, links, images)
5. Optionally supports GFM extensions (tables, strikethrough, task lists)

## Architecture Design

### Renderer Structure

```go
// ADFRenderer implements renderer.NodeRenderer
type ADFRenderer struct {
    // Configuration options
}

func (r *ADFRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
    // Register a function for each AST node kind
    reg.Register(ast.KindDocument, r.renderDocument)
    reg.Register(ast.KindParagraph, r.renderParagraph)
    reg.Register(ast.KindHeading, r.renderHeading)
    // ... etc
}
```

### JSON Building Strategy

Unlike HTML (which is text-based), ADF is structured JSON. Options:

1. **String Builder**: Build JSON as strings (simple but error-prone)
2. **Struct-based**: Define Go structs matching ADF schema, marshal at end (type-safe)
3. **Hybrid**: Use structs internally, serialize during walk

Recommended: **Struct-based approach** with types like:

```go
type Document struct {
    Version int    `json:"version"`
    Type    string `json:"type"`
    Content []Node `json:"content"`
}

type Node struct {
    Type    string                 `json:"type"`
    Content []Node                 `json:"content,omitempty"`
    Attrs   map[string]interface{} `json:"attrs,omitempty"`
    Marks   []Mark                 `json:"marks,omitempty"`
    Text    string                 `json:"text,omitempty"`
}
```

### Handling the AST Walk

The goldmark renderer walks the AST depth-first, calling render functions with `entering=true` on descent and `entering=false` on ascent. For JSON output:

- On `entering=true`: Create new ADF node, push onto stack
- On `entering=false`: Pop from stack, append to parent's content

## Implementation Phases

### Phase 1: Core Block Nodes
- Document structure (doc wrapper)
- Paragraph
- Heading (levels 1-6)
- Blockquote
- ThematicBreak → rule

### Phase 2: Lists
- BulletList → bulletList
- OrderedList → orderedList
- ListItem → listItem

### Phase 3: Code
- CodeBlock/FencedCodeBlock → codeBlock
- CodeSpan → text with code mark

### Phase 4: Inline Elements
- Text nodes
- Emphasis → em/strong marks
- Link → text with link mark
- Image → mediaSingle/media (requires special handling)

### Phase 5: GFM Extensions
- Tables → table/tableRow/tableHeader/tableCell
- Strikethrough → strike mark
- Task lists → custom handling

## Testing Strategy

1. **Unit tests**: Individual node rendering
2. **Integration tests**: Full document conversion
3. **Validation**: Ensure output matches ADF schema
4. **Roundtrip tests**: Compare with Atlassian's own rendering

## Dependencies

- `github.com/yuin/goldmark` - Core parser and renderer framework

No other external dependencies required for basic functionality.
