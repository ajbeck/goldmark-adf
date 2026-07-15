//go:build goexperiment.jsonv2

package adf

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/ajbeck/goldmark-adf/astext"
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
	document      *Document
	nodeStack     []*Node
	markStack     []Mark
	openTaskItems map[ast.Node]bool
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

	// ADF round-trip extension nodes (inline)
	reg.Register(astext.KindStatus, r.renderStatus)
	reg.Register(astext.KindMention, r.renderMention)
	reg.Register(astext.KindDate, r.renderDate)
	reg.Register(astext.KindPlaceholder, r.renderPlaceholder)
	reg.Register(astext.KindCard, r.renderCard)
	reg.Register(astext.KindEmbed, r.renderEmbed)
	reg.Register(astext.KindEmoji, r.renderEmoji)

	// ADF round-trip extension nodes (block)
	reg.Register(astext.KindPanel, r.renderPanel)
	reg.Register(astext.KindDecisionList, r.renderDecisionList)
	reg.Register(astext.KindDecisionItem, r.renderDecisionItem)
}

// reset prepares the renderer for a new document.
func (r *Renderer) reset() {
	r.document = NewDocument()
	r.nodeStack = []*Node{}
	r.markStack = []Mark{}
	r.openTaskItems = map[ast.Node]bool{}
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

// discardCurrentNode removes the current node from the stack without appending it.
// Used to discard empty paragraphs during image handling.
func (r *Renderer) discardCurrentNode() {
	if len(r.nodeStack) > 0 {
		r.nodeStack = r.nodeStack[:len(r.nodeStack)-1]
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
	list := node.(*ast.List)
	if entering {
		if isTaskList(list) {
			// ADF taskItem nodes can only contain inline content. Goldmark nests a
			// child list beneath its parent ListItem, so close the parent task item
			// before emitting the child task list as its sibling.
			if parentItem, ok := list.Parent().(*ast.ListItem); ok && r.openTaskItems[parentItem] {
				r.popNode()
				r.openTaskItems[parentItem] = false
			}
			r.pushNode(NewTaskList())
		} else if list.IsOrdered() {
			r.pushNode(NewOrderedList(list.Start))
		} else {
			r.pushNode(NewBulletList())
		}
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	item := node.(*ast.ListItem)
	if list, ok := item.Parent().(*ast.List); ok && isTaskList(list) {
		if entering {
			checkBox, _ := taskCheckBox(item)
			state := "TODO"
			if checkBox.IsChecked {
				state = "DONE"
			}
			r.pushNode(NewTaskItem(state))
			r.openTaskItems[item] = true
		} else if r.openTaskItems[item] {
			r.popNode()
			r.openTaskItems[item] = false
		}
		return ast.WalkContinue, nil
	}

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
		// Check if the paragraph is empty and discard it if so
		// This handles cases where images split paragraphs and leave empty ones
		current := r.currentNode()
		if current != nil && current.Type == "paragraph" && len(current.Content) == 0 {
			r.discardCurrentNode()
		} else {
			r.popNode()
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if parentItem, ok := node.Parent().(*ast.ListItem); ok {
		if list, ok := parentItem.Parent().(*ast.List); ok && isTaskList(list) {
			return ast.WalkContinue, nil
		}
	}

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

	title := ""
	if n.Title != nil {
		title = string(n.Title)
	}

	// taskItem content only permits inline nodes, so external media must fall
	// back to linked text even when external media is otherwise enabled.
	if r.config.ExternalMedia && (r.currentNode() == nil || r.currentNode().Type != "taskItem") {
		// Handle external media with paragraph splitting
		r.renderExternalMedia(dest, alt, title)
	} else {
		// Fallback: convert image to a link
		textNode := NewTextWithMarks(alt, []Mark{NewLinkMark(dest, title)})
		r.appendToCurrentOrDocument(*textNode)
	}

	return ast.WalkSkipChildren, nil
}

// renderExternalMedia renders an image as an external media node.
// If we're inside a paragraph, it splits the paragraph around the image.
func (r *Renderer) renderExternalMedia(url, alt, title string) {
	// Check if we're inside a paragraph
	current := r.currentNode()
	if current != nil && current.Type == "paragraph" {
		// If the paragraph has content, pop it (appends to parent)
		// If the paragraph is empty, just discard it
		if len(current.Content) > 0 {
			r.popNode()
		} else {
			r.discardCurrentNode()
		}

		// Emit the mediaSingle with media (and caption if title present)
		r.emitMediaSingle(url, alt, title)

		// Push a new empty paragraph for remaining content
		r.pushNode(NewParagraph())
	} else {
		// Not in a paragraph, just emit mediaSingle directly
		r.emitMediaSingle(url, alt, title)
	}
}

// emitMediaSingle creates and appends a mediaSingle node with the given media content.
func (r *Renderer) emitMediaSingle(url, alt, title string) {
	layout := r.config.ImageLayout
	if layout == "" {
		layout = "center"
	}

	mediaSingle := NewMediaSingle(layout)
	media := NewExternalMedia(url, alt)
	mediaSingle.AppendChild(*media)

	// Add caption if title is provided
	if title != "" {
		caption := NewCaption()
		caption.AppendChild(*NewText(title))
		mediaSingle.AppendChild(*caption)
	}

	r.appendToCurrentOrDocument(*mediaSingle)
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
	// Task checkboxes are handled by the task-aware list rendering.
	// If we reach here, it means the list wasn't converted to a task list
	// (e.g. mixed task/non-task items). Fall back to text prefix.
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*extast.TaskCheckBox)
	if parentTextBlock, ok := n.Parent().(*ast.TextBlock); ok {
		if parentItem, ok := parentTextBlock.Parent().(*ast.ListItem); ok {
			if list, ok := parentItem.Parent().(*ast.List); ok && isTaskList(list) {
				return ast.WalkContinue, nil
			}
		}
	}

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

// taskCheckBox returns the leading task checkbox in a list item, if present.
func taskCheckBox(item *ast.ListItem) (*extast.TaskCheckBox, bool) {
	textBlock, ok := item.FirstChild().(*ast.TextBlock)
	if !ok {
		return nil, false
	}
	checkBox, ok := textBlock.FirstChild().(*extast.TaskCheckBox)
	return checkBox, ok
}

// isTaskList reports whether every item in a list is a task item and every
// nested list can also be represented as an ADF task list. This avoids mixing
// listItem and taskItem nodes in a single ADF taskList.
func isTaskList(list *ast.List) bool {
	if !isTaskListItems(list) {
		return false
	}

	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			switch child := child.(type) {
			case *ast.TextBlock:
				// The leading checkbox and all task-item content are inline.
			case *ast.List:
				if isTaskList(child) {
					continue
				}
				return false
			default:
				// taskItem cannot contain paragraphs or other block nodes.
				return false
			}
		}
	}
	return true
}

func isTaskListItems(list *ast.List) bool {
	if list.FirstChild() == nil {
		return false
	}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		listItem, ok := item.(*ast.ListItem)
		if !ok {
			return false
		}
		if _, ok := taskCheckBox(listItem); !ok {
			return false
		}
	}
	return true
}

// ADF round-trip extension renderers (inline)

func (r *Renderer) renderStatus(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Status)
	r.appendToCurrentOrDocument(*NewStatusNode(n.StatusText, n.Color))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderMention(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Mention)
	r.appendToCurrentOrDocument(*NewMentionNode(n.ID, n.DisplayName))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderDate(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Date)
	r.appendToCurrentOrDocument(*NewDateNode(n.Timestamp))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderPlaceholder(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Placeholder)
	r.appendToCurrentOrDocument(*NewPlaceholderNode(n.Label))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderCard(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Card)
	r.appendToCurrentOrDocument(*NewInlineCardNode(n.URL))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderEmbed(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Embed)
	r.appendToCurrentOrDocument(*NewEmbedCardNode(n.URL))
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderEmoji(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Emoji)
	r.appendToCurrentOrDocument(*NewEmojiNode(n.ShortName))
	return ast.WalkSkipChildren, nil
}

// ADF round-trip extension renderers (block)

func (r *Renderer) renderPanel(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewPanelNode(node.(*astext.Panel).PanelType))
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderDecisionList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushNode(NewDecisionListNode())
	} else {
		r.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderDecisionItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*astext.DecisionItem)
		r.pushNode(NewDecisionItemNode(n.State))
	} else {
		r.popNode()
	}
}

// Ensure we implement the interface
var _ renderer.NodeRenderer = (*Renderer)(nil)
