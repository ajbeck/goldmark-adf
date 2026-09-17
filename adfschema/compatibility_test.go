package adfschema

import (
	"strings"
	"testing"
)

func TestValidate_DataIDs(t *testing.T) {
	paragraph := func(inline string) string {
		return `{"type":"paragraph","content":[` + inline + `]}`
	}
	for _, tc := range []struct{ name, block string }{
		{"annotation", paragraph(`{"type":"text","text":"Hello","marks":[{"type":"annotation","attrs":{"id":"a1","annotationType":"inlineComment"}}]}`)},
		{"emoji", paragraph(`{"type":"emoji","attrs":{"shortName":":smile:","id":"e1"}}`)},
		{"link", paragraph(`{"type":"text","text":"Hello","marks":[{"type":"link","attrs":{"href":"https://example.com","id":"l1"}}]}`)},
		{"inline media", paragraph(`{"type":"mediaInline","attrs":{"id":"m1","collection":"c1","type":"file"}}`)},
		{"file media", `{"type":"mediaSingle","attrs":{"layout":"center"},"content":[{"type":"media","attrs":{"type":"file","id":"m1","collection":"c1"}}]}`},
		{"datasource", `{"type":"blockCard","attrs":{"datasource":{"id":"d1","parameters":{},"views":[{"type":"table"}]}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := `{"version":1,"type":"doc","content":[` + tc.block + `]}`
			if err := Validate([]byte(doc)); err != nil {
				t.Errorf("valid data id rejected: %v", err)
			}
			if err := Validate([]byte(strings.ReplaceAll(doc, `"id":`, `"$id":`))); err == nil {
				t.Error("$id must not be accepted as an ADF data property")
			}
		})
	}
}

func TestValidate_Schema575(t *testing.T) {
	for _, tc := range []struct {
		name, block string
		valid       bool
	}{
		{"hex status", `{"type":"paragraph","content":[{"type":"status","attrs":{"text":"Ready","color":"#12AbEf"}}]}`, true},
		{"invalid hex status", `{"type":"paragraph","content":[{"type":"status","attrs":{"text":"Ready","color":"#123"}}]}`, false},
		{"code display attrs", `{"type":"codeBlock","attrs":{"wrap":true,"hideLineNumbers":false},"content":[{"type":"text","text":"code"}]}`, true},
		{"invalid code display attrs", `{"type":"codeBlock","attrs":{"wrap":"true"}}`, false},
		{"paragraph font size", `{"type":"paragraph","marks":[{"type":"fontSize","attrs":{"fontSize":"small"}}],"content":[{"type":"text","text":"Small"}]}`, true},
		{"invalid paragraph font size", `{"type":"paragraph","marks":[{"type":"fontSize","attrs":{"fontSize":"large"}}]}`, false},
		{"table vertical alignment", `{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","attrs":{"valign":"middle"},"content":[{"type":"paragraph"}]},{"type":"tableCell","attrs":{"valign":"bottom"},"content":[{"type":"paragraph"}]}]}]}`, true},
		{"layout vertical alignment", `{"type":"layoutSection","content":[{"type":"layoutColumn","attrs":{"width":50,"valign":"middle"},"content":[{"type":"paragraph"}]},{"type":"layoutColumn","attrs":{"width":50,"valign":"bottom"},"content":[{"type":"paragraph"}]}]}`, true},
		{"media data consumer", `{"type":"mediaSingle","attrs":{"layout":"center"},"content":[{"type":"media","attrs":{"type":"file","id":"m1","collection":"c1"},"marks":[{"type":"dataConsumer","attrs":{"sources":["source1"]}}]}]}`, true},
		{"inline media data consumer", `{"type":"paragraph","content":[{"type":"mediaInline","attrs":{"id":"m1","collection":"c1"},"marks":[{"type":"dataConsumer","attrs":{"sources":["source1"]}}]}]}`, true},
		{"leading nested list", `{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph"}]}]}]}]}`, true},
		{"leading nested task list", `{"type":"taskList","attrs":{"localId":"t1"},"content":[{"type":"taskList","attrs":{"localId":"t2"},"content":[{"type":"taskItem","attrs":{"localId":"t3","state":"TODO"}}]}]}`, true},
		{"invalid third list child", `{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph"},{"type":"paragraph"},{"type":"text","text":"invalid"}]}]}`, false},
		{"invalid third task child", `{"type":"taskList","attrs":{"localId":"t1"},"content":[{"type":"taskItem","attrs":{"localId":"t2","state":"TODO"}},{"type":"taskItem","attrs":{"localId":"t3","state":"TODO"}},{"type":"paragraph"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate([]byte(`{"version":1,"type":"doc","content":[` + tc.block + `]}`))
			if (err == nil) != tc.valid {
				t.Errorf("valid = %v, want %v; error: %v", err == nil, tc.valid, err)
			}
		})
	}
}
