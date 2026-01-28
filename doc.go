//go:build goexperiment.jsonv2

// Package adf provides a goldmark renderer for Atlassian Document Format.
//
// This package implements a custom renderer for the goldmark Markdown parser
// that outputs ADF (Atlassian Document Format) JSON instead of HTML.
// ADF is the native document format used by Atlassian products like Jira Cloud
// and Confluence Cloud.
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
//	md := adf.New()
//	var buf bytes.Buffer
//	if err := md.Convert([]byte("# Hello World"), &buf); err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(buf.String())
//
// # With GFM Extensions
//
//	md := adf.NewWithGFM()
//
// # Convenience Functions
//
//	output, err := adf.Convert([]byte("# Hello"))
//	output, err := adf.ConvertWithGFM([]byte("| A | B |\n|---|---|"))
package adf
