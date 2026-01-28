//go:build goexperiment.jsonv2

package adf

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Renderer is a goldmark [renderer.NodeRenderer] that outputs Atlassian Document
// Format (ADF) JSON.
//
// The Renderer maintains internal state during the AST walk, using a node stack
// to track the current position in the ADF document tree and a mark stack to
// accumulate active text marks (bold, italic, links, etc.). This state is reset
// for each new document.
//
// Use [NewRenderer] to create a Renderer, or use the higher-level [New] and
// [NewWithGFM] functions which configure a complete goldmark instance.
type Renderer struct {
	config Config

	// State for rendering
	document  *Document
	nodeStack []*Node
	markStack []Mark
}

// NewRenderer creates a new ADF renderer with the given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		config: NewConfig(),
	}
	for _, opt := range opts {
		opt.SetADFOption(&r.config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Block nodes
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// Inline nodes
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)

	// GFM extension nodes
	reg.Register(extast.KindTable, r.renderTable)
	reg.Register(extast.KindTableHeader, r.renderTableHeader)
	reg.Register(extast.KindTableRow, r.renderTableRow)
	reg.Register(extast.KindTableCell, r.renderTableCell)
	reg.Register(extast.KindStrikethrough, r.renderStrikethrough)
	reg.Register(extast.KindTaskCheckBox, r.renderTaskCheckBox)
}

// reset prepares the renderer for a new document.
func (r *Renderer) reset() {
	r.document = NewDocument()
	r.nodeStack = []*Node{}
	r.markStack = []Mark{}
}

// currentNode returns the current node being built, or nil if at document level.
func (r *Renderer) currentNode() *Node {
	if len(r.nodeStack) == 0 {
		return nil
	}
	return r.nodeStack[len(r.nodeStack)-1]
}

// pushNode pushes a new node onto the stack.
func (r *Renderer) pushNode(n *Node) {
	r.nodeStack = append(r.nodeStack, n)
}

// popNode pops the current node from the stack and appends it to its parent.
func (r *Renderer) popNode() {
	if len(r.nodeStack) == 0 {
		return
	}
	n := r.nodeStack[len(r.nodeStack)-1]
	r.nodeStack = r.nodeStack[:len(r.nodeStack)-1]

	if len(r.nodeStack) > 0 {
		parent := r.nodeStack[len(r.nodeStack)-1]
		parent.AppendChild(*n)
	} else {
		r.document.Content = append(r.document.Content, *n)
	}
}

// appendToCurrentOrDocument appends a node to the current node or document.
func (r *Renderer) appendToCurrentOrDocument(n Node) {
	if len(r.nodeStack) > 0 {
		r.nodeStack[len(r.nodeStack)-1].AppendChild(n)
	} else {
		r.document.Content = append(r.document.Content, n)
	}
}

// pushMark adds a mark to the current mark stack.
func (r *Renderer) pushMark(m Mark) {
	r.markStack = append(r.markStack, m)
}

// popMark removes the last mark from the stack.
func (r *Renderer) popMark() {
	if len(r.markStack) > 0 {
		r.markStack = r.markStack[:len(r.markStack)-1]
	}
}

// currentMarks returns a copy of the current marks.
func (r *Renderer) currentMarks() []Mark {
	if len(r.markStack) == 0 {
		return nil
	}
	marks := make([]Mark, len(r.markStack))
	copy(marks, r.markStack)
	return marks
}

