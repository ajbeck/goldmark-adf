package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestNormalize_PreservesInstanceData(t *testing.T) {
	input := []byte(`{
		"$schema":"http://json-schema.org/draft-04/schema#",
		"id":"https://example.com/schema",
		"properties":{
			"id":{"type":"string"},
			"minimum":{"type":"number"},
			"nested":{"id":"nested","properties":{"id":{"type":"string"}}}
		},
		"required":["id"],
		"definitions":{"id":{"type":"string"}},
		"patternProperties":{"^id$":{"id":"pattern","type":"string"}},
		"dependencies":{"id":["nested"],"nested":{"id":"dependency","required":["id"]}},
		"default":{"id":"data","exclusiveMinimum":true,"minimum":0},
		"enum":[{"id":"choice","exclusiveMaximum":false}],
		"items":[{"minimum":1,"exclusiveMinimum":true},{"maximum":5,"exclusiveMaximum":false}],
		"additionalItems":{"maximum":10,"exclusiveMaximum":true},
		"additionalProperties":{"id":"extra","type":"string"},
		"not":{"id":"negated","type":"null"},
		"allOf":[{"id":"all","type":"object"}],
		"anyOf":[{"id":"any","type":"object"}],
		"oneOf":[{"id":"one","type":"object"}]
	}`)
	want := []byte(`{
		"$schema":"http://json-schema.org/draft-07/schema#",
		"$id":"https://example.com/schema",
		"properties":{
			"id":{"type":"string"},
			"minimum":{"type":"number"},
			"nested":{"$id":"nested","properties":{"id":{"type":"string"}}}
		},
		"required":["id"],
		"definitions":{"id":{"type":"string"}},
		"patternProperties":{"^id$":{"$id":"pattern","type":"string"}},
		"dependencies":{"id":["nested"],"nested":{"$id":"dependency","required":["id"]}},
		"default":{"id":"data","exclusiveMinimum":true,"minimum":0},
		"enum":[{"id":"choice","exclusiveMaximum":false}],
		"items":[{"exclusiveMinimum":1},{"maximum":5}],
		"additionalItems":{"exclusiveMaximum":10},
		"additionalProperties":{"$id":"extra","type":"string"},
		"not":{"$id":"negated","type":"null"},
		"allOf":[{"$id":"all","type":"object"}],
		"anyOf":[{"$id":"any","type":"object"}],
		"oneOf":[{"$id":"one","type":"object"}]
	}`)
	output, err := normalize(input)
	if err != nil {
		t.Fatal(err)
	}
	var gotValue, wantValue any
	if err := json.Unmarshal(output, &gotValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("unexpected normalized schema:\n%s", output)
	}
}

func TestNormalize_EmbeddedSchemaMatchesSource(t *testing.T) {
	source, err := os.ReadFile("../" + sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := normalize(source)
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := os.ReadFile("../" + outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(generated, embedded) {
		t.Fatal("embedded schema differs from its source; run go run ./scripts from the repository root")
	}
}
