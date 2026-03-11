//go:build goexperiment.jsonv2

package adf

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// New creates a new goldmark.Markdown instance configured to output ADF JSON.
func New(opts ...Option) goldmark.Markdown {
	r := NewRenderer(opts...)
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(r, 1000),
				),
			),
		),
	)
	return md
}

// NewWithGFM creates a new goldmark.Markdown instance with GFM extensions enabled.
// This enables parsing of tables, strikethrough, autolinks, and task lists.
func NewWithGFM(opts ...Option) goldmark.Markdown {
	r := NewRenderer(opts...)

	// Create a custom renderer that ONLY uses our ADF renderer
	// We don't include any HTML renderers
	adfRenderer := renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(r, 1000),
		),
	)

	md := goldmark.New(
		goldmark.WithRenderer(adfRenderer),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	// Manually add only the PARSER parts of GFM extensions
	// (not their HTML renderers)
	addGFMParsers(md)

	// Add ADF round-trip extension parsers
	addADFExtensions(md)

	return md
}

// addGFMParsers adds GFM parser extensions without their HTML renderers.
func addGFMParsers(md goldmark.Markdown) {
	// Table parser
	md.Parser().AddOptions(
		parser.WithParagraphTransformers(
			util.Prioritized(extension.NewTableParagraphTransformer(), 200),
		),
	)

	// Strikethrough parser
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(extension.NewStrikethroughParser(), 500),
		),
	)

	// Linkify parser (autolinks)
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(extension.NewLinkifyParser(), 999),
		),
	)

	// Task list parser
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(extension.NewTaskCheckBoxParser(), 0),
		),
	)
}

// addADFExtensions adds ADF round-trip parser extensions.
func addADFExtensions(md goldmark.Markdown) {
	// Inline parsers for custom ADF syntax.
	// Priority must be higher (lower number) than default link parser (200)
	// so our [status:...], [card:...], [date:...], [embed:...] are tried before
	// goldmark interprets them as link references.
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&statusParser{}, 90),
			util.Prioritized(&mentionParser{}, 91),
			util.Prioritized(&dateParser{}, 92),
			util.Prioritized(&placeholderParser{}, 93),
			util.Prioritized(&cardParser{}, 94),
			util.Prioritized(&embedParser{}, 95),
			util.Prioritized(&emojiParser{}, 96),
		),
	)

	// AST transformers for block-level conversions.
	md.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&panelTransformer{}, 100),
			util.Prioritized(&decisionTransformer{}, 101),
		),
	)
}

// Convert is a convenience function that converts Markdown to ADF JSON.
// It creates a new goldmark instance for each call, which is suitable for
// simple use cases. For better performance with multiple conversions,
// create a goldmark instance with New() and reuse it.
func Convert(source []byte) ([]byte, error) {
	var buf = make([]byte, 0, len(source)*2)
	w := &bytesWriter{buf: buf}
	if err := New().Convert(source, w); err != nil {
		return nil, err
	}
	return w.buf, nil
}

// ConvertWithGFM is like Convert but with GFM extensions enabled.
func ConvertWithGFM(source []byte) ([]byte, error) {
	var buf = make([]byte, 0, len(source)*2)
	w := &bytesWriter{buf: buf}
	if err := NewWithGFM().Convert(source, w); err != nil {
		return nil, err
	}
	return w.buf, nil
}

// bytesWriter is a simple io.Writer that appends to a byte slice.
type bytesWriter struct {
	buf []byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}
