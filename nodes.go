//go:build goexperiment.jsonv2

package adf

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Document represents the root ADF document node.
//
// Every ADF document has a version (always 1), a type of "doc", and a content
// array containing the top-level block nodes. Use [NewDocument] to create a
// properly initialized Document.
type Document struct {
	Version int    `json:"version"`
	Type    string `json:"type"`
	Content []Node `json:"content"`
}

// NewDocument creates a new empty ADF document.
func NewDocument() *Document {
	return &Document{
		Version: 1,
		Type:    "doc",
		Content: []Node{},
	}
}

// Node represents an ADF node. It uses a flexible structure that can represent
// any node type in ADF, including block nodes (paragraph, heading, codeBlock),
// inline nodes (hardBreak), and text nodes.
//
// Block and inline nodes use the Content field to hold child nodes.
// Text nodes use the Text field for their content and may have Marks applied.
// The Attrs field holds type-specific attributes like heading level or link href.
//
// Use the constructor functions (e.g., [NewParagraph], [NewHeading], [NewText])
// to create properly initialized nodes.
type Node struct {
	Type    string         `json:"type"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node         `json:"content,omitempty"`
	Marks   []Mark         `json:"marks,omitempty"`
	Text    string         `json:"text,omitempty"`
}

// Mark represents a mark applied to a text node.
//
// Marks are used for inline formatting such as bold ([NewStrongMark]), italic
// ([NewEmMark]), code ([NewCodeMark]), strikethrough ([NewStrikeMark]), and
// links ([NewLinkMark]). Multiple marks can be applied to a single text node.
type Mark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// NewParagraph creates a new paragraph node.
func NewParagraph() *Node {
	return &Node{
		Type:    "paragraph",
		Content: []Node{},
	}
}

// NewHeading creates a new heading node with the specified level (1-6).
func NewHeading(level int) *Node {
	return &Node{
		Type:    "heading",
		Attrs:   map[string]any{"level": level},
		Content: []Node{},
	}
}

// NewBlockquote creates a new blockquote node.
func NewBlockquote() *Node {
	return &Node{
		Type:    "blockquote",
		Content: []Node{},
	}
}

// NewCodeBlock creates a new code block node with an optional language.
func NewCodeBlock(language string) *Node {
	n := &Node{
		Type:    "codeBlock",
		Content: []Node{},
	}
	if language != "" {
		n.Attrs = map[string]any{"language": language}
	}
	return n
}

// NewRule creates a new horizontal rule node.
func NewRule() *Node {
	return &Node{Type: "rule"}
}

// NewBulletList creates a new unordered list node.
func NewBulletList() *Node {
	return &Node{
		Type:    "bulletList",
		Content: []Node{},
	}
}

// NewOrderedList creates a new ordered list node with an optional start number.
func NewOrderedList(start int) *Node {
	n := &Node{
		Type:    "orderedList",
		Content: []Node{},
	}
	if start != 1 {
		n.Attrs = map[string]any{"order": start}
	}
	return n
}

// NewListItem creates a new list item node.
func NewListItem() *Node {
	return &Node{
		Type:    "listItem",
		Content: []Node{},
	}
}

// NewTable creates a new table node.
func NewTable() *Node {
	return &Node{
		Type: "table",
		Attrs: map[string]any{
			"isNumberColumnEnabled": false,
			"layout":                "default",
		},
		Content: []Node{},
	}
}

// NewTableRow creates a new table row node.
func NewTableRow() *Node {
	return &Node{
		Type:    "tableRow",
		Content: []Node{},
	}
}

// NewTableHeader creates a new table header cell node.
func NewTableHeader() *Node {
	return &Node{
		Type:    "tableHeader",
		Attrs:   map[string]any{},
		Content: []Node{},
	}
}

// NewTableCell creates a new table cell node.
func NewTableCell() *Node {
	return &Node{
		Type:    "tableCell",
		Attrs:   map[string]any{},
		Content: []Node{},
	}
}

// NewText creates a new text node with the given content.
func NewText(text string) *Node {
	return &Node{
		Type: "text",
		Text: text,
	}
}

// NewTextWithMarks creates a new text node with marks.
func NewTextWithMarks(text string, marks []Mark) *Node {
	return &Node{
		Type:  "text",
		Text:  text,
		Marks: marks,
	}
}

// NewHardBreak creates a new hard break node.
func NewHardBreak() *Node {
	return &Node{Type: "hardBreak"}
}

// Mark constructors

// NewStrongMark creates a bold/strong mark.
func NewStrongMark() Mark {
	return Mark{Type: "strong"}
}

// NewEmMark creates an italic/emphasis mark.
func NewEmMark() Mark {
	return Mark{Type: "em"}
}

// NewCodeMark creates an inline code mark.
func NewCodeMark() Mark {
	return Mark{Type: "code"}
}

// NewStrikeMark creates a strikethrough mark.
func NewStrikeMark() Mark {
	return Mark{Type: "strike"}
}

// NewUnderlineMark creates an underline mark.
func NewUnderlineMark() Mark {
	return Mark{Type: "underline"}
}

// NewLinkMark creates a link mark with the given URL and optional title.
func NewLinkMark(href, title string) Mark {
	attrs := map[string]any{"href": href}
	if title != "" {
		attrs["title"] = title
	}
	return Mark{
		Type:  "link",
		Attrs: attrs,
	}
}

// NewSubSupMark creates a subscript or superscript mark.
// subType should be "sub" or "sup".
func NewSubSupMark(subType string) Mark {
	return Mark{
		Type:  "subsup",
		Attrs: map[string]any{"type": subType},
	}
}

// NewTextColorMark creates a text color mark with a hex color code.
func NewTextColorMark(color string) Mark {
	return Mark{
		Type:  "textColor",
		Attrs: map[string]any{"color": color},
	}
}

// AppendChild appends a child node to this node's content.
func (n *Node) AppendChild(child Node) {
	n.Content = append(n.Content, child)
}

// AppendMark appends a mark to this node's marks.
func (n *Node) AppendMark(mark Mark) {
	n.Marks = append(n.Marks, mark)
}

// Marshal serializes the document to JSON.
func (d *Document) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// MarshalIndent serializes the document to indented JSON.
func (d *Document) MarshalIndent(indent string) ([]byte, error) {
	return json.Marshal(d, jsontext.WithIndent(indent))
}
