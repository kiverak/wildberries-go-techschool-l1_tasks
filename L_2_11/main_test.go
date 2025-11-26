package main

import (
	"reflect"
	"testing"
)

func TestFindAnagramSets(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		wanted map[string][]string
	}{
		{
			name:  "Standard case",
			input: []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"},
			wanted: map[string][]string{
				"листок": {"листок", "слиток", "столик"},
				"пятак":  {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "Case with different words order",
			input: []string{"тяпка", "пятак", "пятка", "листок", "слиток", "столик", "стол"},
			wanted: map[string][]string{
				"листок": {"листок", "слиток", "столик"},
				"тяпка":  {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "Case with different registers",
			input: []string{"Пятак", "пятка", "Тяпка", "Листок"},
			wanted: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "Case with duplicates in input",
			input: []string{"пятак", "пятак", "тяпка", "тяпка"},
			wanted: map[string][]string{
				"пятак": {"пятак", "тяпка"},
			},
		},
		{
			name:   "No anagrams found",
			input:  []string{"один", "два", "три"},
			wanted: map[string][]string{},
		},
		{
			name:   "Empty input slice",
			input:  []string{},
			wanted: map[string][]string{},
		},
		{
			name:  "Multiple anagram groups",
			input: []string{"нос", "сон", "парк", "карп", "кот", "ток"},
			wanted: map[string][]string{
				"нос":  {"нос", "сон"},
				"парк": {"карп", "парк1"},
				"кот":  {"кот", "ток"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FindAnagramSets(tc.input)

			if !reflect.DeepEqual(result, tc.wanted) {
				t.Errorf("FindAnagramSets() = %v, want %v", result, tc.wanted)
			}
		})
	}
}
