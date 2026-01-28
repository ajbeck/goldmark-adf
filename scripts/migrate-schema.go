//go:build ignore

// migrate-schema.go transforms the ADF JSON Schema from draft-04 to draft-07 format.
//
// Usage:
//
//	go run scripts/migrate-schema.go
//
// This downloads the original schema from Atlassian, applies the necessary
// transformations for draft-07 compatibility, and saves it to adfschema/adf-schema.json.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	schemaURL    = "https://unpkg.com/@atlaskit/adf-schema@51.5.6/dist/json-schema/v1/full.json"
	outputPath   = "adfschema/adf-schema.json"
	draft07URI   = "http://json-schema.org/draft-07/schema#"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Download the schema
	fmt.Println("Downloading ADF schema...")
	resp, err := http.Get(schemaURL)
	if err != nil {
		return fmt.Errorf("downloading schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Parse the schema
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	// Migrate to draft-07
	fmt.Println("Migrating to draft-07...")
	migrate(schema)

	// Marshal with indentation
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("writing schema: %w", err)
	}

	fmt.Printf("Schema migrated successfully to %s\n", outputPath)
	return nil
}

// migrate transforms a draft-04 schema to draft-07 format in-place.
func migrate(schema map[string]any) {
	// Update $schema URI
	if s, ok := schema["$schema"].(string); ok && s == "http://json-schema.org/draft-04/schema#" {
		schema["$schema"] = draft07URI
	}

	// Recursively process all nested objects
	for key, value := range schema {
		switch v := value.(type) {
		case map[string]any:
			migrateObject(v)
		case []any:
			for _, item := range v {
				if obj, ok := item.(map[string]any); ok {
					migrateObject(obj)
				}
			}
		}
		// Handle id -> $id rename at top level
		if key == "id" {
			if _, hasID := schema["$id"]; !hasID {
				schema["$id"] = value
				delete(schema, "id")
			}
		}
	}
}

// migrateObject transforms a schema object in-place.
func migrateObject(obj map[string]any) {
	// Handle exclusiveMinimum/exclusiveMaximum conversion
	// Draft-04: { "minimum": 0, "exclusiveMinimum": true }
	// Draft-07: { "exclusiveMinimum": 0 }
	if exMin, ok := obj["exclusiveMinimum"].(bool); ok && exMin {
		if min, ok := obj["minimum"]; ok {
			obj["exclusiveMinimum"] = min
			delete(obj, "minimum")
		}
	}
	if exMax, ok := obj["exclusiveMaximum"].(bool); ok && exMax {
		if max, ok := obj["maximum"]; ok {
			obj["exclusiveMaximum"] = max
			delete(obj, "maximum")
		}
	}

	// Handle id -> $id rename
	if id, ok := obj["id"]; ok {
		if _, has := obj["$id"]; !has {
			obj["$id"] = id
			delete(obj, "id")
		}
	}

	// Recursively process nested objects
	for _, value := range obj {
		switch v := value.(type) {
		case map[string]any:
			migrateObject(v)
		case []any:
			for _, item := range v {
				if nested, ok := item.(map[string]any); ok {
					migrateObject(nested)
				}
			}
		}
	}
}
