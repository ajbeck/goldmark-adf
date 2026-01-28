// Package adfschema provides JSON Schema validation for Atlassian Document Format.
package adfschema

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed adf-schema.json
var schemaJSON []byte

var (
	resolvedSchema *jsonschema.Resolved
	initOnce       sync.Once
	initErr        error
)

// initSchema lazily initializes the schema on first use.
func initSchema() error {
	initOnce.Do(func() {
		var s jsonschema.Schema
		if err := json.Unmarshal(schemaJSON, &s); err != nil {
			initErr = err
			return
		}
		resolvedSchema, initErr = s.Resolve(nil)
	})
	return initErr
}

// Validate validates ADF JSON against the Atlassian Document Format schema.
// It returns nil if the document is valid, or an error describing the validation failure.
func Validate(data []byte) error {
	if err := initSchema(); err != nil {
		return err
	}

	var instance any
	if err := json.Unmarshal(data, &instance); err != nil {
		return err
	}

	return resolvedSchema.Validate(instance)
}

// MustValidate is like Validate but panics on error.
// It is intended for use in tests.
func MustValidate(data []byte) {
	if err := Validate(data); err != nil {
		panic(err)
	}
}
