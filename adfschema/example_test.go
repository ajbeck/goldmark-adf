package adfschema_test

import (
	"fmt"

	"github.com/ajbeck/goldmark-adf/adfschema"
)

// This example demonstrates validating ADF JSON against the official Atlassian
// Document Format schema.
func ExampleValidate() {
	// Valid ADF document
	validADF := []byte(`{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Hello, world!"
					}
				]
			}
		]
	}`)

	err := adfschema.Validate(validADF)
	if err != nil {
		fmt.Println("Invalid:", err)
	} else {
		fmt.Println("Valid ADF document")
	}
	// Output:
	// Valid ADF document
}

// This example demonstrates how [Validate] reports errors for invalid ADF.
func ExampleValidate_invalid() {
	// Invalid: missing required "version" field
	invalidADF := []byte(`{
		"type": "doc",
		"content": []
	}`)

	err := adfschema.Validate(invalidADF)
	if err != nil {
		fmt.Println("Validation failed")
	}
	// Output:
	// Validation failed
}
