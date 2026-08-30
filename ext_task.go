package adf

import (
	"regexp"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// safeGFMParser uses Goldmark's GFM parsers except for task checkboxes. In
// Goldmark v2.0.0 the task parser returns the package-global parser.Nil,
// which is mutated while parsing and races when parses run concurrently. The
// replacement below records the identical public task status and returns a
// fresh empty node instead.
type safeGFMParser struct{}

func newSafeGFMParser() parser.Extension { return &safeGFMParser{} }

func (*safeGFMParser) ParserOptions(c *parser.Config) []parser.Option {
	var options []parser.Option
	options = append(options, extension.NewLinkifyParser().ParserOptions(c)...)
	options = append(options, extension.NewTableParser().ParserOptions(c)...)
	options = append(options, extension.NewStrikethroughParser().ParserOptions(c)...)
	options = append(options, parser.WithInlineParsers(util.Prioritized[parser.InlineParser](&taskItemParser{}, 0)))
	return options
}

const taskStatusAttribute = "task-status"

var taskCheckboxPattern = regexp.MustCompile(`^\[([\sxX])\]\s*`)

type taskItemParser struct{}

func (*taskItemParser) Trigger() []byte { return []byte{'['} }

func (*taskItemParser) Parse(parent ast.Node, block text.Reader, _ parser.Context) ast.Node {
	if parent.Parent() == nil || parent.Parent().FirstChild() != parent || parent.HasChildren() {
		return nil
	}
	item, ok := parent.Parent().(*ast.ListItem)
	if !ok {
		return nil
	}
	if _, found := item.Attribute(taskStatusAttribute); found {
		return nil
	}
	line, _ := block.PeekLine()
	match := taskCheckboxPattern.FindSubmatchIndex(line)
	if match == nil {
		return nil
	}

	status := extension.TaskStatusActive
	if line[match[2]] == 'x' || line[match[2]] == 'X' {
		status = extension.TaskStatusCompleted
	}
	item.SetAttribute(taskStatusAttribute, text.NewMultiLineValueFromString(string(status), text.IdentityDecoder))
	block.Advance(match[1])

	// This must not be parser.Nil: Parse assigns a source position before it
	// compares against that global sentinel. A fresh empty text node is ignored
	// by the renderer and is safe for concurrent parses.
	return ast.NewText(text.NewSingleLineValueFromString("", text.IdentityDecoder))
}
