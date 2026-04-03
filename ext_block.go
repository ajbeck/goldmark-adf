//go:build goexperiment.jsonv2

package adf

import (
	"bytes"
	"regexp"

	"github.com/ajbeck/goldmark-adf/astext"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// panelTransformer converts blockquotes that start with [!TYPE] (GitHub alert
// syntax) into Panel AST nodes.
type panelTransformer struct{}

var alertPattern = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]`)

var alertToPanelType = map[string]string{
	"NOTE":      "info",
	"TIP":       "success",
	"IMPORTANT": "info",
	"WARNING":   "warning",
	"CAUTION":   "error",
}

func (t *panelTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	// Collect all alert blockquotes first. Goldmark's ast.Walk advances
	// via c.NextSibling() after visiting each child. ReplaceChild sets
	// the old node's NextSibling to nil, which terminates the walk and
	// causes subsequent blockquotes to be skipped. Collecting first and
	// transforming second avoids mutating the tree during traversal.
	var targets []*ast.Blockquote
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindBlockquote {
			return ast.WalkContinue, nil
		}
		bq := n.(*ast.Blockquote)
		first := bq.FirstChild()
		if first == nil || !ast.IsParagraph(first) {
			return ast.WalkContinue, nil
		}
		para := first.(*ast.Paragraph)
		if para.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}
		firstSeg := para.Lines().At(0)
		if alertPattern.Match(firstSeg.Value(source)) {
			targets = append(targets, bq)
		}
		return ast.WalkContinue, nil
	})

	for _, bq := range targets {
		transformBlockquoteToPanel(bq, source)
	}
}

// transformBlockquoteToPanel converts a single alert blockquote into a
// Panel node and splices it into the tree.
func transformBlockquoteToPanel(bq *ast.Blockquote, source []byte) {
	first := bq.FirstChild()
	if first == nil || !ast.IsParagraph(first) {
		return
	}
	para := first.(*ast.Paragraph)
	if para.Lines().Len() == 0 {
		return
	}
	firstSeg := para.Lines().At(0)
	firstLine := firstSeg.Value(source)
	m := alertPattern.FindSubmatch(firstLine)
	if m == nil {
		return
	}

	keyword := string(m[1])
	panelType := alertToPanelType[keyword]
	prefixLen := len(m[0])

	panel := astext.NewPanel(panelType)

	// Strip the [!TYPE] prefix from the paragraph's inline children.
	// Because inline parsing may have split the prefix across multiple
	// text nodes, walk forward and consume bytes until we've skipped
	// past the prefix.
	remaining := prefixLen
	for remaining > 0 {
		child := para.FirstChild()
		if child == nil {
			break
		}
		if child.Kind() != ast.KindText {
			break
		}
		tn := child.(*ast.Text)
		seg := tn.Segment
		nodeLen := seg.Stop - seg.Start
		if nodeLen <= remaining {
			remaining -= nodeLen
			para.RemoveChild(para, child)
		} else {
			tn.Segment = text.NewSegment(seg.Start+remaining, seg.Stop)
			remaining = 0
		}
	}
	// Also trim any leading space from the next text node.
	if fc := para.FirstChild(); fc != nil && fc.Kind() == ast.KindText {
		tn := fc.(*ast.Text)
		seg := tn.Segment
		val := seg.Value(source)
		trimmed := bytes.TrimLeft(val, " ")
		if len(trimmed) < len(val) {
			tn.Segment = text.NewSegment(seg.Start+(len(val)-len(trimmed)), seg.Stop)
		}
		if tn.Segment.Stop == tn.Segment.Start {
			para.RemoveChild(para, fc)
		}
	}
	// If the paragraph is now empty, remove it.
	if para.ChildCount() == 0 {
		bq.RemoveChild(bq, para)
	}

	// Move all children from blockquote to panel.
	for c := bq.FirstChild(); c != nil; {
		next := c.NextSibling()
		bq.RemoveChild(bq, c)
		panel.AppendChild(panel, c)
		c = next
	}

	// Replace blockquote with panel in the tree.
	parent := bq.Parent()
	parent.ReplaceChild(parent, bq, panel)
}

// decisionTransformer converts list items starting with [!] or [?] into
// DecisionList/DecisionItem AST nodes.
type decisionTransformer struct{}

func (t *decisionTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindList {
			return ast.WalkContinue, nil
		}

		list := n.(*ast.List)
		if list.IsOrdered() {
			return ast.WalkContinue, nil
		}

		// Check if ALL items are decision items
		allDecision := true
		list.FirstChild() // just checking
		for c := list.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Kind() != ast.KindListItem {
				allDecision = false
				break
			}
			if !isDecisionItem(c, source) {
				allDecision = false
				break
			}
		}
		if !allDecision {
			return ast.WalkContinue, nil
		}

		// Convert to DecisionList
		decList := astext.NewDecisionList()
		for c := list.FirstChild(); c != nil; {
			next := c.NextSibling()
			state := extractDecisionState(c, source)
			removeDecisionPrefix(c, source)

			decItem := astext.NewDecisionItem(state)
			// Move list item's children to decision item
			for gc := c.FirstChild(); gc != nil; {
				gnext := gc.NextSibling()
				c.RemoveChild(c, gc)
				decItem.AppendChild(decItem, gc)
				gc = gnext
			}
			list.RemoveChild(list, c)
			decList.AppendChild(decList, decItem)
			c = next
		}

		parent := list.Parent()
		parent.ReplaceChild(parent, list, decList)

		return ast.WalkContinue, nil
	})
}

var decisionPattern = regexp.MustCompile(`^\[([!?])\]\s*`)

// paraRawLine returns the first raw source line of a paragraph, using
// Lines() which is unaffected by inline parsing. Returns nil if unavailable.
// blockLeadingText collects text content from the leading text nodes of a
// paragraph or text block. This is needed because inline parsing may split
// bracket-based syntax across multiple text nodes.
// Handles both Paragraph (loose lists, blockquotes) and TextBlock (tight lists).
func blockLeadingText(n ast.Node, source []byte) []byte {
	if !ast.IsParagraph(n) && n.Kind() != ast.KindTextBlock {
		return nil
	}
	// Concatenate leading text node values.
	var buf []byte
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() != ast.KindText {
			break
		}
		buf = append(buf, c.(*ast.Text).Segment.Value(source)...)
	}
	return buf
}

func isDecisionItem(n ast.Node, source []byte) bool {
	first := n.FirstChild()
	if first == nil {
		return false
	}
	line := blockLeadingText(first, source)
	return line != nil && decisionPattern.Match(line)
}

func extractDecisionState(n ast.Node, source []byte) string {
	first := n.FirstChild()
	line := blockLeadingText(first, source)
	if line == nil {
		return ""
	}
	m := decisionPattern.FindSubmatch(line)
	if m == nil {
		return ""
	}
	if string(m[1]) == "!" {
		return "DECIDED"
	}
	return "UNDECIDED"
}

func removeDecisionPrefix(n ast.Node, source []byte) {
	first := n.FirstChild()
	if first == nil {
		return
	}
	if !ast.IsParagraph(first) && first.Kind() != ast.KindTextBlock {
		return
	}
	para := first // use ast.Node interface
	line := blockLeadingText(first, source)
	if line == nil {
		return
	}
	m := decisionPattern.FindSubmatch(line)
	if m == nil {
		return
	}
	prefixLen := len(m[0])

	// Walk inline children and consume bytes corresponding to the prefix.
	remaining := prefixLen
	for remaining > 0 {
		child := para.FirstChild()
		if child == nil {
			break
		}
		if child.Kind() != ast.KindText {
			break
		}
		tn := child.(*ast.Text)
		seg := tn.Segment
		nodeLen := seg.Stop - seg.Start
		if nodeLen <= remaining {
			remaining -= nodeLen
			para.RemoveChild(para, child)
		} else {
			tn.Segment = text.NewSegment(seg.Start+remaining, seg.Stop)
			remaining = 0
		}
	}
}
