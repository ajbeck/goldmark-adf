// Package adf provides a [goldmark] renderer for Atlassian Document Format (ADF).
//
// This package implements a custom renderer for the goldmark Markdown parser
// that outputs ADF JSON instead of HTML. ADF is the native document format
// used by Atlassian products like Jira Cloud and Confluence Cloud.
//
// # Build Requirements
//
// This package requires Go 1.27 or later.
//
// # Basic Usage
//
// Use [New] to create a reusable converter:
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
// Use [NewWithGFM] to enable GitHub Flavored Markdown extensions including
// tables, strikethrough, autolinks, and task lists:
//
//	md := adf.NewWithGFM()
//
// # Convenience Functions
//
// For simple one-off conversions, use [Convert] or [ConvertWithGFM]:
//
//	output, err := adf.Convert([]byte("# Hello"))
//	output, err := adf.ConvertWithGFM([]byte("| A | B |\n|---|---|"))
//
// # Configuration
//
// The renderer can be configured using functional options:
//
//	md := adf.New(
//	    adf.WithTableLayout(adf.TableLayoutWide),
//	    adf.WithImageHandler(customHandler),
//	)
//
// # Schema Validation
//
// The [adfschema] subpackage provides JSON Schema validation for ADF documents:
//
//	import "github.com/ajbeck/goldmark-adf/v2/adfschema"
//
//	if err := adfschema.Validate(jsonBytes); err != nil {
//	    log.Printf("Invalid ADF: %v", err)
//	}
//
// # Supported Markdown Features
//
// Block elements: headings, paragraphs, blockquotes, code blocks (fenced and
// indented), bullet lists, ordered lists, horizontal rules, and hard breaks.
//
// Inline elements: bold, italic, inline code, links, and autolinks. Images are
// converted to links with the image URL.
//
// GFM extensions (with [NewWithGFM]): tables, strikethrough, autolinks, and
// task lists. Fully representable unordered task lists are emitted as ADF
// task lists; ordered or mixed lists retain their textual checkbox markers.
//
// Raw HTML blocks and inline tags are skipped as ADF does not support arbitrary
// HTML content; text surrounding inline tags is preserved.
//
// [goldmark]: https://github.com/yuin/goldmark
// [adfschema]: https://pkg.go.dev/github.com/ajbeck/goldmark-adf/v2/adfschema
package adf
