//go:build goexperiment.jsonv2

// Package adf provides a [goldmark] renderer for Atlassian Document Format (ADF).
//
// This package implements a custom renderer for the goldmark Markdown parser
// that outputs ADF JSON instead of HTML. ADF is the native document format
// used by Atlassian products like Jira Cloud and Confluence Cloud.
//
// # Build Requirements
//
// This package requires Go 1.25+ with the experimental json/v2 package:
//
//	GOEXPERIMENT=jsonv2 go build ./...
//	GOEXPERIMENT=jsonv2 go test ./...
//
// # Basic Usage
//
// Use [New] to create a reusable goldmark instance:
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
//	    adf.WithTableLayout("wide"),
//	    adf.WithImageHandler(customHandler),
//	)
//
// # Schema Validation
//
// The [adfschema] subpackage provides JSON Schema validation for ADF documents:
//
//	import "github.com/ajbeck/goldmark-adf/adfschema"
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
// task lists (rendered as "[x]" or "[ ]" text prefixes).
//
// Raw HTML is skipped as ADF does not support arbitrary HTML content.
//
// [goldmark]: https://github.com/yuin/goldmark
// [adfschema]: https://pkg.go.dev/github.com/ajbeck/goldmark-adf/adfschema
package adf
