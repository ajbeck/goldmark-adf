package adf

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/ajbeck/goldmark-adf/v2/adfschema"
)

func TestConvert_StatusColors(t *testing.T) {
	for _, tc := range []struct {
		color string
		valid bool
	}{
		{"neutral", true}, {"purple", true}, {"blue", true},
		{"red", true}, {"yellow", true}, {"green", true},
		{"#000000", true}, {"#abcdef", true}, {"#ABCDEF", true}, {"#12AbEf", true},
		{"#123", false}, {"#12345678", false}, {"123456", false},
		{"#12ABCG", false}, {"#１２３", false}, {"orange", false}, {"", false},
	} {
		t.Run(tc.color, func(t *testing.T) {
			input := `[status:Ready|` + tc.color + `]`
			output, err := ConvertWithGFM([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			if err := adfschema.Validate(output); err != nil {
				t.Fatalf("invalid ADF: %v", err)
			}
			var doc Document
			if err := json.Unmarshal(output, &doc); err != nil {
				t.Fatal(err)
			}
			if len(doc.Content) != 1 || doc.Content[0].Type != "paragraph" {
				t.Fatalf("expected one paragraph: %s", output)
			}
			nodes := doc.Content[0].Content
			if tc.valid {
				if len(nodes) != 1 || nodes[0].Type != "status" {
					t.Fatalf("expected status node: %s", output)
				}
				if nodes[0].Attrs["color"] != tc.color || nodes[0].Attrs["text"] != "Ready" {
					t.Errorf("unexpected status attributes: %v", nodes[0].Attrs)
				}
			} else {
				var text strings.Builder
				for _, node := range nodes {
					if node.Type != "text" {
						t.Fatalf("invalid color should remain text: %s", output)
					}
					text.WriteString(node.Text)
				}
				if text.String() != input {
					t.Errorf("fallback text = %q, want %q", text.String(), input)
				}
			}
		})
	}
}
