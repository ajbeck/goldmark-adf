# Goldmark HTML Renderer Patterns

This document analyzes key patterns from goldmark's HTML renderer (`renderer/html/html.go`) that we should follow when building the ADF renderer.

## Core Renderer Structure

### Registration Pattern

```go
type Renderer struct {
    Config
}

func NewRenderer(opts ...Option) renderer.NodeRenderer {
    r := &Renderer{
        Config: NewConfig(),
    }
    for _, opt := range opts {
        opt.SetHTMLOption(&r.Config)
    }
    return r
}

func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
    // Block nodes
    reg.Register(ast.KindDocument, r.renderDocument)
    reg.Register(ast.KindHeading, r.renderHeading)
    reg.Register(ast.KindParagraph, r.renderParagraph)
    // ... etc

    // Inline nodes
    reg.Register(ast.KindText, r.renderText)
    reg.Register(ast.KindEmphasis, r.renderEmphasis)
    reg.Register(ast.KindLink, r.renderLink)
    // ... etc
}
```

### Render Function Signature

All render functions have this signature:

```go
func (r *Renderer) renderNodeType(
    w util.BufWriter,
    source []byte,
    node ast.Node,
    entering bool,
) (ast.WalkStatus, error)
```

- `w`: Buffer to write output
- `source`: Original Markdown source bytes
- `node`: Current AST node
- `entering`: `true` when descending into node, `false` when leaving

### Return Values

- `ast.WalkContinue`: Continue normal tree traversal
- `ast.WalkSkipChildren`: Don't visit children (useful when rendering children manually)
- `ast.WalkStop`: Stop entire traversal (usually on error)

## Key Patterns

### Pattern 1: Opening/Closing Tags

For container nodes, write opening content on enter, closing on exit:

```go
func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
    if entering {
        _, _ = w.WriteString("<blockquote>\n")
    } else {
        _, _ = w.WriteString("</blockquote>\n")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 2: Self-Closing / No Children

For nodes without children, do everything on enter:

```go
func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
    if !entering {
        return ast.WalkContinue, nil
    }
    _, _ = w.WriteString("<hr>\n")
    return ast.WalkContinue, nil
}
```

### Pattern 3: Type Assertion for Node Data

Access node-specific fields by type assertion:

```go
func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
    n := node.(*ast.Heading)
    if entering {
        _, _ = w.WriteString("<h")
        _ = w.WriteByte("0123456"[n.Level])
        _ = w.WriteByte('>')
    } else {
        _, _ = w.WriteString("</h")
        _ = w.WriteByte("0123456"[n.Level])
        _, _ = w.WriteString(">\n")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 4: Manual Child Handling

For nodes needing special child handling, use `WalkSkipChildren`:

```go
func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
    if entering {
        _, _ = w.WriteString("<code>")
        // Manually handle children
        for c := n.FirstChild(); c != nil; c = c.NextSibling() {
            segment := c.(*ast.Text).Segment
            value := segment.Value(source)
            r.Writer.RawWrite(w, value)
        }
        return ast.WalkSkipChildren, nil  // Don't auto-visit children
    }
    _, _ = w.WriteString("</code>")
    return ast.WalkContinue, nil
}
```

### Pattern 5: Reading Text Content

Text nodes store position in source, not the actual text:

```go
func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
    if !entering {
        return ast.WalkContinue, nil
    }
    n := node.(*ast.Text)
    segment := n.Segment
    value := segment.Value(source)  // Get actual text bytes

    // Handle line breaks
    if n.HardLineBreak() {
        _, _ = w.WriteString("<br>\n")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 6: Accessing Link/Image Attributes

Link and Image nodes have Destination and Title:

```go
func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
    n := node.(*ast.Link)
    if entering {
        _, _ = w.WriteString("<a href=\"")
        _, _ = w.Write(util.URLEscape(n.Destination, true))
        _ = w.WriteByte('"')
        if n.Title != nil {
            _, _ = w.WriteString(` title="`)
            r.Writer.Write(w, n.Title)
            _ = w.WriteByte('"')
        }
        _ = w.WriteByte('>')
    } else {
        _, _ = w.WriteString("</a>")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 7: List Type Detection

```go
func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
    n := node.(*ast.List)
    tag := "ul"
    if n.IsOrdered() {
        tag = "ol"
    }
    if entering {
        _, _ = w.WriteString("<" + tag)
        if n.IsOrdered() && n.Start != 1 {
            _, _ = fmt.Fprintf(w, " start=\"%d\"", n.Start)
        }
        _, _ = w.WriteString(">\n")
    } else {
        _, _ = w.WriteString("</" + tag + ">\n")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 8: Code Block Language

```go
func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
    n := node.(*ast.FencedCodeBlock)
    if entering {
        _, _ = w.WriteString("<pre><code")
        language := n.Language(source)  // Get language identifier
        if language != nil {
            _, _ = w.WriteString(" class=\"language-")
            r.Writer.Write(w, language)
            _, _ = w.WriteString("\"")
        }
        _ = w.WriteByte('>')
        r.writeLines(w, source, n)  // Helper to write all lines
    } else {
        _, _ = w.WriteString("</code></pre>\n")
    }
    return ast.WalkContinue, nil
}
```

### Pattern 9: Writing Code Block Lines

```go
func (r *Renderer) writeLines(w util.BufWriter, source []byte, n ast.Node) {
    l := n.Lines().Len()
    for i := 0; i < l; i++ {
        line := n.Lines().At(i)
        r.Writer.RawWrite(w, line.Value(source))
    }
}
```

## ADF-Specific Adaptations Needed

### JSON vs String Output

HTML renderer writes strings directly. For ADF, we need to build structured JSON. Options:

1. **Build structs, marshal at end**: Collect nodes in Go structs, `json.Marshal()` after walk
2. **Streaming JSON**: Write JSON tokens as we go (more complex)
3. **Node stack**: Push/pop nodes during walk, serialize collected structure

### Mark Accumulation

HTML nests tags: `<strong><em>text</em></strong>`
ADF flattens marks: `{ "text": "...", "marks": [{"type":"strong"}, {"type":"em"}] }`

Need to:
1. Track active marks in a stack during traversal
2. Apply all active marks when reaching text nodes

### Context-Aware Rendering

Some ADF nodes have constraints. May need context tracking:
- Are we inside a blockquote? (limits allowed children)
- Are we inside a code block? (no marks allowed)
- Are we inside a list item? (content must be wrapped in paragraph)

### Suggested Renderer State

```go
type ADFRenderer struct {
    Config

    // For JSON building
    document    *ADFDocument
    nodeStack   []*ADFNode

    // For mark tracking
    markStack   []ADFMark

    // For context
    inCodeBlock bool
    inBlockquote bool
}
```
