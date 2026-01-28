//go:build goexperiment.jsonv2

// Package adf provides a goldmark renderer for Atlassian Document Format.
//
// This package implements a custom renderer for the goldmark Markdown parser
// that outputs ADF (Atlassian Document Format) JSON instead of HTML.
// ADF is the native document format used by Atlassian products like Jira Cloud
// and Confluence Cloud.
//
// # Basic Usage
//
//	import (
//	    "bytes"
//	    "github.com/ajbeck/goldmark-adf"
//	    "github.com/yuin/goldmark"
//	)
//
//	md := adf.New()
//	var buf bytes.Buffer
//	if err := md.Convert([]byte("# Hello World"), &buf); err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(buf.String())
//
// # With GFM Extensions
//
//	md := adf.New(adf.WithGFM())
//
// # Build Requirements
//
// This package requires Go 1.25+ with the experimental json/v2 package enabled:
//
//	GOEXPERIMENT=jsonv2 go build ./...
//	GOEXPERIMENT=jsonv2 go test ./...
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
