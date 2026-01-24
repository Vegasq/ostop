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

func TestApp_CountFields_SimpleFields(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"field1": map[string]interface{}{
			"type": "text",
		},
		"field2": map[string]interface{}{
			"type": "keyword",
		},
		"field3": map[string]interface{}{
			"type": "integer",
		},
	}

	count := app.countFields(properties)
	expected := 3

	if count != expected {
		t.Errorf("countFields() = %d, want %d", count, expected)
	}
}

func TestApp_CountFields_NestedFields(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"user": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "text",
				},
				"email": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	count := app.countFields(properties)
	// Should count "user" + "name" + "email" = 3
	expected := 3

	if count != expected {
		t.Errorf("countFields() with nested fields = %d, want %d", count, expected)
	}
}

func TestApp_CountFields_DeepNesting(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{
		"level1": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"level2": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"level3": map[string]interface{}{
							"type": "text",
						},
					},
				},
			},
		},
	}

	count := app.countFields(properties)
	// Should count all levels: level1 + level2 + level3 = 3
	expected := 3

	if count != expected {
		t.Errorf("countFields() with deep nesting = %d, want %d", count, expected)
	}
}

func TestApp_CountFields_EmptyProperties(t *testing.T) {
	app := &App{}

	properties := map[string]interface{}{}

	count := app.countFields(properties)
	expected := 0

	if count != expected {
		t.Errorf("countFields() with empty properties = %d, want %d", count, expected)
	}
}

func TestApp_FormatFieldType(t *testing.T) {
	app := &App{}

	tests := []struct {
		name      string
		fieldType string
	}{
		{"text", "text"},
		{"keyword", "keyword"},
		{"long", "long"},
		{"integer", "integer"},
		{"short", "short"},
		{"byte", "byte"},
		{"double", "double"},
		{"float", "float"},
		{"half_float", "half_float"},
		{"scaled_float", "scaled_float"},
		{"date", "date"},
		{"boolean", "boolean"},
		{"object", "object"},
		{"nested", "nested"},
		{"unknown_type", "unknown_type"},
		{"binary", "binary"},
		{"geo_point", "geo_point"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.formatFieldType(tt.fieldType)
			// Just verify it returns a non-empty string
			// The actual formatting includes ANSI codes which are hard to test
			if result == "" {
				t.Errorf("formatFieldType(%q) returned empty string", tt.fieldType)
			}
		})
	}
}

func TestApp_RenderResourcesView_NoNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{},
	}

	result := app.renderResourcesView()

	if !strings.Contains(result, "No node data available") {
		t.Error("renderResourcesView() with no nodes should display 'No node data available'")
	}
}

func TestApp_RenderResourcesView_WithNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{
				Name:            "node1",
				HeapPercent:     "50",
				CPU:             "30",
				RAMPercent:      "60",
				DiskUsedPercent: "40",
			},
			{
				Name:            "node2",
				HeapPercent:     "70",
				CPU:             "50",
				RAMPercent:      "80",
				DiskUsedPercent: "60",
			},
		},
	}

	result := app.renderResourcesView()

	expectedStrings := []string{
		"Resource Utilization Dashboard",
		"Cluster Averages",
		"JVM Heap:",
		"CPU:",
		"RAM:",
		"Disk:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderResourcesView() should contain %q", expected)
		}
	}
}

func TestApp_RenderIndexSchemaView_NoMapping(t *testing.T) {
	app := &App{
		selectedIndexName: "test-index",
		indexMapping:      nil,
	}

	result := app.renderIndexSchemaView()

	if !strings.Contains(result, "test-index") {
		t.Error("renderIndexSchemaView() should contain index name")
	}

	if !strings.Contains(result, "Loading mapping...") {
		t.Error("renderIndexSchemaView() with nil mapping should show 'Loading mapping...'")
	}
}

func TestApp_RenderIndexSchemaView_NoProperties(t *testing.T) {
	app := &App{
		selectedIndexName: "test-index",
		indexMapping: &IndexMapping{
			Mappings: map[string]interface{}{},
		},
	}

	result := app.renderIndexSchemaView()

	if !strings.Contains(result, "No properties found in mapping") {
		t.Error("renderIndexSchemaView() without properties should show error message")
	}
}

func TestApp_RenderIndexSchemaView_WithProperties(t *testing.T) {
	app := &App{
		selectedIndexName: "test-index",
		indexMapping: &IndexMapping{
			Mappings: map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
					"status": map[string]interface{}{
						"type": "keyword",
					},
				},
			},
		},
	}

	result := app.renderIndexSchemaView()

	expectedStrings := []string{
		"test-index",
		"Fields (2)",
		"title",
		"status",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderIndexSchemaView() should contain %q", expected)
		}
	}
}
