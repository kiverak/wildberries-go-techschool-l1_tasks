package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunGrep(t *testing.T) {
	testCases := []struct {
		name        string
		config      GrepConfig
		pattern     string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "Simple match",
			config:   GrepConfig{},
			pattern:  "world",
			input:    "hello\nworld\nand universe",
			expected: "world\n",
		},
		{
			name:     "Ignore case (-i)",
			config:   GrepConfig{ignoreCase: true},
			pattern:  "go",
			input:    "Go is great\ngo is fast\nstop",
			expected: "Go is great\ngo is fast\n",
		},
		{
			name:     "Invert match (-v)",
			config:   GrepConfig{invert: true},
			pattern:  "skip",
			input:    "line1\nskip this\nline3",
			expected: "line1\n--\nline3\n",
		},
		{
			name:     "Count matches (-c)",
			config:   GrepConfig{count: true},
			pattern:  "a",
			input:    "alpha\nbeta\ngamma",
			expected: "3\n",
		},
		{
			name:     "Line numbers (-n)",
			config:   GrepConfig{lineNum: true},
			pattern:  "second",
			input:    "first\nsecond\nthird",
			expected: "2:second\n",
		},
		{
			name:     "Fixed string (-F)",
			config:   GrepConfig{fixed: true},
			pattern:  "a.b",
			input:    "a.b\nacb\na_b",
			expected: "a.b\n",
		},
		{
			name:     "After context (-A)",
			config:   GrepConfig{after: 2},
			pattern:  "match",
			input:    "before\nmatch\nafter1\nafter2\nanother",
			expected: "match\nafter1\nafter2\n",
		},
		{
			name:     "Before context (-B)",
			config:   GrepConfig{before: 1},
			pattern:  "match",
			input:    "before1\nmatch\nafter1",
			expected: "before1\nmatch\n",
		},
		{
			name:     "Context (-C)",
			config:   GrepConfig{after: 1, before: 1},
			pattern:  "match",
			input:    "line1\nmatch\nline3",
			expected: "line1\nmatch\nline3\n",
		},
		{
			name:     "Overlapping context",
			config:   GrepConfig{after: 1},
			pattern:  "match",
			input:    "match1\nmatch2\nline3",
			expected: "match1\nmatch2\nline3\n",
		},
		{
			name:     "Separated context with --",
			config:   GrepConfig{after: 1},
			pattern:  "match",
			input:    "match1\nline2\nline3\nmatch2",
			expected: "match1\nline2\n--\nmatch2\n",
		},
		{
			name:        "Invalid regex",
			config:      GrepConfig{},
			pattern:     "[",
			input:       "some data",
			expectError: true,
		},
		{
			name:     "Combination of flags (-i, -n, -C 1)",
			config:   GrepConfig{ignoreCase: true, lineNum: true, after: 1, before: 1},
			pattern:  "Test",
			input:    "line before\ntest line\nline after",
			expected: "1:line before\n2:test line\n3:line after\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputReader := strings.NewReader(tc.input)
			var outputBuffer bytes.Buffer

			err := RunGrep(tc.config, tc.pattern, inputReader, &outputBuffer)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got := outputBuffer.String(); got != tc.expected {
				t.Errorf("unexpected output:\ngot:\n%s\nwant:\n%s", got, tc.expected)
			}
		})
	}
}
