package adf

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"strings"

	"github.com/ajbeck/goldmark-adf/v2/astext"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	extast "github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/renderer"
)

// Renderer renders Goldmark's AST as Atlassian Document Format (ADF) JSON.
// A Renderer is safe for concurrent use after construction, provided a
// configured ImageHandler is safe for concurrent calls.
type Renderer struct {
	*renderer.Helper[io.Writer, Config]
}

type renderState struct {
	document            *Document
	nodeStack           []*Node
	markStack           []Mark
	openTaskItems       map[ast.Node]bool
	suppressedParagraph map[ast.Node]bool
	splitParagraph      map[ast.Node]bool
}

var renderStateKey = renderer.NewContextKey()

// NewRenderer creates a reusable ADF renderer with the given options.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{}
	builder := renderer.HelperBuilder[io.Writer, Config]{}
	options := make([]renderer.Option[Config], 0, len(opts)+1)
	for _, opt := range opts {
		options = append(options, opt)
	}
	options = append(options, renderer.WithNodeRenderers[io.Writer, Config](r.nodeRenderers()))
	r.Helper = builder.Options(options...).OnBeforeRender(func(_ io.Writer, _ []byte, _ ast.Node, rc renderer.Context) error {
		if err := r.Config().validate(); err != nil {
			return err
		}
		rc.Set(renderStateKey, newRenderState())
		return nil
	}).Build()
	return r
}

func newRenderState() *renderState {
	return &renderState{
		document:            NewDocument(),
		openTaskItems:       make(map[ast.Node]bool),
		suppressedParagraph: make(map[ast.Node]bool),
		splitParagraph:      make(map[ast.Node]bool),
	}
}

func (r *Renderer) nodeRenderers() map[ast.NodeKind]renderer.NodeRenderer[io.Writer] {
	return map[ast.NodeKind]renderer.NodeRenderer[io.Writer]{
		ast.KindDocument:      renderer.NodeRendererFunc[io.Writer](r.renderDocument),
		ast.KindHeading:       renderer.NodeRendererFunc[io.Writer](r.renderHeading),
		ast.KindBlockquote:    renderer.NodeRendererFunc[io.Writer](r.renderBlockquote),
		ast.KindCodeBlock:     renderer.NodeRendererFunc[io.Writer](r.renderCodeBlock),
		ast.KindHTMLBlock:     renderer.NodeRendererFunc[io.Writer](r.renderHTMLBlock),
		ast.KindList:          renderer.NodeRendererFunc[io.Writer](r.renderList),
		ast.KindListItem:      renderer.NodeRendererFunc[io.Writer](r.renderListItem),
		ast.KindParagraph:     renderer.NodeRendererFunc[io.Writer](r.renderParagraph),
		ast.KindThematicBreak: renderer.NodeRendererFunc[io.Writer](r.renderThematicBreak),

		ast.KindAutoLink: renderer.NodeRendererFunc[io.Writer](r.renderAutoLink),
		ast.KindCodeSpan: renderer.NodeRendererFunc[io.Writer](r.renderCodeSpan),
		ast.KindEmphasis: renderer.NodeRendererFunc[io.Writer](r.renderEmphasis),
		ast.KindStrong:   renderer.NodeRendererFunc[io.Writer](r.renderStrong),
		ast.KindImage:    renderer.NodeRendererFunc[io.Writer](r.renderImage),
		ast.KindLink:     renderer.NodeRendererFunc[io.Writer](r.renderLink),
		ast.KindRawHTML:  renderer.NodeRendererFunc[io.Writer](r.renderRawHTML),
		ast.KindText:     renderer.NodeRendererFunc[io.Writer](r.renderText),

		extast.KindTable:         renderer.NodeRendererFunc[io.Writer](r.renderTable),
		extast.KindTableHeader:   renderer.NodeRendererFunc[io.Writer](r.renderTableHeader),
		extast.KindTableBody:     renderer.NodeRendererFunc[io.Writer](r.renderTableBody),
		extast.KindTableRow:      renderer.NodeRendererFunc[io.Writer](r.renderTableRow),
		extast.KindTableCell:     renderer.NodeRendererFunc[io.Writer](r.renderTableCell),
		extast.KindStrikethrough: renderer.NodeRendererFunc[io.Writer](r.renderStrikethrough),

		astext.KindStatus:       renderer.NodeRendererFunc[io.Writer](r.renderStatus),
		astext.KindMention:      renderer.NodeRendererFunc[io.Writer](r.renderMention),
		astext.KindDate:         renderer.NodeRendererFunc[io.Writer](r.renderDate),
		astext.KindPlaceholder:  renderer.NodeRendererFunc[io.Writer](r.renderPlaceholder),
		astext.KindCard:         renderer.NodeRendererFunc[io.Writer](r.renderCard),
		astext.KindEmbed:        renderer.NodeRendererFunc[io.Writer](r.renderEmbed),
		astext.KindEmoji:        renderer.NodeRendererFunc[io.Writer](r.renderEmoji),
		astext.KindPanel:        renderer.NodeRendererFunc[io.Writer](r.renderPanel),
		astext.KindDecisionList: renderer.NodeRendererFunc[io.Writer](r.renderDecisionList),
		astext.KindDecisionItem: renderer.NodeRendererFunc[io.Writer](r.renderDecisionItem),
	}
}

