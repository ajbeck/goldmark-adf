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

// convertWithOptions is a test helper that converts markdown with options
func convertWithOptions(source []byte, opts ...Option) ([]byte, error) {
	var buf bytes.Buffer
	md := New(opts...)
	if err := md.Convert(source, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestConvert_Image_DefaultFallback(t *testing.T) {
	// By default, images are converted to text with link marks
	input := []byte("![Alt text](https://example.com/image.png)")
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

	// Should have a paragraph with text that has a link mark
	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "paragraph" {
		t.Errorf("Expected paragraph, got %s", doc.Content[0].Type)
	}
	if len(doc.Content[0].Content) != 1 {
		t.Fatalf("Expected 1 child in paragraph, got %d", len(doc.Content[0].Content))
	}
	textNode := doc.Content[0].Content[0]
	if textNode.Type != "text" {
		t.Errorf("Expected text, got %s", textNode.Type)
	}
	if textNode.Text != "Alt text" {
		t.Errorf("Expected text 'Alt text', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "link" {
		t.Errorf("Expected link mark on text")
	}
}

func TestConvert_Image_ExternalMedia_Simple(t *testing.T) {
	input := []byte("![Alt text](https://example.com/image.png)")
	output, err := convertWithOptions(input, WithExternalMedia(true))
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

	// Should have a mediaSingle containing a media node
	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	mediaSingle := doc.Content[0]
	if mediaSingle.Type != "mediaSingle" {
		t.Errorf("Expected mediaSingle, got %s", mediaSingle.Type)
	}
	if mediaSingle.Attrs["layout"] != "center" {
		t.Errorf("Expected layout 'center', got %v", mediaSingle.Attrs["layout"])
	}
	if len(mediaSingle.Content) != 1 {
		t.Fatalf("Expected 1 child in mediaSingle, got %d", len(mediaSingle.Content))
	}
	media := mediaSingle.Content[0]
	if media.Type != "media" {
		t.Errorf("Expected media, got %s", media.Type)
	}
	if media.Attrs["type"] != "external" {
		t.Errorf("Expected type 'external', got %v", media.Attrs["type"])
	}
	if media.Attrs["url"] != "https://example.com/image.png" {
		t.Errorf("Expected url 'https://example.com/image.png', got %v", media.Attrs["url"])
	}
	if media.Attrs["alt"] != "Alt text" {
		t.Errorf("Expected alt 'Alt text', got %v", media.Attrs["alt"])
	}
}

func TestConvert_Image_ExternalMedia_WithCaption(t *testing.T) {
	input := []byte(`![Alt text](https://example.com/image.png "This is a caption")`)
	output, err := convertWithOptions(input, WithExternalMedia(true))
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

	// Should have a mediaSingle with media + caption
	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d", len(doc.Content))
	}
	mediaSingle := doc.Content[0]
	if mediaSingle.Type != "mediaSingle" {
		t.Errorf("Expected mediaSingle, got %s", mediaSingle.Type)
	}
	if len(mediaSingle.Content) != 2 {
		t.Fatalf("Expected 2 children in mediaSingle (media + caption), got %d", len(mediaSingle.Content))
	}

	// Check media
	media := mediaSingle.Content[0]
	if media.Type != "media" {
		t.Errorf("Expected media, got %s", media.Type)
	}

	// Check caption
	caption := mediaSingle.Content[1]
	if caption.Type != "caption" {
		t.Errorf("Expected caption, got %s", caption.Type)
	}
	if len(caption.Content) != 1 {
		t.Fatalf("Expected 1 child in caption, got %d", len(caption.Content))
	}
	if caption.Content[0].Text != "This is a caption" {
		t.Errorf("Expected caption text 'This is a caption', got %s", caption.Content[0].Text)
	}
}

func TestConvert_Image_ExternalMedia_InlineParagraphSplitting(t *testing.T) {
	input := []byte("Before image ![img](https://example.com/image.png) after image")
	output, err := convertWithOptions(input, WithExternalMedia(true))
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

	// Should have: paragraph, mediaSingle, paragraph
	if len(doc.Content) != 3 {
		t.Fatalf("Expected 3 content nodes (para, media, para), got %d\nOutput: %s", len(doc.Content), output)
	}

	// First paragraph with "Before image "
	if doc.Content[0].Type != "paragraph" {
		t.Errorf("Expected first node to be paragraph, got %s", doc.Content[0].Type)
	}
	if len(doc.Content[0].Content) == 0 || doc.Content[0].Content[0].Text != "Before image " {
		t.Errorf("Expected first paragraph to contain 'Before image ', got %v", doc.Content[0].Content)
	}

	// MediaSingle
	if doc.Content[1].Type != "mediaSingle" {
		t.Errorf("Expected second node to be mediaSingle, got %s", doc.Content[1].Type)
	}

	// Second paragraph with " after image"
	if doc.Content[2].Type != "paragraph" {
		t.Errorf("Expected third node to be paragraph, got %s", doc.Content[2].Type)
	}
	if len(doc.Content[2].Content) == 0 || doc.Content[2].Content[0].Text != " after image" {
		t.Errorf("Expected second paragraph to contain ' after image', got %v", doc.Content[2].Content)
	}
}

func TestConvert_Image_ExternalMedia_InList(t *testing.T) {
	input := []byte("- Item with ![image](https://example.com/image.png) inline")
	output, err := convertWithOptions(input, WithExternalMedia(true))
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

	// Should have bulletList -> listItem -> [para, mediaSingle, para]
	if len(doc.Content) != 1 || doc.Content[0].Type != "bulletList" {
		t.Fatalf("Expected bulletList, got %v", doc.Content)
	}
	bulletList := doc.Content[0]
	if len(bulletList.Content) != 1 || bulletList.Content[0].Type != "listItem" {
		t.Fatalf("Expected listItem, got %v", bulletList.Content)
	}
	listItem := bulletList.Content[0]

	// List item should contain: paragraph, mediaSingle, paragraph
	if len(listItem.Content) != 3 {
		t.Fatalf("Expected 3 nodes in listItem, got %d\nOutput: %s", len(listItem.Content), output)
	}
	if listItem.Content[0].Type != "paragraph" {
		t.Errorf("Expected first child to be paragraph, got %s", listItem.Content[0].Type)
	}
	if listItem.Content[1].Type != "mediaSingle" {
		t.Errorf("Expected second child to be mediaSingle, got %s", listItem.Content[1].Type)
	}
	if listItem.Content[2].Type != "paragraph" {
		t.Errorf("Expected third child to be paragraph, got %s", listItem.Content[2].Type)
	}
}

func TestConvert_Image_ExternalMedia_InBlockquote(t *testing.T) {
	input := []byte("> Quote with ![image](https://example.com/image.png)")
	output, err := convertWithOptions(input, WithExternalMedia(true))
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

	// Should have blockquote -> [para, mediaSingle]
	if len(doc.Content) != 1 || doc.Content[0].Type != "blockquote" {
		t.Fatalf("Expected blockquote, got %v", doc.Content)
	}
	blockquote := doc.Content[0]

	// Blockquote should contain: paragraph, mediaSingle
	if len(blockquote.Content) != 2 {
		t.Fatalf("Expected 2 nodes in blockquote, got %d\nOutput: %s", len(blockquote.Content), output)
	}
	if blockquote.Content[0].Type != "paragraph" {
		t.Errorf("Expected first child to be paragraph, got %s", blockquote.Content[0].Type)
	}
	if blockquote.Content[1].Type != "mediaSingle" {
		t.Errorf("Expected second child to be mediaSingle, got %s", blockquote.Content[1].Type)
	}
}

func TestConvert_Image_ExternalMedia_CustomLayout(t *testing.T) {
	input := []byte("![Alt text](https://example.com/image.png)")
	output, err := convertWithOptions(input, WithExternalMedia(true), WithImageLayout("wide"))
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

	mediaSingle := doc.Content[0]
	if mediaSingle.Attrs["layout"] != "wide" {
		t.Errorf("Expected layout 'wide', got %v", mediaSingle.Attrs["layout"])
	}
}

// --- ADF Round-Trip Extension Tests ---

func convertGFMWithOptions(source []byte, opts ...Option) ([]byte, error) {
	var buf bytes.Buffer
	md := NewWithGFM(opts...)
	if err := md.Convert(source, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestConvert_TaskList(t *testing.T) {
	input := []byte("- [x] Done task\n- [ ] Todo task")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d\nOutput: %s", len(doc.Content), output)
	}
	taskList := doc.Content[0]
	if taskList.Type != "taskList" {
		t.Fatalf("Expected taskList, got %s\nOutput: %s", taskList.Type, output)
	}
	if len(taskList.Content) != 2 {
		t.Fatalf("Expected 2 task items, got %d", len(taskList.Content))
	}

	item0 := taskList.Content[0]
	if item0.Type != "taskItem" {
		t.Errorf("Expected taskItem, got %s", item0.Type)
	}
	if item0.Attrs["state"] != "DONE" {
		t.Errorf("Expected state DONE, got %v", item0.Attrs["state"])
	}

	item1 := taskList.Content[1]
	if item1.Type != "taskItem" {
		t.Errorf("Expected taskItem, got %s", item1.Type)
	}
	if item1.Attrs["state"] != "TODO" {
		t.Errorf("Expected state TODO, got %v", item1.Attrs["state"])
	}
}

func TestConvert_Status(t *testing.T) {
	input := []byte("[status:In Progress|yellow]")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d\nOutput: %s", len(doc.Content), output)
	}
	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("Expected 1 inline node, got %d\nOutput: %s", len(para.Content), output)
	}
	status := para.Content[0]
	if status.Type != "status" {
		t.Errorf("Expected status, got %s", status.Type)
	}
	if status.Attrs["text"] != "In Progress" {
		t.Errorf("Expected text 'In Progress', got %v", status.Attrs["text"])
	}
	if status.Attrs["color"] != "yellow" {
		t.Errorf("Expected color 'yellow', got %v", status.Attrs["color"])
	}
}

func TestConvert_StatusWithEscapedPipe(t *testing.T) {
	input := []byte(`[status:Pass \| Fail|red]`)
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	status := doc.Content[0].Content[0]
	if status.Type != "status" {
		t.Fatalf("Expected status, got %s\nOutput: %s", status.Type, output)
	}
	if status.Attrs["text"] != "Pass | Fail" {
		t.Errorf("Expected text 'Pass | Fail', got %v", status.Attrs["text"])
	}
}

func TestConvert_Mention(t *testing.T) {
	input := []byte("@[Brad](abc123)")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	mention := doc.Content[0].Content[0]
	if mention.Type != "mention" {
		t.Fatalf("Expected mention, got %s\nOutput: %s", mention.Type, output)
	}
	if mention.Attrs["id"] != "abc123" {
		t.Errorf("Expected id 'abc123', got %v", mention.Attrs["id"])
	}
	if mention.Attrs["text"] != "Brad" {
		t.Errorf("Expected text 'Brad', got %v", mention.Attrs["text"])
	}
}

func TestConvert_Date(t *testing.T) {
	input := []byte("[date:1582152559]")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	date := doc.Content[0].Content[0]
	if date.Type != "date" {
		t.Fatalf("Expected date, got %s\nOutput: %s", date.Type, output)
	}
	if date.Attrs["timestamp"] != "1582152559" {
		t.Errorf("Expected timestamp '1582152559', got %v", date.Attrs["timestamp"])
	}
}

func TestConvert_Placeholder(t *testing.T) {
	input := []byte("{{name}}")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	ph := doc.Content[0].Content[0]
	if ph.Type != "placeholder" {
		t.Fatalf("Expected placeholder, got %s\nOutput: %s", ph.Type, output)
	}
	if ph.Attrs["text"] != "name" {
		t.Errorf("Expected text 'name', got %v", ph.Attrs["text"])
	}
}

func TestConvert_Card(t *testing.T) {
	input := []byte("[card:https://atlassian.com]")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	card := doc.Content[0].Content[0]
	if card.Type != "inlineCard" {
		t.Fatalf("Expected inlineCard, got %s\nOutput: %s", card.Type, output)
	}
	if card.Attrs["url"] != "https://atlassian.com" {
		t.Errorf("Expected url 'https://atlassian.com', got %v", card.Attrs["url"])
	}
}

func TestConvert_Embed(t *testing.T) {
	input := []byte("[embed:https://youtube.com/watch?v=abc]")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	embed := doc.Content[0].Content[0]
	if embed.Type != "embedCard" {
		t.Fatalf("Expected embedCard, got %s\nOutput: %s", embed.Type, output)
	}
	if embed.Attrs["url"] != "https://youtube.com/watch?v=abc" {
		t.Errorf("Expected url, got %v", embed.Attrs["url"])
	}
}

func TestConvert_Emoji(t *testing.T) {
	input := []byte(":smile:")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	emoji := doc.Content[0].Content[0]
	if emoji.Type != "emoji" {
		t.Fatalf("Expected emoji, got %s\nOutput: %s", emoji.Type, output)
	}
	if emoji.Attrs["shortName"] != ":smile:" {
		t.Errorf("Expected shortName ':smile:', got %v", emoji.Attrs["shortName"])
	}
}

func TestConvert_Panel(t *testing.T) {
	input := []byte("> [!WARNING]\n> Be careful here.")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, output)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d\nOutput: %s", len(doc.Content), output)
	}
	panel := doc.Content[0]
	if panel.Type != "panel" {
		t.Fatalf("Expected panel, got %s\nOutput: %s", panel.Type, output)
	}
	if panel.Attrs["panelType"] != "warning" {
		t.Errorf("Expected panelType 'warning', got %v", panel.Attrs["panelType"])
	}
}

