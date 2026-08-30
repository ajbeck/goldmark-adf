// Package astext defines AST node types for ADF round-trip markdown extensions.
package astext

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
)

// Inline node kinds.
var (
	KindStatus      = ast.NewNodeKind("Status")
	KindMention     = ast.NewNodeKind("Mention")
	KindDate        = ast.NewNodeKind("Date")
	KindPlaceholder = ast.NewNodeKind("Placeholder")
	KindCard        = ast.NewNodeKind("Card")
	KindEmbed       = ast.NewNodeKind("Embed")
	KindEmoji       = ast.NewNodeKind("Emoji")
)

// Block node kinds.
var (
	KindPanel        = ast.NewNodeKind("Panel")
	KindDecisionList = ast.NewNodeKind("DecisionList")
	KindDecisionItem = ast.NewNodeKind("DecisionItem")
)

// Status represents an ADF status inline node: [status:text|color].
type Status struct {
	ast.BaseInline
	Text  text.SingleLineValue
	Color string
}

// NewStatus creates a Status node.
func NewStatus(value text.SingleLineValue, color string) *Status {
	n := &Status{Text: value, Color: color}
	n.Init(n)
	return n
}

func (n *Status) Kind() ast.NodeKind { return KindStatus }

// Dump implements ast.Node.Dump.
func (n *Status) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"Text": n.Text.Value(source), "Color": n.Color})
}

// Mention represents an ADF mention inline node: @[name](id).
type Mention struct {
	ast.BaseInline
	DisplayName text.SingleLineValue
	ID          text.SingleLineValue
}

// NewMention creates a Mention node.
func NewMention(displayName, id text.SingleLineValue) *Mention {
	n := &Mention{DisplayName: displayName, ID: id}
	n.Init(n)
	return n
}

func (n *Mention) Kind() ast.NodeKind { return KindMention }

// Dump implements ast.Node.Dump.
func (n *Mention) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{
		"DisplayName": n.DisplayName.Value(source),
		"ID":          n.ID.Value(source),
	})
}

// Date represents an ADF date inline node: [date:timestamp].
type Date struct {
	ast.BaseInline
	Timestamp text.SingleLineValue
}

// NewDate creates a Date node.
func NewDate(timestamp text.SingleLineValue) *Date {
	n := &Date{Timestamp: timestamp}
	n.Init(n)
	return n
}

func (n *Date) Kind() ast.NodeKind { return KindDate }

// Dump implements ast.Node.Dump.
func (n *Date) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"Timestamp": n.Timestamp.Value(source)})
}

// Placeholder represents an ADF placeholder inline node: {{text}}.
type Placeholder struct {
	ast.BaseInline
	Label text.SingleLineValue
}

// NewPlaceholder creates a Placeholder node.
func NewPlaceholder(label text.SingleLineValue) *Placeholder {
	n := &Placeholder{Label: label}
	n.Init(n)
	return n
}

func (n *Placeholder) Kind() ast.NodeKind { return KindPlaceholder }

// Dump implements ast.Node.Dump.
func (n *Placeholder) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"Label": n.Label.Value(source)})
}

// Card represents an ADF inline or block card node: [card:url].
type Card struct {
	ast.BaseInline
	URL text.SingleLineValue
}

// NewCard creates a Card node.
func NewCard(url text.SingleLineValue) *Card {
	n := &Card{URL: url}
	n.Init(n)
	return n
}

func (n *Card) Kind() ast.NodeKind { return KindCard }

// Dump implements ast.Node.Dump.
func (n *Card) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"URL": n.URL.Value(source)})
}

// Embed represents an ADF embed card node: [embed:url].
type Embed struct {
	ast.BaseInline
	URL text.SingleLineValue
}

// NewEmbed creates an Embed node.
func NewEmbed(url text.SingleLineValue) *Embed {
	n := &Embed{URL: url}
	n.Init(n)
	return n
}

func (n *Embed) Kind() ast.NodeKind { return KindEmbed }

// Dump implements ast.Node.Dump.
func (n *Embed) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"URL": n.URL.Value(source)})
}

// Emoji represents an ADF emoji inline node: :shortcode:.
type Emoji struct {
	ast.BaseInline
	ShortName text.SingleLineValue
}

// NewEmoji creates an Emoji node.
func NewEmoji(shortName text.SingleLineValue) *Emoji {
	n := &Emoji{ShortName: shortName}
	n.Init(n)
	return n
}

func (n *Emoji) Kind() ast.NodeKind { return KindEmoji }

// Dump implements ast.Node.Dump.
func (n *Emoji) Dump(source []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"ShortName": n.ShortName.Value(source)})
}

// Panel represents an ADF panel block node parsed from GitHub alert syntax.
// Children are the blockquote content.
type Panel struct {
	ast.BaseBlock
	PanelType string
}

// NewPanel creates a Panel node.
func NewPanel(panelType string) *Panel {
	n := &Panel{PanelType: panelType}
	n.Init(n)
	return n
}

func (n *Panel) Kind() ast.NodeKind { return KindPanel }

// Dump implements ast.Node.Dump.
func (n *Panel) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"PanelType": n.PanelType})
}

// DecisionList represents an ADF decision list block node.
type DecisionList struct {
	ast.BaseBlock
}

// NewDecisionList creates a DecisionList node.
func NewDecisionList() *DecisionList {
	n := &DecisionList{}
	n.Init(n)
	return n
}

func (n *DecisionList) Kind() ast.NodeKind { return KindDecisionList }

// Dump implements ast.Node.Dump.
func (n *DecisionList) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, nil)
}

// DecisionItem represents an ADF decision item within a DecisionList.
type DecisionItem struct {
	ast.BaseBlock
	State string
}

// NewDecisionItem creates a DecisionItem node.
func NewDecisionItem(state string) *DecisionItem {
	n := &DecisionItem{State: state}
	n.Init(n)
	return n
}

func (n *DecisionItem) Kind() ast.NodeKind { return KindDecisionItem }

// Dump implements ast.Node.Dump.
func (n *DecisionItem) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"State": n.State})
}
