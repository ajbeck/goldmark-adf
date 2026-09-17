// Command migrate-schema normalizes the vendored ADF schema to draft-07.
// Run from the repository root: go run ./scripts
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	sourcePath = "adfschema/upstream/full-57.5.0.json"
	outputPath = "adfschema/adf-schema.json"
	draft07URI = "http://json-schema.org/draft-07/schema#"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading upstream schema: %w", err)
	}
	output, err := normalize(data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, output, 0o644); err != nil {
		return fmt.Errorf("writing schema: %w", err)
	}
	return nil
}

func normalize(data []byte) ([]byte, error) {
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}
	if schema["$schema"] != "http://json-schema.org/draft-04/schema#" {
		return nil, fmt.Errorf("expected draft-04 schema, got %v", schema["$schema"])
	}
	migrate(schema)
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling schema: %w", err)
	}
	return append(output, '\n'), nil
}

// migrate visits schema objects, never instance data or maps of property names.
func migrate(schema map[string]any) {
	if schema["$schema"] == "http://json-schema.org/draft-04/schema#" {
		schema["$schema"] = draft07URI
	}
	if id, ok := schema["id"].(string); ok {
		if _, exists := schema["$id"]; !exists {
			schema["$id"] = id
		}
		delete(schema, "id")
	}
	for _, bound := range []struct{ exclusive, inclusive string }{
		{"exclusiveMinimum", "minimum"},
		{"exclusiveMaximum", "maximum"},
	} {
		if exclusive, ok := schema[bound.exclusive].(bool); ok {
			delete(schema, bound.exclusive)
			if exclusive {
				schema[bound.exclusive] = schema[bound.inclusive]
				delete(schema, bound.inclusive)
			}
		}
	}

	// Each value in these maps can be a schema; its key is a data property name.
	for _, keyword := range []string{"definitions", "properties", "patternProperties", "dependencies"} {
		if children, ok := schema[keyword].(map[string]any); ok {
			for _, child := range children {
				if sub, ok := child.(map[string]any); ok {
					migrate(sub)
				}
			}
		}
	}
	for _, keyword := range []string{"additionalProperties", "additionalItems", "not", "items", "allOf", "anyOf", "oneOf"} {
		switch children := schema[keyword].(type) {
		case map[string]any:
			migrate(children)
		case []any:
			for _, child := range children {
				if sub, ok := child.(map[string]any); ok {
					migrate(sub)
				}
			}
		}
	}
}
