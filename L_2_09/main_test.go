package main

import (
	"testing"
)

func TestUnpackString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "Standard case",
			input:    "a4bc2d5e",
			expected: "aaaabccddddde",
		},
		{
			name:     "No repetition case",
			input:    "abcd",
			expected: "abcd",
		},
		{
			name:        "Invalid string (only numbers)",
			input:       "45",
			expectError: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Escaped digits",
			input:    "qwe\\4\\5",
			expected: "qwe45",
		},
		{
			name:     "Escaped digit with repetition",
			input:    "qwe\\45",
			expected: "qwe44444",
		},
		{
			name:     "Escaped backslash with repetition",
			input:    "qwe\\\\5",
			expected: "qwe\\\\\\\\\\",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := UnpackString(tc.input)

			// Случай, когда мы ожидаем ошибку
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error for input %q, but got none", tc.input)
				}
				return
			}

			// Случай, когда мы НЕ ожидаем ошибку
			if err != nil {
				t.Errorf("did not expect an error for input %q, but got: %v", tc.input, err)
			}

			// Проверяем, что результат соответствует ожидаемому
			if result != tc.expected {
				t.Errorf("for input %q, expected %q, but got %q", tc.input, tc.expected, result)
			}
		})
	}
}
