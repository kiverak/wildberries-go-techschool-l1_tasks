package main

//Реализовать упрощённый аналог UNIX-утилиты sort (сортировка строк).
//Программа должна читать строки (из файла или STDIN) и выводить их отсортированными.
//Обязательные флаги (как в GNU sort):
//-k N — сортировать по столбцу (колонке) №N (разделитель — табуляция по умолчанию).
//Например, «sort -k 2» отсортирует строки по второму столбцу каждой строки.
//-n — сортировать по числовому значению (строки интерпретируются как числа).
//-r — сортировать в обратном порядке (reverse).
//-u — не выводить повторяющиеся строки (только уникальные).
//Дополнительные флаги:
//-M — сортировать по названию месяца (Jan, Feb, ... Dec), т.е. распознавать специфический формат дат.
//-b — игнорировать хвостовые пробелы (trailing blanks).
//-c — проверить, отсортированы ли данные; если нет, вывести сообщение об этом.
//-h — сортировать по числовому значению с учётом суффиксов (например, К = килобайт, М = мегабайт — человекочитаемые размеры).
//Программа должна корректно обрабатывать комбинации флагов (например, -nr — числовая сортировка в обратном порядке, и т.д.).
//Необходимо предусмотреть эффективную обработку больших файлов.
//Код должен проходить все тесты, а также проверки go vet и golint (понимание, что требуются надлежащие комментарии,
//имена и структура программы).

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Config holds the command-line flags
type Config struct {
	K int  // сортировать по столбцу (колонке)
	N bool // сортировать по числовому значению
	R bool // сортировать в обратном порядке
	U bool // выводить только уникальные строки
	M bool // сортировать по названию месяца
	B bool // игнорировать хвостовые пробелы
	C bool // проверить, отсортированы ли данные
	H bool // сортировать по человекочитаемым размерам
}

func main() {
	cfg := parseFlags()

	lines, err := readLines(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	sorter := NewLineSorter(lines, cfg)

	if cfg.C {
		if !sorter.IsSorted() {
			fmt.Println("file is not sorted")
			os.Exit(1)
		}
		os.Exit(0)
	}

	sorter.Sort()

	if cfg.U {
		lines = unique(lines)
	}

	printLines(lines)
}

func parseFlags() Config {
	var cfg Config
	flag.IntVar(&cfg.K, "k", 1, "sort by column (1-indexed)")
	flag.BoolVar(&cfg.N, "n", false, "sort numerically")
	flag.BoolVar(&cfg.R, "r", false, "reverse the result of comparisons")
	flag.BoolVar(&cfg.U, "u", false, "output only the first of an equal run")
	flag.BoolVar(&cfg.M, "M", false, "compare (unknown) < 'JAN' < ... < 'DEC'")
	flag.BoolVar(&cfg.B, "b", false, "ignore leading blanks")
	flag.BoolVar(&cfg.C, "c", false, "check for sorted input; do not sort")
	flag.BoolVar(&cfg.H, "h", false, "compare human readable numbers (e.g., 2K 1G)")
	flag.Parse()

	if cfg.K < 1 {
		fmt.Fprintln(os.Stderr, "error: column index must be greater than 0")
		os.Exit(1)
	}
	return cfg
}

func readLines(filename string) ([]string, error) {
	var reader io.Reader
	if filename == "" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// LineSorter encapsulates the sorting logic
type LineSorter struct {
	lines []string
	cfg   Config
}

func NewLineSorter(lines []string, cfg Config) *LineSorter {
	return &LineSorter{lines: lines, cfg: cfg}
}

func (s *LineSorter) IsSorted() bool {
	for i := 1; i < len(s.lines); i++ {
		if s.less(i, i-1) {
			return false
		}
	}
	return true
}

func (s *LineSorter) Sort() {
	sort.SliceStable(s.lines, s.less)
}

func (s *LineSorter) less(i, j int) bool {
	lineA, lineB := s.lines[i], s.lines[j]

	valA := s.getCompareValue(lineA)
	valB := s.getCompareValue(lineB)

	var isLess bool
	switch {
	case s.cfg.N:
		numA, errA := strconv.ParseFloat(valA, 64)
		numB, errB := strconv.ParseFloat(valB, 64)
		if errA != nil && errB == nil { // A - не число, B - число
			isLess = true // Нечисловые значения считаем меньше числовых
		} else if errA == nil && errB != nil { // A - число, B - не число
			isLess = false
		} else if errA != nil { // Оба не числа, сравниваем как строки
			isLess = valA < valB
		} else { // Оба числа
			isLess = numA < numB
		}
	case s.cfg.M:
		monthA := parseMonth(valA)
		monthB := parseMonth(valB)
		isLess = monthA < monthB
	case s.cfg.H:
		numA := parseHumanReadable(valA)
		numB := parseHumanReadable(valB)
		isLess = numA < numB
	default:
		isLess = valA < valB
	}

	if s.cfg.R {
		return !isLess
	}
	return isLess
}

func (s *LineSorter) getCompareValue(line string) string {
	if s.cfg.B {
		line = strings.TrimSpace(line)
	}

	fields := strings.Fields(line)
	if s.cfg.K > 0 && s.cfg.K <= len(fields) {
		return fields[s.cfg.K-1]
	}
	// если колонка -k не определена или вне диапазона, сортируем по всей строке
	return line
}

func printLines(lines []string) {
	for _, line := range lines {
		fmt.Println(line)
	}
}

func unique(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	result := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			result = append(result, lines[i])
		}
	}
	return result
}

var monthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

func parseMonth(s string) int {
	s = strings.ToLower(s)
	if len(s) > 3 {
		s = s[:3]
	}
	if val, ok := monthMap[s]; ok {
		return val
	}

	// неизвестный месяц возвращаем как 0
	return 0
}

func parseHumanReadable(s string) int64 {
	s = strings.ToUpper(s)
	var multiplier int64 = 1
	suffix := ""
	if len(s) > 1 {
		lastChar := s[len(s)-1]
		if lastChar >= 'A' && lastChar <= 'Z' {
			suffix = string(lastChar)
		}
	}

	switch suffix {
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	}

	numStr := strings.TrimSuffix(s, suffix)
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0
	}

	return num * multiplier
}
