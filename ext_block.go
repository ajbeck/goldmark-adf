package adf

import (
	"bytes"
	"regexp"

	"github.com/ajbeck/goldmark-adf/v2/astext"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

type panelTransformer struct{}

var alertPattern = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]`)

var alertToPanelType = map[string]string{
	"NOTE":      "info",
	"TIP":       "success",
	"IMPORTANT": "info",
	"WARNING":   "warning",
	"CAUTION":   "error",
}

func (*panelTransformer) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindBlockquote {
			return ast.WalkContinue, nil
		}
		blockquote := n.(*ast.Blockquote)
		paragraph, ok := blockquote.FirstChild().(*ast.Paragraph)
		if !ok || len(paragraph.Source()) == 0 {
			return ast.WalkContinue, nil
		}
		match := alertPattern.FindSubmatch(paragraph.Source()[0].Bytes(source))
		if match == nil {
			return ast.WalkContinue, nil
		}

		panel := astext.NewPanel(alertToPanelType[string(match[1])])
		panel.SetPos(blockquote.Pos())
		removeLeadingText(paragraph, len(match[0]), source)

		if paragraph.ChildCount() == 0 {
			// ADF panels require at least one child. Preserve a marker-only alert
			// as an intentionally empty paragraph.
			blockquote.RemoveChild(paragraph)
			panel.AppendChild(ast.NewParagraph())
		}
		for child := blockquote.FirstChild(); child != nil; {
			next := child.NextSibling()
			blockquote.RemoveChild(child)
			panel.AppendChild(child)
			child = next
		}
		blockquote.Parent().ReplaceChild(blockquote, panel)
		return ast.WalkSkipChildren, nil
	})
}

// removeLeadingText removes a raw prefix from the first contiguous text nodes
// in a paragraph. Panel and decision prefixes are parsed before normal inline
// content, so retaining source-backed text values is sufficient here.
func removeLeadingText(paragraph ast.Node, remaining int, source []byte) {
	for remaining > 0 {
		child, ok := paragraph.FirstChild().(*ast.Text)
		if !ok {
			return
		}
		value := child.Value
		index := value.Index()
		length := index.Stop - index.Start
		if length <= remaining {
			remaining -= length
			paragraph.RemoveChild(child)
			continue
		}
		child.Value = text.NewSingleLineValueFromIndex(
			text.NewIndex(index.Start+remaining, index.Stop), text.NewDecoder(),
		)
		remaining = 0
	}

	if child, ok := paragraph.FirstChild().(*ast.Text); ok {
		index := child.Value.Index()
		trimmed := bytes.TrimLeft(child.Value.Bytes(source), " ")
		if len(trimmed) < index.Stop-index.Start {
			child.Value = text.NewSingleLineValueFromIndex(
				text.NewIndex(index.Stop-len(trimmed), index.Stop), text.NewDecoder(),
			)
		}
		if child.Value.IsEmpty() {
			paragraph.RemoveChild(child)
		}
	}
}

type decisionTransformer struct{}

var decisionPattern = regexp.MustCompile(`^\[([!?])\]\s*`)

func (*decisionTransformer) Transform(node *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindList {
			return ast.WalkContinue, nil
		}
		list := n.(*ast.List)
		if list.IsOrdered() || !isDecisionList(list, source) {
			return ast.WalkContinue, nil
		}

		decisionList := astext.NewDecisionList()
		decisionList.SetPos(list.Pos())
		for item := list.FirstChild(); item != nil; {
			next := item.NextSibling()
			paragraph := item.FirstChild().(*ast.Paragraph)
			state := decisionState(paragraph, source)
			removeLeadingText(paragraph, len(decisionPattern.Find(blockLeadingText(paragraph, source))), source)

			decisionItem := astext.NewDecisionItem(state)
			decisionItem.SetPos(item.Pos())
			for child := paragraph.FirstChild(); child != nil; {
				nextChild := child.NextSibling()
				paragraph.RemoveChild(child)
				decisionItem.AppendChild(child)
				child = nextChild
			}
			list.RemoveChild(item)
			decisionList.AppendChild(decisionItem)
			item = next
		}
		list.Parent().ReplaceChild(list, decisionList)
		return ast.WalkSkipChildren, nil
	})
}

func isDecisionList(list *ast.List, source []byte) bool {
	if list.FirstChild() == nil {
		return false
	}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		if item.Kind() != ast.KindListItem || item.ChildCount() != 1 {
			return false
		}
		paragraph, ok := item.FirstChild().(*ast.Paragraph)
		if !ok || decisionPattern.Find(blockLeadingText(paragraph, source)) == nil {
			return false
		}
	}
	return true
}

func blockLeadingText(node ast.Node, source []byte) []byte {
	var output []byte
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		textNode, ok := child.(*ast.Text)
		if !ok {
			break
		}
		output = append(output, textNode.Value.Bytes(source)...)
	}
	return output
}

func decisionState(paragraph *ast.Paragraph, source []byte) string {
	match := decisionPattern.FindSubmatch(blockLeadingText(paragraph, source))
	if len(match) == 2 && match[1][0] == '!' {
		return "DECIDED"
	}
	return "UNDECIDED"
}