func state(rc renderer.Context) *renderState {
	s, ok := rc.Get(renderStateKey).(*renderState)
	if !ok || s == nil {
		panic("adf: missing render state")
	}
	return s
}

func (s *renderState) currentNode() *Node {
	if len(s.nodeStack) == 0 {
		return nil
	}
	return s.nodeStack[len(s.nodeStack)-1]
}

func (s *renderState) pushNode(n *Node) { s.nodeStack = append(s.nodeStack, n) }

func (s *renderState) popNode() {
	if len(s.nodeStack) == 0 {
		return
	}
	n := s.nodeStack[len(s.nodeStack)-1]
	s.nodeStack = s.nodeStack[:len(s.nodeStack)-1]
	if parent := s.currentNode(); parent != nil {
		parent.AppendChild(*n)
		return
	}
	s.document.Content = append(s.document.Content, *n)
}

func (s *renderState) discardCurrentNode() {
	if len(s.nodeStack) != 0 {
		s.nodeStack = s.nodeStack[:len(s.nodeStack)-1]
	}
}

func (s *renderState) appendNode(n Node) {
	if parent := s.currentNode(); parent != nil {
		parent.AppendChild(n)
		return
	}
	s.document.Content = append(s.document.Content, n)
}

func (s *renderState) pushMark(m Mark) { s.markStack = append(s.markStack, m) }

func (s *renderState) popMark() {
	if len(s.markStack) != 0 {
		s.markStack = s.markStack[:len(s.markStack)-1]
	}
}

func (s *renderState) marks() []Mark {
	if len(s.markStack) == 0 {
		return nil
	}
	return normalizeMarks(append([]Mark(nil), s.markStack...))
}

// normalizeMarks enforces ADF's code-mark restriction: code can only be
// combined with a link mark, so other formatting is removed when code is set.
func normalizeMarks(marks []Mark) []Mark {
	hasCode := false
	for _, mark := range marks {
		if mark.Type == "code" {
			hasCode = true
			break
		}
	}
	if !hasCode {
		return marks
	}
	normalized := make([]Mark, 0, 2)
	for _, mark := range marks {
		if mark.Type == "code" || mark.Type == "link" {
			normalized = append(normalized, mark)
		}
	}
	return normalized
}

