package main

import (
	"testing"
)

func TestConvertBoolean(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "boolean"}

	tests := []struct {
		input    string
		expected bool
		hasError bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"1", true, false},
		{"0", false, false},
		{"on", true, false},
		{"off", false, false},
		{"TRUE", true, false},
		{"FALSE", false, false},
		{"invalid", false, true},
		{"maybe", false, true},
	}

	for _, test := range tests {
		result, err := convertBoolean(test.input, schema)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			} else if result != test.expected {
				t.Errorf("For input %q, expected %v, got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestConvertInteger(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "integer"}

	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"42", 42, false},
		{"-10", -10, false},
		{"0", 0, false},
		{"3.14", 0, true}, // Should reject decimals
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := convertInteger(test.input, schema)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			} else if result != test.expected {
				t.Errorf("For input %q, expected %v, got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestConvertNumber(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "number"}

	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"3.14", 3.14, false},
		{"-0.5", -0.5, false},
		{"42", 42.0, false},
		{"0", 0.0, false},
		{"abc", 0, true},
	}

	for _, test := range tests {
		result, err := convertNumber(test.input, schema)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			} else if result != test.expected {
				t.Errorf("For input %q, expected %v, got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestConvertArray(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "array"}

	tests := []struct {
		input    string
		expected []any
		hasError bool
	}{
		{`["a","b","c"]`, []any{"a", "b", "c"}, false},
		{"a,b,c", []any{"a", "b", "c"}, false},
		{"", []any{}, false},
		{`[1,2,3]`, []any{float64(1), float64(2), float64(3)}, false}, // JSON numbers are float64
		{"single", []any{"single"}, false},
	}

	for _, test := range tests {
		result, err := convertArray(test.input, schema)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			} else {
				resultSlice, ok := result.([]any)
				if !ok {
					t.Errorf("Result is not []any for input %q", test.input)
					continue
				}
				if len(resultSlice) != len(test.expected) {
					t.Errorf("For input %q, expected length %d, got %d", test.input, len(test.expected), len(resultSlice))
					continue
				}
				for i, v := range resultSlice {
					if v != test.expected[i] {
						t.Errorf("For input %q at index %d, expected %v, got %v", test.input, i, test.expected[i], v)
					}
				}
			}
		}
	}
}

func TestExtractFullSchema(t *testing.T) {
	// Test with a typical JSON schema
	schemaData := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The name parameter",
			},
			"count": map[string]any{
				"type":    "integer",
				"default": 10,
			},
			"enabled": map[string]any{
				"type": "boolean",
			},
		},
		"required": []any{"name"},
	}

	schema, err := extractFullSchema(schemaData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(schema.Parameters) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(schema.Parameters))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required=['name'], got %v", schema.Required)
	}

	nameParam := schema.Parameters["name"]
	if nameParam == nil {
		t.Fatal("name parameter not found")
	}
	if nameParam.Type != "string" {
		t.Errorf("Expected name type 'string', got %q", nameParam.Type)
	}
	if !nameParam.Required {
		t.Error("Expected name to be required")
	}

	countParam := schema.Parameters["count"]
	if countParam == nil {
		t.Fatal("count parameter not found")
	}
	if countParam.Type != "integer" {
		t.Errorf("Expected count type 'integer', got %q", countParam.Type)
	}
	if countParam.Default != 10 {
		t.Errorf("Expected count default 10, got %v", countParam.Default)
	}
}
