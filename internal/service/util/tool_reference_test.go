package util

import (
	"reflect"
	"testing"
)

func TestExtractToolIDs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no tool references",
			content:  "This is a prompt without any tools.",
			expected: []string{},
		},
		{
			name:     "single tool reference",
			content:  "Call {{tool:60b8d295f1d4f20001000001}} to find results.",
			expected: []string{"60b8d295f1d4f20001000001"},
		},
		{
			name:     "multiple unique tool references",
			content:  "Use {{tool:60b8d295f1d4f20001000001}} and then {{tool:60b8d295f1d4f20001000002}} for analysis.",
			expected: []string{"60b8d295f1d4f20001000001", "60b8d295f1d4f20001000002"},
		},
		{
			name:     "duplicate tool references",
			content:  "First {{tool:60b8d295f1d4f20001000001}}, then check, then {{tool:60b8d295f1d4f20001000001}} again.",
			expected: []string{"60b8d295f1d4f20001000001"},
		},
		{
			name:     "tool references with whitespace",
			content:  "Execute {{tool:  60b8d295f1d4f20001000001  }} now.",
			expected: []string{"60b8d295f1d4f20001000001"},
		},
		{
			name:     "empty tool reference",
			content:  "Invalid {{tool:}} reference.",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ExtractToolIDs(tt.content)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("ExtractToolIDs() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}
