package main

import (
	"bufio"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestSortingLogic проверяет основную логику сравнения, используя сортировку в памяти
func TestSortingLogic(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cfg   Config
		want  []string
	}{
		{
			name:  "Simple sort",
			lines: []string{"c", "a", "b"},
			cfg:   Config{K: 1},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Reverse sort",
			lines: []string{"c", "a", "b"},
			cfg:   Config{K: 1, R: true},
			want:  []string{"c", "b", "a"},
		},
		{
			name:  "Numeric sort",
			lines: []string{"10", "2", "1"},
			cfg:   Config{K: 1, N: true},
			want:  []string{"1", "2", "10"},
		},
		{
			name:  "Numeric reverse sort",
			lines: []string{"10", "2", "1"},
			cfg:   Config{K: 1, N: true, R: true},
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
			cfg:   Config{K: 1, M: true},
			want:  []string{"Jan", "Feb", "Mar"},
		},
		{
			name:  "Human-numeric sort",
			lines: []string{"1G", "2K", "3M"},
			cfg:   Config{K: 1, H: true},
			want:  []string{"2K", "3M", "1G"},
		},
		{
			name:  "Ignore leading blanks",
			lines: []string{" b", "a "},
			cfg:   Config{K: 1, B: true},
			want:  []string{"a ", " b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем LineSorter для сортировки среза в памяти
			sorter := NewLineSorter(tt.lines, tt.cfg)
			sorter.Sort()
			lines := sorter.lines

			if !reflect.DeepEqual(lines, tt.want) {
				t.Errorf("got %v, want %v", lines, tt.want)
			}
		})
	}
}

// TestIsSorted проверяет функцию IsSorted, которая работает с файлами
func TestIsSorted_File(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cfg   Config
		want  bool
	}{
		{
			name:  "Sorted",
			lines: []string{"a", "b", "c"},
			cfg:   Config{K: 1},
			want:  true,
		},
		{
			name:  "Not sorted",
			lines: []string{"c", "a", "b"},
			cfg:   Config{K: 1},
			want:  false,
		},
		{
			name:  "Reverse sorted",
			lines: []string{"c", "b", "a"},
			cfg:   Config{K: 1, R: true},
			want:  true,
		},
		{
			name:  "Not reverse sorted",
			lines: []string{"a", "b", "c"},
			cfg:   Config{K: 1, R: true},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл с тестовыми данными
			tmpfile, err := os.CreateTemp("", "test-is-sorted-*.txt")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name()) // Гарантируем удаление файла

			content := strings.Join(tt.lines, "\n")
			if _, err := tmpfile.Write([]byte(content)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Повторно открываем файл для чтения, чтобы передать его в IsSorted
			file, err := os.Open(tmpfile.Name())
			if err != nil {
				t.Fatalf("Failed to open temp file for reading: %v", err)
			}
			defer file.Close()

			// Вызываем IsSorted с файлом в качестве io.Reader
			got, err := IsSorted(file, tt.cfg)
			if err != nil {
				t.Fatalf("IsSorted() returned an unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("IsSorted() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsSorted_Stdin проверяет функцию IsSorted при чтении из STDIN
func TestIsSorted_Stdin(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cfg   Config
		want  bool
	}{
		{
			name:  "STDIN Sorted",
			lines: []string{"1", "2", "10"},
			cfg:   Config{K: 1, N: true},
			want:  true,
		},
		{
			name:  "STDIN Not sorted",
			lines: []string{"c", "a", "b"},
			cfg:   Config{K: 1},
			want:  false,
		},
		{
			name:  "STDIN Reverse sorted",
			lines: []string{"c", "b", "a"},
			cfg:   Config{K: 1, R: true},
			want:  true,
		},
		{
			name:  "STDIN Not reverse sorted",
			lines: []string{"a", "b", "c"},
			cfg:   Config{K: 1, R: true},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем pipe: будем писать в 'w' и читать из 'r'
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe() failed: %v", err)
			}

			// В отдельной горутине пишем тестовые данные в pipe writer
			go func() {
				defer w.Close() // Закрываем writer, чтобы сигнализировать конец ввода (EOF)
				content := strings.Join(tt.lines, "\n")
				w.Write([]byte(content))
			}()

			// Вызываем IsSorted, передавая читающую часть pipe напрямую
			got, err := IsSorted(r, tt.cfg)
			if err != nil {
				t.Fatalf("IsSorted() returned an unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("IsSorted() with STDIN = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExternalSort выполняет интеграционный тест для всего процесса внешней сортировки.
func TestExternalSort(t *testing.T) {
	lines := []string{"c 1", "a 3", "b 2", "a 1", "c 2"}
	expected := []string{"a 3", "a 1", "b 2", "c 1", "c 2"}
	cfg := Config{K: 1}

	// Создаем временный входной файл
	inputFile, err := os.CreateTemp("", "test-input-*.txt")
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	content := strings.Join(lines, "\n")
	if _, err := inputFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to input file: %v", err)
	}
	inputFile.Close()

	// Создаем имя для временного выходного файла
	outputFile, err := os.CreateTemp("", "test-output-*.txt")
	if err != nil {
		t.Fatalf("Failed to create output file placeholder: %v", err)
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	// Запускаем внешнюю сортировку
	err = externalSort(inputFile.Name(), outputFile.Name(), cfg)
	if err != nil {
		t.Fatalf("externalSort() failed: %v", err)
	}

	// Читаем и проверяем результат
	file, err := os.Open(outputFile.Name())
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	var gotLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		gotLines = append(gotLines, scanner.Text())
	}

	if !reflect.DeepEqual(gotLines, expected) {
		t.Errorf("externalSort() result is incorrect:\ngot:  %v\nwant: %v", gotLines, expected)
	}
}