// Block node renderers

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.reset()
	} else {
		// Write the final JSON output
		data, err := json.Marshal(r.document, jsontext.WithIndent("  "))
		if err != nil {
			return ast.WalkStop, err
		}
		_, err = w.Write(data)
		if err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Heading)
		r.pushNode(NewHeading(n.Level))
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewBlockquote())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := NewCodeBlock("")
		// Collect all lines as text content
		lines := node.Lines()
		var text string
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			text += string(line.Value(source))
		}
		if text != "" {
			n.AppendChild(*NewText(text))
		}
		r.appendToCurrentOrDocument(*n)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.FencedCodeBlock)
		lang := ""
		if n.Info != nil {
			lang = string(n.Language(source))
		}
		codeNode := NewCodeBlock(lang)
		// Collect all lines as text content
		lines := n.Lines()
		var text string
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			text += string(line.Value(source))
		}
		if text != "" {
			codeNode.AppendChild(*NewText(text))
		}
		r.appendToCurrentOrDocument(*codeNode)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// HTML blocks are not supported in ADF, skip them
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.List)
		if n.IsOrdered() {
			r.pushNode(NewOrderedList(n.Start))
		} else {
			r.pushNode(NewBulletList())
		}
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewListItem())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewParagraph())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// TextBlock is a lightweight paragraph used in tight lists
	// In ADF, we still need to wrap content in a paragraph
	if entering {
		r.pushNode(NewParagraph())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.appendToCurrentOrDocument(*NewRule())
	}
	return ast.WalkContinue, nil
}

// Inline node renderers

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.AutoLink)
		url := string(n.URL(source))
		label := string(n.Label(source))

		textNode := NewTextWithMarks(label, []Mark{NewLinkMark(url, "")})
		r.appendToCurrentOrDocument(*textNode)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushMark(NewCodeMark())
	} else {
		r.popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	if entering {
		if n.Level == 2 {
			r.pushMark(NewStrongMark())
		} else {
			r.pushMark(NewEmMark())
		}
	} else {
		r.popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	dest := string(n.Destination)

	// Get alt text from children
	alt := ""
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			alt += string(t.Segment.Value(source))
		}
	}
	if alt == "" {
		alt = dest
	}

	// Convert image to a link (as per plan)
	title := ""
	if n.Title != nil {
		title = string(n.Title)
	}
	textNode := NewTextWithMarks(alt, []Mark{NewLinkMark(dest, title)})
	r.appendToCurrentOrDocument(*textNode)

	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		title := ""
		if n.Title != nil {
			title = string(n.Title)
		}
		r.pushMark(NewLinkMark(string(n.Destination), title))
	} else {
		r.popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Raw HTML is not supported in ADF, skip it
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	text := string(segment.Value(source))

	if text != "" {
		marks := r.currentMarks()
		var textNode *Node
		if len(marks) > 0 {
			textNode = NewTextWithMarks(text, marks)
		} else {
			textNode = NewText(text)
		}
		r.appendToCurrentOrDocument(*textNode)
	}

	// Handle hard line break
	if n.HardLineBreak() {
		r.appendToCurrentOrDocument(*NewHardBreak())
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	text := string(n.Value)

	if text != "" {
		marks := r.currentMarks()
		var textNode *Node
		if len(marks) > 0 {
			textNode = NewTextWithMarks(text, marks)
		} else {
			textNode = NewText(text)
		}
		r.appendToCurrentOrDocument(*textNode)
	}

	return ast.WalkContinue, nil
}

// GFM extension renderers

func (r *Renderer) renderTable(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewTable())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableHeader(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewTableRow())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableRow(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewTableRow())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableCell(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*extast.TableCell)
		// Determine if this is a header cell based on parent
		parent := n.Parent()
		if _, isHeader := parent.(*extast.TableHeader); isHeader {
			cell := NewTableHeader()
			r.pushNode(cell)
			// Table cells need paragraph wrapper
			r.pushNode(NewParagraph())
		} else {
			cell := NewTableCell()
			r.pushNode(cell)
			// Table cells need paragraph wrapper
			r.pushNode(NewParagraph())
		}
	} else {
		// Pop the paragraph
		r.popNode()
		// Pop the cell
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrikethrough(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushMark(NewStrikeMark())
	} else {
		r.popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTaskCheckBox(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*extast.TaskCheckBox)
	// Render checkbox as text prefix
	var text string
	if n.IsChecked {
		text = "[x] "
	} else {
		text = "[ ] "
	}
	r.appendToCurrentOrDocument(*NewText(text))
	return ast.WalkContinue, nil
}

// Ensure we implement the interface
var _ renderer.NodeRenderer = (*Renderer)(nil)
