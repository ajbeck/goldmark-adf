# Goldmark AST to ADF Node Mapping

This document describes how Markdown elements (represented as goldmark AST nodes) map to Atlassian Document Format (ADF) nodes.

## Block Nodes

### Document

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindDocument` | `doc` | Root wrapper with `version: 1` |

**Goldmark AST:**
```go
type Document struct {
    BaseBlock
}
```

**ADF Output:**
```json
{
  "version": 1,
  "type": "doc",
  "content": [...]
}
```

---

### Paragraph

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindParagraph` | `paragraph` | Container for inline content |

**Goldmark AST:**
```go
type Paragraph struct {
    BaseBlock
}
```

**ADF Output:**
```json
{
  "type": "paragraph",
  "content": [
    { "type": "text", "text": "Hello world" }
  ]
}
```

---

### Heading

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindHeading` | `heading` | Level 1-6, stored in `attrs.level` |

**Goldmark AST:**
```go
type Heading struct {
    BaseBlock
    Level int  // 1-6
}
```

**ADF Output:**
```json
{
  "type": "heading",
  "attrs": { "level": 1 },
  "content": [
    { "type": "text", "text": "Heading Text" }
  ]
}
```

---

### Blockquote

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindBlockquote` | `blockquote` | Contains paragraphs, lists, code blocks |

**Goldmark AST:**
```go
type Blockquote struct {
    BaseBlock
}
```

**ADF Output:**
```json
{
  "type": "blockquote",
  "content": [
    {
      "type": "paragraph",
      "content": [{ "type": "text", "text": "Quoted text" }]
    }
  ]
}
```

**Allowed Content:** `paragraph`, `bulletList`, `orderedList`, `codeBlock`, `mediaGroup`, `mediaSingle`

---

### Code Block

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindCodeBlock` | `codeBlock` | Indented code block |
| `ast.KindFencedCodeBlock` | `codeBlock` | Fenced with optional language |

**Goldmark AST:**
```go
type FencedCodeBlock struct {
    BaseBlock
    Info *Text  // Language identifier
}
// Access language: n.Language(source)
```

**ADF Output:**
```json
{
  "type": "codeBlock",
  "attrs": { "language": "javascript" },
  "content": [
    { "type": "text", "text": "var foo = 'bar';" }
  ]
}
```

**Note:** ADF code blocks contain text nodes without marks. Language is optional.

---

### Thematic Break (Horizontal Rule)

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindThematicBreak` | `rule` | `---`, `***`, or `___` in Markdown |

**Goldmark AST:**
```go
type ThematicBreak struct {
    BaseBlock
}
```

**ADF Output:**
```json
{
  "type": "rule"
}
```

---

### Lists

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindList` (unordered) | `bulletList` | Uses `-`, `*`, or `+` marker |
| `ast.KindList` (ordered) | `orderedList` | Uses `1.` style markers |
| `ast.KindListItem` | `listItem` | Child of either list type |

**Goldmark AST:**
```go
type List struct {
    BaseBlock
    Marker byte    // '-', '*', '+', ')' or '.'
    IsTight bool
    Start int      // Starting number for ordered lists
}

func (l *List) IsOrdered() bool  // Check if ordered
```

**ADF bulletList:**
```json
{
  "type": "bulletList",
  "content": [
    {
      "type": "listItem",
      "content": [
        {
          "type": "paragraph",
          "content": [{ "type": "text", "text": "Item 1" }]
        }
      ]
    }
  ]
}
```

**ADF orderedList:**
```json
{
  "type": "orderedList",
  "attrs": { "order": 1 },
  "content": [
    {
      "type": "listItem",
      "content": [
        {
          "type": "paragraph",
          "content": [{ "type": "text", "text": "First item" }]
        }
      ]
    }
  ]
}
```

**listItem Allowed Content:** `paragraph`, `bulletList`, `orderedList`, `codeBlock`, `mediaSingle`

---

## Inline Nodes

### Text

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindText` | `text` | Plain text content |

**Goldmark AST:**
```go
type Text struct {
    BaseInline
    Segment textm.Segment  // Position in source
}
// Access value: n.Value(source) or segment.Value(source)
// Check line breaks: n.HardLineBreak(), n.SoftLineBreak()
```

**ADF Output:**
```json
{
  "type": "text",
  "text": "Hello world"
}
```

**With marks:**
```json
{
  "type": "text",
  "text": "Bold text",
  "marks": [{ "type": "strong" }]
}
```

---

### Hard Break

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindText` with `HardLineBreak()` | `hardBreak` | Two spaces + newline or `\` + newline |

**Goldmark:** Check `text.HardLineBreak() == true`

**ADF Output:**
```json
{
  "type": "hardBreak"
}
```

---

### Emphasis (Italic/Bold)

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindEmphasis` (Level=1) | `em` mark | `*text*` or `_text_` |
| `ast.KindEmphasis` (Level=2) | `strong` mark | `**text**` or `__text__` |

**Goldmark AST:**
```go
type Emphasis struct {
    BaseInline
    Level int  // 1 = italic, 2 = bold
}
```

**ADF Output (emphasis):**
```json
{
  "type": "text",
  "text": "italic text",
  "marks": [{ "type": "em" }]
}
```

**ADF Output (strong):**
```json
{
  "type": "text",
  "text": "bold text",
  "marks": [{ "type": "strong" }]
}
```

**Combined:**
```json
{
  "type": "text",
  "text": "bold and italic",
  "marks": [{ "type": "strong" }, { "type": "em" }]
}
```