func TestConvert_DecisionList(t *testing.T) {
	input := []byte("- [!] Use json/v2\n- [?] Pending review")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, output)
	}

	if len(doc.Content) != 1 {
		t.Fatalf("Expected 1 content node, got %d\nOutput: %s", len(doc.Content), output)
	}
	decList := doc.Content[0]
	if decList.Type != "decisionList" {
		t.Fatalf("Expected decisionList, got %s\nOutput: %s", decList.Type, output)
	}
	if len(decList.Content) != 2 {
		t.Fatalf("Expected 2 items, got %d\nOutput: %s", len(decList.Content), output)
	}

	item0 := decList.Content[0]
	if item0.Type != "decisionItem" {
		t.Errorf("Expected decisionItem, got %s", item0.Type)
	}
	if item0.Attrs["state"] != "DECIDED" {
		t.Errorf("Expected state DECIDED, got %v", item0.Attrs["state"])
	}

	item1 := decList.Content[1]
	if item1.Attrs["state"] != "UNDECIDED" {
		t.Errorf("Expected state UNDECIDED, got %v", item1.Attrs["state"])
	}
}

func TestConvert_MixedInline(t *testing.T) {
	input := []byte(":smile: @[Brad](abc) [date:1582152559] [status:In Progress|yellow] [card:https://atlassian.com]")
	output, err := convertGFMWithOptions(input)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var doc Document
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, output)
	}

	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Fatalf("Expected paragraph, got %s", para.Type)
	}

	// Check that we got the right node types in order
	expected := []string{"emoji", "text", "mention", "text", "date", "text", "status", "text", "inlineCard"}
	if len(para.Content) != len(expected) {
		t.Fatalf("Expected %d inline nodes, got %d\nOutput: %s", len(expected), len(para.Content), output)
	}
	for i, want := range expected {
		if para.Content[i].Type != want {
			t.Errorf("node[%d]: expected %s, got %s", i, want, para.Content[i].Type)
		}
	}
}
