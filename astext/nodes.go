// Package astext defines AST node types for ADF round-trip markdown extensions.
package astext

import "github.com/yuin/goldmark/ast"

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

// Status represents an ADF status inline node: [status:text|color]
type Status struct {
	ast.BaseInline
	StatusText string
	Color      string
}

func NewStatus(text, color string) *Status {
	return &Status{StatusText: text, Color: color}
}

func (n *Status) Kind() ast.NodeKind { return KindStatus }
func (n *Status) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"StatusText": n.StatusText, "Color": n.Color}, nil)
}

// Mention represents an ADF mention inline node: @[name](id)
type Mention struct {
	ast.BaseInline
	DisplayName string
	ID          string
}

func NewMention(displayName, id string) *Mention {
	return &Mention{DisplayName: displayName, ID: id}
}

func (n *Mention) Kind() ast.NodeKind { return KindMention }
func (n *Mention) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"DisplayName": n.DisplayName, "ID": n.ID}, nil)
}

// Date represents an ADF date inline node: [date:timestamp]
type Date struct {
	ast.BaseInline
	Timestamp string
}

func NewDate(timestamp string) *Date {
	return &Date{Timestamp: timestamp}
}

func (n *Date) Kind() ast.NodeKind { return KindDate }
func (n *Date) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Timestamp": n.Timestamp}, nil)
}

// Placeholder represents an ADF placeholder inline node: {{text}}
type Placeholder struct {
	ast.BaseInline
	Label string
}

func NewPlaceholder(label string) *Placeholder {
	return &Placeholder{Label: label}
}

func (n *Placeholder) Kind() ast.NodeKind { return KindPlaceholder }
func (n *Placeholder) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Label": n.Label}, nil)
}

// Card represents an ADF inline/block card node: [card:url]
type Card struct {
	ast.BaseInline
	URL string
}

func NewCard(url string) *Card {
	return &Card{URL: url}
}

func (n *Card) Kind() ast.NodeKind { return KindCard }
func (n *Card) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"URL": n.URL}, nil)
}

// Embed represents an ADF embed card node: [embed:url]
type Embed struct {
	ast.BaseInline
	URL string
}

func NewEmbed(url string) *Embed {
	return &Embed{URL: url}
}

func (n *Embed) Kind() ast.NodeKind { return KindEmbed }
func (n *Embed) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"URL": n.URL}, nil)
}

// Emoji represents an ADF emoji inline node: :shortcode:
type Emoji struct {
	ast.BaseInline
	ShortName string
}

func NewEmoji(shortName string) *Emoji {
	return &Emoji{ShortName: shortName}
}

func (n *Emoji) Kind() ast.NodeKind { return KindEmoji }
func (n *Emoji) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"ShortName": n.ShortName}, nil)
}

// Panel represents an ADF panel block node parsed from GitHub alert syntax.
// Children are the blockquote content.
type Panel struct {
	ast.BaseBlock
	PanelType string
}

func NewPanel(panelType string) *Panel {
	return &Panel{PanelType: panelType}
}

func (n *Panel) Kind() ast.NodeKind { return KindPanel }
func (n *Panel) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"PanelType": n.PanelType}, nil)
}

// DecisionList represents an ADF decision list block node.
type DecisionList struct {
	ast.BaseBlock
}

func NewDecisionList() *DecisionList {
	return &DecisionList{}
}

func (n *DecisionList) Kind() ast.NodeKind { return KindDecisionList }
func (n *DecisionList) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// DecisionItem represents an ADF decision item within a DecisionList.
type DecisionItem struct {
	ast.BaseBlock
	State string // "DECIDED" or other
}

func NewDecisionItem(state string) *DecisionItem {
	return &DecisionItem{State: state}
}

func (n *DecisionItem) Kind() ast.NodeKind { return KindDecisionItem }
func (n *DecisionItem) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"State": n.State}, nil)
}
