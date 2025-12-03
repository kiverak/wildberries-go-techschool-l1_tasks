package main

import (
	"reflect"
	"testing"
)

func TestLineSorter_Sort(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cfg   Config
		want  []string
	}{
		{
			name:  "Simple sort",
			lines: []string{"c", "a", "b"},
			cfg:   Config{},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Reverse sort",
			lines: []string{"c", "a", "b"},
			cfg:   Config{R: true},
			want:  []string{"c", "b", "a"},
		},
		{
			name:  "Numeric sort",
			lines: []string{"10", "2", "1"},
			cfg:   Config{N: true},
			want:  []string{"1", "2", "10"},
		},
		{
			name:  "Numeric reverse sort",
			lines: []string{"10", "2", "1"},
			cfg:   Config{N: true, R: true},
			want:  []string{"10", "2", "1"},
		},
		{
			name:  "Column sort",
			lines: []string{"a 3", "c 1", "b 2"},
			cfg:   Config{K: 2},
			want:  []string{"c 1", "b 2", "a 3"},
		},
		{
			name:  "Column numeric sort",
			lines: []string{"a 10", "c 2", "b 1"},
			cfg:   Config{K: 2, N: true},
			want:  []string{"b 1", "c 2", "a 10"},
		},
		{
			name:  "Month sort",
			lines: []string{"Mar", "Jan", "Feb"},
			cfg:   Config{M: true},
			want:  []string{"Jan", "Feb", "Mar"},
		},
		{
			name:  "Human-numeric sort",
			lines: []string{"1G", "2K", "3M"},
			cfg:   Config{H: true},
			want:  []string{"2K", "3M", "1G"},
		},
		{
			name:  "Unique sort",
			lines: []string{"c", "a", "b", "a"},
			cfg:   Config{U: true},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Ignore trailing blanks",
			lines: []string{" b", "a "},
			cfg:   Config{B: true},
			want:  []string{"a ", " b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewLineSorter(tt.lines, tt.cfg)
			sorter.Sort()
			lines := sorter.lines
			if tt.cfg.U {
				lines = unique(lines)
			}
			if !reflect.DeepEqual(lines, tt.want) {
				t.Errorf("got %v, want %v", lines, tt.want)
			}
		})
	}
}

func TestLineSorter_IsSorted(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cfg   Config
		want  bool
	}{
		{
			name:  "Sorted",
			lines: []string{"a", "b", "c"},
			cfg:   Config{},
			want:  true,
		},
		{
			name:  "Not sorted",
			lines: []string{"c", "a", "b"},
			cfg:   Config{},
			want:  false,
		},
		{
			name:  "Reverse sorted",
			lines: []string{"c", "b", "a"},
			cfg:   Config{R: true},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorter := NewLineSorter(tt.lines, tt.cfg)
			if got := sorter.IsSorted(); got != tt.want {
				t.Errorf("IsSorted() = %v, want %v", got, tt.want)
			}
		})
	}
}
