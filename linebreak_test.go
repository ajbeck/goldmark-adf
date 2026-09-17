//go:build goexperiment.jsonv2

package adf

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/ajbeck/goldmark-adf/adfschema"
)

func TestConvert_LineBreaks(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       string
		wantStrong string
	}{
		{name: "soft break", input: "Hello\nThere", want: "Hello There"},
		{name: "CRLF soft break", input: "Hello\r\nThere", want: "Hello There"},
		{name: "multiple soft breaks", input: "Hello\nThere\nFriend", want: "Hello There Friend"},
		{name: "space before soft break", input: "Hello \nThere", want: "Hello There"},
		{name: "soft break within strong", input: "**Hello\nThere**", want: "Hello There", wantStrong: "Hello There"},
		{name: "soft break after strong", input: "**Hello**\nThere", want: "Hello There", wantStrong: "Hello"},
		{name: "soft break before strong", input: "Hello\n**There**", want: "Hello There", wantStrong: "There"},
		{name: "adjacent formatting", input: "Hello**There**", want: "HelloThere", wantStrong: "There"},
		{name: "soft break after code", input: "`Hello`\nThere", want: "Hello There"},
		{name: "soft break after link", input: "[Hello](https://example.com)\nThere", want: "Hello There"},
		{name: "two space hard break", input: "Hello  \nThere", want: "Hello\nThere"},
		{name: "backslash hard break", input: "Hello\\\nThere", want: "Hello\nThere"},
		{name: "separate paragraphs", input: "Hello\n\nThere", want: "Hello\n\nThere"},
		{name: "trailing newline", input: "Hello\n", want: "Hello"},
	}
	for _, converter := range []struct {
		name    string
		convert func([]byte) ([]byte, error)
	}{
		{"Convert", Convert},
		{"ConvertWithGFM", ConvertWithGFM},
	} {
		t.Run(converter.name, func(t *testing.T) {
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					output, err := converter.convert([]byte(tc.input))
					if err != nil {
						t.Fatal(err)
					}
					if err := adfschema.Validate(output); err != nil {
						t.Fatalf("invalid ADF: %v\n%s", err, output)
					}
					var doc Document
					if err := json.Unmarshal(output, &doc); err != nil {
						t.Fatal(err)
					}
					var text, strong strings.Builder
					for i, block := range doc.Content {
						if block.Type != "paragraph" {
							t.Fatalf("expected paragraph, got %s", block.Type)
						}
						if i > 0 {
							text.WriteString("\n\n")
						}
						for _, node := range block.Content {
							switch node.Type {
							case "text":
								text.WriteString(node.Text)
								for _, mark := range node.Marks {
									if mark.Type == "strong" {
										strong.WriteString(node.Text)
									}
								}
							case "hardBreak":
								text.WriteByte('\n')
							default:
								t.Fatalf("unexpected inline node %s", node.Type)
							}
						}
					}
					if got := text.String(); got != tc.want {
						t.Errorf("text = %q, want %q", got, tc.want)
					}
					if got := strong.String(); got != tc.wantStrong {
						t.Errorf("strong text = %q, want %q", got, tc.wantStrong)
					}
				})
			}
		})
	}
}
