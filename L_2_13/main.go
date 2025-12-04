package main

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

// Реализовать утилиту, которая считывает входные данные (STDIN) и разбивает каждую строку по заданному разделителю, после чего выводит определённые поля (колонки).
// Аналог команды cut с поддержкой флагов:
// -f "fields" — указание номеров полей (колонок), которые нужно вывести. Номера через запятую, можно диапазоны.
// Например: «-f 1,3-5» — вывести 1-й и с 3-го по 5-й столбцы.
// -d "delimiter" — использовать другой разделитель (символ). По умолчанию разделитель — табуляция ('\t').
// -s – (separated) только строки, содержащие разделитель. Если флаг указан, то строки без разделителя игнорируются (не выводятся).
// Программа должна корректно парсить аргументы, поддерживать различные комбинации (например, несколько отдельных полей и диапазонов), учитывать,
// что номера полей могут выходить за границы (в таком случае эти поля просто игнорируются).
// Стоит обратить внимание на эффективность при обработке больших файлов. Все стандартные требования по качеству кода и тестам также применимы.

// СгеConfig хранит все настройки утилиты, полученные из флагов
type CutConfig struct {
	fields    string
	delimiter string
	separated bool
}

func main() {
	// Парсинг флагов из команды
	cfg := parseFlags()

	// читаем из файла...
	var reader io.Reader
	args := flag.Args()
	if len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка при открытии файла: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else { // ...или из стандартного ввода
		reader = os.Stdin
	}

	if err := RunCut(cfg, reader, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка выполнения: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags парсит командные флаги и возвращает конфигурацию
func parseFlags() CutConfig {
	fields := flag.String("f", "", "поля для вывода (например: 1,3-5)")
	delimiter := flag.String("d", "\t", "разделитель полей")
	separated := flag.Bool("s", false, "только строки с разделителем")
	flag.Parse()

	return CutConfig{
		fields:    *fields,
		delimiter: *delimiter,
		separated: *separated,
	}
}

// RunCut выполняет основную логику
func RunCut(cfg CutConfig, reader io.Reader, writer io.Writer) error {
	fields, err := parseFields(cfg.fields)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)       // используем буферизированное чтение
	bufferedWriter := bufio.NewWriter(writer) // используем буферизированную запись
	defer bufferedWriter.Flush()

	for scanner.Scan() {
		line := scanner.Text()

		// Проверяем наличие разделителя если требуется (-s флаг)
		if cfg.separated && !strings.Contains(line, cfg.delimiter) {
			continue
		}

		// Разбиваем строку по разделителю
		parts := strings.Split(line, cfg.delimiter)

		// Собираем нужные поля (индексация начинается с 1)
		var result []string
		for _, fieldNum := range fields {
			// fieldNum начинается с 1, а индекс массива с 0
			idx := fieldNum - 1
			if idx < len(parts) {
				result = append(result, parts[idx])
			}
			// Если поле выходит за границы, игнорируем
		}

		if len(result) > 0 {
			_, err := bufferedWriter.WriteString(strings.Join(result, cfg.delimiter) + "\n")
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// parseFields парсит строку полей и возвращает отсортированный список уникальных номеров
func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("не указаны поля для вывода")
	}

	fieldSet := make(map[int]bool)

	// Разбиваем по запятым
	parts := strings.SplitSeq(fieldsStr, ",") // эффективнее, чем strings.Split, нет лишних аллокаций
	for part := range parts {
		part = strings.TrimSpace(part)

		// Проверяем, это диапазон или одиночное число
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неверный формат диапазона: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("неверное число в диапазоне: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("неверное число в диапазоне: %s", rangeParts[1])
			}

			if start < 1 || end < 1 {
				return nil, fmt.Errorf("номера полей должны быть >= 1")
			}

			if start > end {
				return nil, fmt.Errorf("начало диапазона больше конца: %d > %d", start, end)
			}

			for i := start; i <= end; i++ {
				fieldSet[i] = true
			}
		} else {
			field, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("неверное число поля: %s", part)
			}

			if field < 1 {
				return nil, fmt.Errorf("номер поля должен быть >= 1")
			}

			fieldSet[field] = true
		}
	}

	// Преобразуем множество в отсортированный список
	result := make([]int, 0, len(fieldSet))
	for field := range fieldSet {
		result = append(result, field)
	}
	sort.Ints(result)

	return result, nil
}
