//go:build goexperiment.jsonv2

package adf_test

import (
	"bytes"
	"fmt"

	"github.com/ajbeck/goldmark-adf"
)

func Example_basic() {
	markdown := []byte(`# Hello World

This is a **bold** and *italic* paragraph with a [link](https://example.com).

- Item 1
- Item 2
- Item 3

> This is a blockquote

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```")

	output, err := adf.Convert(markdown)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))
}

func Example_withGFM() {
	markdown := []byte(`| Name | Age |
| ---- | --- |
| Alice | 30 |
| Bob | 25 |

This has ~~strikethrough~~ text.`)

	var buf bytes.Buffer
	md := adf.NewWithGFM()
	if err := md.Convert(markdown, &buf); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(buf.String())
}