func (r *Renderer) renderDocument(w io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}
	data, err := json.Marshal(state(rc).document, jsontext.WithIndent("  "))
	if err != nil {
		return ast.WalkStop, err
	}
	n, err := w.Write(data)
	if err != nil {
		return ast.WalkStop, err
	}
	if n != len(data) {
		return ast.WalkStop, io.ErrShortWrite
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	if entering {
		s.pushNode(NewHeading(node.(*ast.Heading).Level))
	} else {
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderBlockquote(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	if entering {
		s.pushNode(NewBlockquote())
	} else {
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.CodeBlock)
	language, _ := n.Language(source)
	code := NewCodeBlock(language)
	if value := n.Value.Str(source); value != "" {
		code.AppendChild(*NewText(value))
	}
	state(rc).appendNode(*code)
	return ast.WalkSkipChildren, nil
}

// ADF has no raw HTML nodes. Blocks are discarded; inline text adjacent to raw
// tags is preserved by renderRawHTML's no-op behavior.
func (r *Renderer) renderHTMLBlock(_ io.Writer, _ []byte, _ ast.Node, _ bool, _ renderer.Context) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderList(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	list := node.(*ast.List)
	if entering {
		if isTaskList(list) {
			if parent, ok := list.Parent().(*ast.ListItem); ok && s.openTaskItems[parent] {
				s.popNode()
				s.openTaskItems[parent] = false
			}
			s.pushNode(NewTaskList())
		} else if list.IsOrdered() {
			s.pushNode(NewOrderedList(list.Start))
		} else {
			s.pushNode(NewBulletList())
		}
	} else {
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	item := node.(*ast.ListItem)
	if list, ok := item.Parent().(*ast.List); ok && isTaskList(list) {
		if entering {
			status, _ := extension.TaskStatusOf(item)
			adfStatus := "TODO"
			if status == extension.TaskStatusCompleted {
				adfStatus = "DONE"
			}
			s.pushNode(NewTaskItem(adfStatus))
			s.openTaskItems[item] = true
		} else if s.openTaskItems[item] {
			s.popNode()
			s.openTaskItems[item] = false
		}
		return ast.WalkContinue, nil
	}
	if entering {
		s.pushNode(NewListItem())
	} else {
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	if isTaskParagraph(node) {
		return ast.WalkContinue, nil
	}
	if entering {
		s.pushNode(NewParagraph())
		if item, ok := node.Parent().(*ast.ListItem); ok && extension.IsTask(item) {
			status, _ := extension.TaskStatusOf(item)
			prefix := "[ ] "
			if status == extension.TaskStatusCompleted {
				prefix = "[x] "
			}
			s.appendNode(*NewText(prefix))
		}
		return ast.WalkContinue, nil
	}
	if s.suppressedParagraph[node] {
		delete(s.suppressedParagraph, node)
		return ast.WalkContinue, nil
	}
	if current := s.currentNode(); current != nil && current.Type == "paragraph" && len(current.Content) == 0 && s.splitParagraph[node] {
		delete(s.splitParagraph, node)
		s.discardCurrentNode()
	} else {
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderThematicBreak(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).appendNode(*NewRule())
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderAutoLink(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.AutoLink)
		state(rc).appendNode(*NewTextWithMarks(n.Label.Value(source), []Mark{NewLinkMark(n.Destination.Value(source), "")}))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeSpan(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		s := state(rc)
		marks := append(s.marks(), NewCodeMark())
		s.appendNode(*NewTextWithMarks(node.(*ast.CodeSpan).Value.Value(source), normalizeMarks(marks)))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushMark(NewEmMark())
	} else {
		state(rc).popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrong(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushMark(NewStrongMark())
	} else {
		state(rc).popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderImage(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	image := Image{Destination: n.Destination.Value(source), Alt: inlineText(n, source), Title: n.Title.Value(source)}
	if image.Alt == "" {
		image.Alt = image.Destination
	}
	s := state(rc)
	if handler := r.Config().ImageHandler; handler != nil {
		result, err := handler(image)
		if err != nil {
			return ast.WalkStop, err
		}
		if result.Node.Type == "" {
			return ast.WalkStop, fmt.Errorf("adf: ImageHandler returned a node with no type")
		}
		switch result.Placement {
		case ImageInline:
			s.appendNode(result.Node)
		case ImageBlock:
			r.appendBlockNode(s, result.Node, node)
		default:
			return ast.WalkStop, fmt.Errorf("adf: ImageHandler returned unknown image placement %d", result.Placement)
		}
		return ast.WalkSkipChildren, nil
	}
	if r.Config().ExternalMedia && (s.currentNode() == nil || s.currentNode().Type != "taskItem") {
		r.appendBlockNode(s, r.externalMediaNode(image), node)
	} else {
		s.appendNode(*NewTextWithMarks(image.Alt, []Mark{NewLinkMark(image.Destination, image.Title)}))
	}
	return ast.WalkSkipChildren, nil
}

func inlineText(node ast.Node, source []byte) string {
	var out strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if text, ok := n.(*ast.Text); ok {
				out.WriteString(text.Value.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return out.String()
}

func (r *Renderer) externalMediaNode(image Image) Node {
	mediaSingle := NewMediaSingle(string(r.Config().ImageLayout))
	mediaSingle.AppendChild(*NewExternalMedia(image.Destination, image.Alt))
	if image.Title != "" {
		caption := NewCaption()
		caption.AppendChild(*NewText(image.Title))
		mediaSingle.AppendChild(*caption)
	}
	return *mediaSingle
}

// appendBlockNode splits a surrounding paragraph so block content is a sibling
// of the text before and after the source image.
func (r *Renderer) appendBlockNode(s *renderState, node Node, sourceNode ast.Node) {
	if current := s.currentNode(); current != nil && current.Type == "paragraph" {
		if paragraph := enclosingParagraph(sourceNode); paragraph != nil {
			s.splitParagraph[paragraph] = true
		}
		if len(current.Content) == 0 {
			s.discardCurrentNode()
		} else {
			s.popNode()
		}
		s.appendNode(node)
		s.pushNode(NewParagraph())
		return
	}
	s.appendNode(node)
}

func enclosingParagraph(node ast.Node) *ast.Paragraph {
	for node != nil {
		if paragraph, ok := node.(*ast.Paragraph); ok {
			return paragraph
		}
		node = node.Parent()
	}
	return nil
}

func (r *Renderer) renderLink(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Link)
		state(rc).pushMark(NewLinkMark(n.Destination.Value(source), n.Title.Value(source)))
	} else {
		state(rc).popMark()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderRawHTML(_ io.Writer, _ []byte, _ ast.Node, _ bool, _ renderer.Context) (ast.WalkStatus, error) {
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	s := state(rc)
	if value := n.Value.Value(source); value != "" {
		if marks := s.marks(); len(marks) != 0 {
			s.appendNode(*NewTextWithMarks(value, marks))
		} else {
			s.appendNode(*NewText(value))
		}
	}
	if n.HardLineBreak() {
		s.appendNode(*NewHardBreak())
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTable(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		table := NewTable()
		table.Attrs["layout"] = string(r.Config().TableLayout)
		state(rc).pushNode(table)
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableHeader(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushNode(NewTableRow())
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

// TableBody is only an AST grouping node; its rows attach directly to the table.
func (r *Renderer) renderTableBody(_ io.Writer, _ []byte, _ ast.Node, _ bool, _ renderer.Context) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableRow(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushNode(NewTableRow())
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableCell(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	s := state(rc)
	if entering {
		if _, header := node.Parent().(*extast.TableHeader); header {
			s.pushNode(NewTableHeader())
		} else {
			s.pushNode(NewTableCell())
		}
		s.pushNode(NewParagraph())
	} else {
		s.popNode()
		s.popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrikethrough(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushMark(NewStrikeMark())
	} else {
		state(rc).popMark()
	}
	return ast.WalkContinue, nil
}

func isTaskList(list *ast.List) bool {
	if list.IsOrdered() || list.FirstChild() == nil {
		return false
	}
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		item, ok := child.(*ast.ListItem)
		if !ok || !extension.IsTask(item) || !isTaskItem(item) {
			return false
		}
	}
	return true
}

func isTaskItem(item *ast.ListItem) bool {
	paragraph, ok := item.FirstChild().(*ast.Paragraph)
	if !ok || containsBlockOnlyCard(paragraph) {
		return false
	}
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Paragraph:
			if child != item.FirstChild() {
				return false
			}
		case *ast.List:
			if !isTaskList(n) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func containsBlockOnlyCard(node ast.Node) bool {
	found := false
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case astext.KindCard, astext.KindEmbed:
				found = true
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
	return found
}

func isTaskParagraph(node ast.Node) bool {
	item, ok := node.Parent().(*ast.ListItem)
	if !ok {
		return false
	}
	list, ok := item.Parent().(*ast.List)
	return ok && isTaskList(list)
}

func (r *Renderer) renderStatus(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		n := node.(*astext.Status)
		state(rc).appendNode(*NewStatusNode(n.Text.Value(source), n.Color))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderMention(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		n := node.(*astext.Mention)
		state(rc).appendNode(*NewMentionNode(n.ID.Value(source), n.DisplayName.Value(source)))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderDate(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).appendNode(*NewDateNode(node.(*astext.Date).Timestamp.Value(source)))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderPlaceholder(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).appendNode(*NewPlaceholderNode(node.(*astext.Placeholder).Label.Value(source)))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCard(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Card)
	s := state(rc)
	if isStandaloneParagraph(node) {
		s.discardCurrentNode()
		s.appendNode(*NewBlockCardNode(n.URL.Value(source)))
		s.suppressedParagraph[node.Parent()] = true
	} else {
		s.appendNode(*NewInlineCardNode(n.URL.Value(source)))
	}
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderEmbed(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*astext.Embed)
	s := state(rc)
	if isStandaloneParagraph(node) {
		s.discardCurrentNode()
		embed := NewEmbedCardNode(n.URL.Value(source))
		embed.Attrs["layout"] = "center"
		s.appendNode(*embed)
		s.suppressedParagraph[node.Parent()] = true
	} else {
		s.appendNode(*NewText("[embed:" + n.URL.Value(source) + "]"))
	}
	return ast.WalkSkipChildren, nil
}

func isStandaloneParagraph(node ast.Node) bool {
	parent, ok := node.Parent().(*ast.Paragraph)
	return ok && parent.FirstChild() == node && parent.LastChild() == node
}

func (r *Renderer) renderEmoji(_ io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).appendNode(*NewEmojiNode(node.(*astext.Emoji).ShortName.Value(source)))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderPanel(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushNode(NewPanelNode(node.(*astext.Panel).PanelType))
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderDecisionList(_ io.Writer, _ []byte, _ ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushNode(NewDecisionListNode())
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderDecisionItem(_ io.Writer, _ []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if entering {
		state(rc).pushNode(NewDecisionItemNode(node.(*astext.DecisionItem).State))
	} else {
		state(rc).popNode()
	}
	return ast.WalkContinue, nil
}

var _ renderer.Renderer[io.Writer] = (*Renderer)(nil)
