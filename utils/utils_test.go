package utils_test

import (
	"github.com/stretchr/testify/assert"
	"lopper/utils"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		target   string
		expected bool
	}{
		{
			name:     "Contains Target",
			input:    []string{"a", "b", "c"},
			target:   "a",
			expected: true,
		},
		{
			name:     "Does Not Contain Target",
			input:    []string{"a", "b", "c"},
			target:   "d",
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := utils.Contains(test.input, test.target)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestTrimNewline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trim Newline",
			input:    "a\n",
			expected: "a",
		},
		{
			name:     "Trim Nothing",
			input:    "a",
			expected: "a",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, utils.TrimNewline(test.input))
		})
	}
}
