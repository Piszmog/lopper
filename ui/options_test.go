package ui

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected Model
	}{
		{
			name:   "Path",
			option: Path("/path/to/file"),
			expected: Model{
				path: "/path/to/file",
			},
		},
		{
			name:   "Protected Branches",
			option: ProtectedBranches([]string{"master", "develop"}),
			expected: Model{
				protectedBranches: []string{"master", "develop"},
			},
		},
		{
			name:   "Dry-Run",
			option: DryRun(true),
			expected: Model{
				dryRun: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := Model{}
			test.option(&m)
			assert.Equal(t, test.expected, m)
		})
	}
}
