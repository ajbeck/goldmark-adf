package adfschema

import (
	"testing"
)

func TestValidate_MinimalDocument(t *testing.T) {
	// Minimal valid ADF document
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": []
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("minimal document should be valid: %v", err)
	}
}

func TestValidate_ParagraphDocument(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Hello world"
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("paragraph document should be valid: %v", err)
	}
}

func TestValidate_HeadingDocument(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "heading",
				"attrs": {
					"level": 1
				},
				"content": [
					{
						"type": "text",
						"text": "Title"
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("heading document should be valid: %v", err)
	}
}

func TestValidate_InvalidDocument(t *testing.T) {
	// Missing required version field
	doc := []byte(`{
		"type": "doc",
		"content": []
	}`)

	if err := Validate(doc); err == nil {
		t.Error("document without version should be invalid")
	}
}

func TestValidate_InvalidDocType(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "invalid",
		"content": []
	}`)

	if err := Validate(doc); err == nil {
		t.Error("document with invalid type should be invalid")
	}
}

func TestValidate_TextWithMarks(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Bold text",
						"marks": [
							{"type": "strong"}
						]
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("text with marks should be valid: %v", err)
	}
}

func TestValidate_CodeBlock(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "codeBlock",
				"attrs": {
					"language": "go"
				},
				"content": [
					{
						"type": "text",
						"text": "func main() {}"
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("code block should be valid: %v", err)
	}
}

func TestValidate_BulletList(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "bulletList",
				"content": [
					{
						"type": "listItem",
						"content": [
							{
								"type": "paragraph",
								"content": [
									{"type": "text", "text": "Item 1"}
								]
							}
						]
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("bullet list should be valid: %v", err)
	}
}

func TestValidate_Table(t *testing.T) {
	doc := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "table",
				"attrs": {
					"isNumberColumnEnabled": false,
					"layout": "default"
				},
				"content": [
					{
						"type": "tableRow",
						"content": [
							{
								"type": "tableHeader",
								"attrs": {},
								"content": [
									{
										"type": "paragraph",
										"content": [
											{"type": "text", "text": "Header"}
										]
									}
								]
							}
						]
					},
					{
						"type": "tableRow",
						"content": [
							{
								"type": "tableCell",
								"attrs": {},
								"content": [
									{
										"type": "paragraph",
										"content": [
											{"type": "text", "text": "Cell"}
										]
									}
								]
							}
						]
					}
				]
			}
		]
	}`)

	if err := Validate(doc); err != nil {
		t.Errorf("table should be valid: %v", err)
	}
}
