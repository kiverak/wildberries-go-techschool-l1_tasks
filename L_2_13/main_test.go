package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestParseFields тестирует парсинг полей
func TestParseFields(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []int
		shouldErr bool
	}{
		{
			name:     "single field",
			input:    "1",
			expected: []int{1},
		},
		{
			name:     "multiple fields",
			input:    "1,3,5",
			expected: []int{1, 3, 5},
		},
		{
			name:     "range",
			input:    "1-3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "mixed fields and ranges",
			input:    "1,3-5,7",
			expected: []int{1, 3, 4, 5, 7},
		},
		{
			name:     "overlapping ranges",
			input:    "1-3,2-4",
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "with spaces",
			input:    "1 , 3 - 5 , 7",
			expected: []int{1, 3, 4, 5, 7},
		},
		{
			name:      "empty string",
			input:     "",
			shouldErr: true,
		},
		{
			name:      "invalid number",
			input:     "1,a,3",
			shouldErr: true,
		},
		{
			name:      "zero field",
			input:     "0",
			shouldErr: true,
		},
		{
			name:      "negative field",
			input:     "-1",
			shouldErr: true,
		},
		{
			name:      "range start > end",
			input:     "5-3",
			shouldErr: true,
		},
		{
			name:      "invalid range format",
			input:     "1-2-3",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFields(tt.input)

			if (err != nil) != tt.shouldErr {
				t.Errorf("parseFields(%q) error = %v, shouldErr = %v", tt.input, err, tt.shouldErr)
				return
			}

			if !tt.shouldErr {
				if len(result) != len(tt.expected) {
					t.Errorf("parseFields(%q) length = %d, expected %d", tt.input, len(result), len(tt.expected))
					return
				}

				for i, v := range result {
					if v != tt.expected[i] {
						t.Errorf("parseFields(%q)[%d] = %d, expected %d", tt.input, i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

// TestRunCut тестирует основную функцию
func TestRunCut(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		config    CutConfig
		expected  string
		shouldErr bool
	}{
		{
			name: "basic tab-separated",
			input: "one\ttwo\tthree\n" +
				"alpha\tbeta\tgamma\n",
			config: CutConfig{
				fields:    "1,3",
				delimiter: "\t",
				separated: false,
			},
			expected: "one\tthree\nalpha\tgamma\n",
		},
		{
			name: "single field",
			input: "one\ttwo\tthree\n" +
				"alpha\tbeta\tgamma\n",
			config: CutConfig{
				fields:    "2",
				delimiter: "\t",
				separated: false,
			},
			expected: "two\nbeta\n",
		},
		{
			name: "range of fields",
			input: "a\tb\tc\td\te\n" +
				"1\t2\t3\t4\t5\n",
			config: CutConfig{
				fields:    "2-4",
				delimiter: "\t",
				separated: false,
			},
			expected: "b\tc\td\n2\t3\t4\n",
		},
		{
			name: "comma delimiter",
			input: "a,b,c\n" +
				"1,2,3\n",
			config: CutConfig{
				fields:    "1,3",
				delimiter: ",",
				separated: false,
			},
			expected: "a,c\n1,3\n",
		},
		{
			name: "custom delimiter",
			input: "a|b|c\n" +
				"1|2|3\n",
			config: CutConfig{
				fields:    "2",
				delimiter: "|",
				separated: false,
			},
			expected: "b\n2\n",
		},
		{
			name: "separated flag - skip lines without delimiter",
			input: "one\ttwo\n" +
				"nodash\n" +
				"alpha\tbeta\n",
			config: CutConfig{
				fields:    "1",
				delimiter: "\t",
				separated: true,
			},
			expected: "one\nalpha\n",
		},
		{
			name: "separated flag - all lines have delimiter",
			input: "one\ttwo\n" +
				"alpha\tbeta\n",
			config: CutConfig{
				fields:    "1",
				delimiter: "\t",
				separated: true,
			},
			expected: "one\nalpha\n",
		},
		{
			name: "fields out of bounds",
			input: "a\tb\tc\n" +
				"1\t2\t3\n",
			config: CutConfig{
				fields:    "1,5,10",
				delimiter: "\t",
				separated: false,
			},
			expected: "a\n1\n",
		},
		{
			name:  "empty input",
			input: "",
			config: CutConfig{
				fields:    "1",
				delimiter: "\t",
				separated: false,
			},
			expected: "",
		},
		{
			name:  "single column input",
			input: "first\nsecond\n",
			config: CutConfig{
				fields:    "1",
				delimiter: "\t",
				separated: false,
			},
			expected: "first\nsecond\n",
		},
		{
			name: "mixed ranges and individual fields",
			input: "a\tb\tc\td\te\tf\n" +
				"1\t2\t3\t4\t5\t6\n",
			config: CutConfig{
				fields:    "1,3-4,6",
				delimiter: "\t",
				separated: false,
			},
			expected: "a\tc\td\tf\n1\t3\t4\t6\n",
		},
		{
			name: "separated flag - empty delimiter check",
			input: "no_delimiter_line\n" +
				"with:delimiter\n",
			config: CutConfig{
				fields:    "1",
				delimiter: ":",
				separated: true,
			},
			expected: "with\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var output bytes.Buffer

			err := RunCut(tt.config, reader, &output)

			if (err != nil) != tt.shouldErr {
				t.Errorf("RunCut() error = %v, shouldErr = %v", err, tt.shouldErr)
				return
			}

			result := output.String()
			if result != tt.expected {
				t.Errorf("RunCut() output = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestParseFieldsErrorCases тестирует обработку ошибок
func TestParseFieldsErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"invalid char", "1,a,3"},
		{"negative number", "-5"},
		{"zero number", "0"},
		{"invalid range", "5-3"},
		{"triple dash", "1-2-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFields(tt.input)
			if err == nil {
				t.Errorf("parseFields(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// BenchmarkRunCut проверяет производительность на больших данных
func BenchmarkRunCut(b *testing.B) {
	// Создаем большой текстовый буфер
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("field1\tfield2\tfield3\tfield4\tfield5\n")
	}
	input := sb.String()

	cfg := CutConfig{
		fields:    "1,3,5",
		delimiter: "\t",
		separated: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		var output bytes.Buffer
		_ = RunCut(cfg, reader, &output)
	}
}
