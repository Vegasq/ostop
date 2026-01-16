package ui

import (
	"strings"
	"testing"
)

func TestApp_RenderFields_BasicField(t *testing.T) {
	app := &App{}

	tests := []struct {
		name       string
		properties map[string]interface{}
		contains   []string
	}{
		{
			name: "simple_text_field",
			properties: map[string]interface{}{
				"username": map[string]interface{}{
					"type": "text",
				},
			},
			contains: []string{"username", "text"},
		},
		{
			name: "keyword_field",
			properties: map[string]interface{}{
				"status": map[string]interface{}{
					"type": "keyword",
				},
			},
			contains: []string{"status", "keyword"},
		},
		{
			name: "field_with_analyzer",
			properties: map[string]interface{}{
				"message": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
			},
			contains: []string{"message", "text", "standard"},
		},
		{
			name: "non_searchable_field",
			properties: map[string]interface{}{
				"binary_data": map[string]interface{}{
					"type":  "binary",
					"index": false,
				},
			},
			contains: []string{"binary_data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			app.renderFields(&b, tt.properties, 0)
			result := b.String()

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("renderFields() result should contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestApp_RenderFields_NestedFields(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"user": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "text",
				},
				"age": map[string]interface{}{
					"type": "integer",
				},
			},
		},
	}

	var b strings.Builder
	app.renderFields(&b, properties, 0)
	result := b.String()

	// Check that nested field names appear
	expectedFields := []string{"user", "name", "age"}
	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("renderFields() should contain nested field %q", field)
		}
	}
}

func TestApp_RenderFields_MultiFields(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"title": map[string]interface{}{
			"type": "text",
			"fields": map[string]interface{}{
				"keyword": map[string]interface{}{
					"type":         "keyword",
					"ignore_above": 256,
				},
			},
		},
	}

	var b strings.Builder
	app.renderFields(&b, properties, 0)
	result := b.String()

	// Check that multi-field structure is present
	if !strings.Contains(result, "title") {
		t.Error("renderFields() should contain main field 'title'")
	}

	if !strings.Contains(result, "keyword") {
		t.Error("renderFields() should contain multi-field 'keyword'")
	}
}

func TestApp_RenderFields_Indentation(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"field1": map[string]interface{}{
			"type": "text",
		},
	}

	// Test with different indentation levels
	indentLevels := []int{0, 1, 2, 3}

	for _, indent := range indentLevels {
		t.Run(strings.Repeat("indent_", indent+1), func(t *testing.T) {
			var b strings.Builder
			app.renderFields(&b, properties, indent)
			result := b.String()

			// Just verify it produces output
			if result == "" {
				t.Error("renderFields() should produce output")
			}
		})
	}
}