---

### Code Span (Inline Code)

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindCodeSpan` | `code` mark | `` `code` `` in Markdown |

**Goldmark AST:**
```go
type CodeSpan struct {
    BaseInline
}
// Children are Text nodes with raw content
```

**ADF Output:**
```json
{
  "type": "text",
  "text": "inline code",
  "marks": [{ "type": "code" }]
}
```

**Constraint:** `code` mark can only combine with `link` mark, no other marks.

---

### Link

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindLink` | `link` mark | `[text](url)` syntax |
| `ast.KindAutoLink` | `link` mark | Auto-detected URLs |

**Goldmark AST:**
```go
type Link struct {
    BaseInline
    Destination []byte  // URL
    Title []byte        // Optional title
}
```

**ADF Output:**
```json
{
  "type": "text",
  "text": "Click here",
  "marks": [
    {
      "type": "link",
      "attrs": {
        "href": "https://example.com",
        "title": "Example Site"
      }
    }
  ]
}
```

---

### Image

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `ast.KindImage` | `mediaSingle` + `media` | Complex - requires special handling |

**Goldmark AST:**
```go
type Image struct {
    BaseInline
    Destination []byte  // Image URL
    Title []byte
}
// Alt text is in child Text nodes
```

**ADF Output (external URL):**

Images in ADF are complex. For external URLs, you may need to use `inlineCard` or handle via media services:

```json
{
  "type": "mediaSingle",
  "attrs": { "layout": "center" },
  "content": [
    {
      "type": "media",
      "attrs": {
        "type": "external",
        "url": "https://example.com/image.png"
      }
    }
  ]
}
```

**Alternative:** Convert to link if media services aren't available:
```json
{
  "type": "text",
  "text": "alt text",
  "marks": [{ "type": "link", "attrs": { "href": "https://example.com/image.png" } }]
}
```

---

## GFM Extension Nodes

These require the GFM extension to be enabled in goldmark.

### Strikethrough

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `extension.KindStrikethrough` | `strike` mark | `~~text~~` syntax |

**ADF Output:**
```json
{
  "type": "text",
  "text": "deleted text",
  "marks": [{ "type": "strike" }]
}
```

---

### Table

| Goldmark | ADF | Notes |
|----------|-----|-------|
| `extension.KindTable` | `table` | Table container |
| `extension.KindTableHeader` | `tableRow` with `tableHeader` cells | First row |
| `extension.KindTableRow` | `tableRow` with `tableCell` cells | Body rows |
| `extension.KindTableCell` | `tableCell` or `tableHeader` | Individual cells |

**ADF Output:**
```json
{
  "type": "table",
  "attrs": {
    "isNumberColumnEnabled": false,
    "layout": "center"
  },
  "content": [
    {
      "type": "tableRow",
      "content": [
        {
          "type": "tableHeader",
          "attrs": {},
          "content": [
            {
              "type": "paragraph",
              "content": [{ "type": "text", "text": "Header 1" }]
            }
          ]
        }
      ]
    },
    {
      "type": "tableRow",
      "content": [
        {
          "type": "tableCell",
          "attrs": {},
          "content": [
            {
              "type": "paragraph",
              "content": [{ "type": "text", "text": "Cell 1" }]
            }
          ]
        }
      ]
    }
  ]
}
```

**Table Cell Attributes:**
- `background`: Hex color code
- `colspan`: Number of columns to span
- `rowspan`: Number of rows to span
- `colwidth`: Array of column widths in pixels

---

## ADF Marks Summary

| Mark | Purpose | Attributes |
|------|---------|------------|
| `strong` | Bold text | None |
| `em` | Italic text | None |
| `strike` | Strikethrough | None |
| `underline` | Underlined text | None |
| `code` | Inline code | None |
| `link` | Hyperlink | `href` (required), `title` |
| `subsup` | Subscript/superscript | `type`: `"sub"` or `"sup"` |
| `textColor` | Colored text | `color`: hex code |

---

## Nodes Without Direct Markdown Equivalent

These ADF nodes don't have standard Markdown equivalents:

| ADF Node | Purpose | Possible Mapping |
|----------|---------|------------------|
| `panel` | Highlighted content box | Could map from custom syntax or admonitions |
| `expand` | Collapsible section | Could map from `<details>` HTML |
| `emoji` | Emoji characters | Could detect `:shortcode:` patterns |
| `mention` | User mentions | Could detect `@username` patterns |
| `status` | Status badges | No standard equivalent |
| `date` | Date picker value | No standard equivalent |
| `inlineCard` | Rich link preview | Could upgrade from regular links |

---

## Implementation Considerations

### Mark Accumulation

Unlike HTML where tags nest, ADF marks are a flat array on text nodes. When traversing emphasis/link nodes:

1. Track active marks in a stack
2. When reaching a text node, apply all accumulated marks
3. Pop marks when leaving emphasis/link nodes

### Nested Structures

Some ADF constraints to handle:

1. **Blockquote content**: Only allows paragraph, lists, codeBlock, media - not headings
2. **ListItem content**: Text must be wrapped in paragraph
3. **Code marks**: Cannot combine with most other marks (only link)
4. **Panel content**: No marks allowed on paragraph/heading children

### Text Node Handling

Goldmark's `ast.Text` nodes include:
- `HardLineBreak()`: Convert to ADF `hardBreak`
- `SoftLineBreak()`: Usually becomes space in output
- `IsRaw()`: Used in code blocks

Access text value: `segment.Value(source)` where `source` is the original Markdown bytes.
