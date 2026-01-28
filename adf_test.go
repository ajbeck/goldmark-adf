//go:build goexperiment.jsonv2

package adf

import (
	"bytes"
	"encoding/json/v2"
	"testing"

	"github.com/ajbeck/goldmark-adf/adfschema"
)

func TestConvert_Paragraph(t *testing.T) {
	input := []byte("Hello world")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Validate against schema
	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	// Parse and check structure
	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}
	if doc.Type != "doc" {
		t.Errorf("Expected type 'doc', got %s", doc.Type)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "paragraph" {
		t.Errorf("Expected paragraph, got %s", doc.Content[0].Type)
	}
}

func TestConvert_Heading(t *testing.T) {
	tests := []struct {
		input string
		level int
	}{
		{"# Heading 1", 1},
		{"## Heading 2", 2},
		{"### Heading 3", 3},
		{"#### Heading 4", 4},
		{"##### Heading 5", 5},
		{"###### Heading 6", 6},
	}

	for _, tc := range tests {
		output, err := Convert([]byte(tc.input))
		if err != nil {
			t.Fatalf("Convert failed for %q: %v", tc.input, err)
		}

		if err := adfschema.Validate(output); err != nil {
			t.Errorf("Invalid ADF for %q: %v", tc.input, err)
		}

		var doc Document
		if err := json.Unmarshal(output, &doc); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if len(doc.Content) != 1 {
			t.Fatalf("Expected 1 content node for %q, got %d", tc.input, len(doc.Content))
		}
		if doc.Content[0].Type != "heading" {
			t.Errorf("Expected heading for %q, got %s", tc.input, doc.Content[0].Type)
		}
		if doc.Content[0].Attrs["level"] != float64(tc.level) {
			t.Errorf("Expected level %d for %q, got %v", tc.level, tc.input, doc.Content[0].Attrs["level"])
		}
	}
}

func TestConvert_Blockquote(t *testing.T) {
	input := []byte("> This is a quote")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "blockquote" {
		t.Errorf("Expected blockquote, got %s", doc.Content[0].Type)
	}
}

func TestConvert_CodeBlock(t *testing.T) {
	input := []byte("```go\nfunc main() {}\n```")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "codeBlock" {
		t.Errorf("Expected codeBlock, got %s", doc.Content[0].Type)
	}
	if doc.Content[0].Attrs["language"] != "go" {
		t.Errorf("Expected language 'go', got %v", doc.Content[0].Attrs["language"])
	}
}

func TestConvert_BulletList(t *testing.T) {
	input := []byte("- Item 1\n- Item 2\n- Item 3")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "bulletList" {
		t.Errorf("Expected bulletList, got %s", doc.Content[0].Type)
	}
	if len(doc.Content[0].Content) != 3 {
		t.Errorf("Expected 3 list items, got %d", len(doc.Content[0].Content))
	}
}

func TestConvert_OrderedList(t *testing.T) {
	input := []byte("1. First\n2. Second\n3. Third")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "orderedList" {
		t.Errorf("Expected orderedList, got %s", doc.Content[0].Type)
	}
}

func TestConvert_ThematicBreak(t *testing.T) {
	input := []byte("Above\n\n---\n\nBelow")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	// Should have: paragraph, rule, paragraph
	foundRule := false
	for _, node := range doc.Content {
		if node.Type == "rule" {
			foundRule = true
			break
		}
	}
	if !foundRule {
		t.Error("Expected to find a rule node")
	}
}

func TestConvert_Emphasis(t *testing.T) {
	input := []byte("*italic* and **bold**")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}
}

func TestConvert_Link(t *testing.T) {
	input := []byte("[click here](https://example.com)")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}
}

func TestConvert_InlineCode(t *testing.T) {
	input := []byte("Use `fmt.Println` for output")
	output, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}
}

func TestConvertWithGFM_Table(t *testing.T) {
	input := []byte(`| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |`)

	output, err := ConvertWithGFM(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "table" {
		t.Errorf("Expected table, got %s", doc.Content[0].Type)
	}
}

func TestConvertWithGFM_Strikethrough(t *testing.T) {
	input := []byte("~~deleted~~")
	output, err := ConvertWithGFM(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if err := adfschema.Validate(output); err != nil {
		t.Errorf("Invalid ADF output: %v\nOutput: %s", err, output)
	}
}

func TestNew_ReusableInstance(t *testing.T) {
	md := New()

	inputs := []string{
		"# Hello",
		"Paragraph",
		"- List item",
	}

	for _, input := range inputs {
		var buf bytes.Buffer
		if err := md.Convert([]byte(input), &buf); err != nil {
			t.Errorf("Convert failed for %q: %v", input, err)
		}
		if err := adfschema.Validate(buf.Bytes()); err != nil {
			t.Errorf("Invalid ADF for %q: %v", input, err)
		}
	}
}
