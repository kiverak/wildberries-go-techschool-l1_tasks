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
	"container/heap"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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

	if cfg.C {
		var reader io.Reader = os.Stdin
		filename := flag.Arg(0)
		if filename != "" {
			file, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()
			reader = file
		}

		isSorted, err := IsSorted(reader, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !isSorted {
			fmt.Println("file is not sorted")
			os.Exit(1)
		}
		os.Exit(0)
	}

	// большой файл не получится прочитать целиком, поэтому будем использовать внешнюю сортировку
	err := externalSort(flag.Arg(0), "sorted_output.txt", cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Sorted output written to sorted_output.txt")
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

// externalSort выполняет внешнюю сортировку для больших файлов
func externalSort(inputFile, outputFile string, cfg Config) error {
	// Разделение на отсортированные чанки
	tempFiles, err := createSortedChunks(inputFile, cfg)
	if err != nil {
		return err
	}
	// Гарантируем удаление временных файлов
	defer func() {
		for _, f := range tempFiles { // tempFiles может быть nil, если произошла ошибка до создания файлов
			if f != "" { // Проверяем, что имя файла не пустое
				os.Remove(f)
			}
		}
	}()

	// Слияние временных файлов в один выходной файл
	return mergeChunks(tempFiles, outputFile, cfg)
}

// createSortedChunks читает файл по частям, сортирует и пишет во временные файлы
func createSortedChunks(filename string, cfg Config) ([]string, error) {
	// читаем из стандартного ввода или из файла
	var reader io.Reader
	if filename == "" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	// Каналы для коммуникации между горутинами
	chunkChan := make(chan []string, 10) // канал для сырых чанков
	resultChan := make(chan string, 10)  // канал для имен временных файлов
	errChan := make(chan error, 1)       // канал для ошибок от горутин

	scanner := bufio.NewScanner(reader)
	chunkSize := 100000 // устанавливаем размер чанка - 100_000 строк на чанк
	lines := make([]string, 0, chunkSize)

	// Горутина-читатель
	go func() {
		defer close(chunkChan) // Закрываем канал чанков, когда чтение завершено
		for {
			readCount := 0
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				readCount++
				if readCount >= chunkSize {
					break
				}
			}

			if len(lines) == 0 {
				break // Достигли конца файла
			}

			// Отправляем чанк в канал. Делаем копию, чтобы избежать гонки.
			// Если lines[:0] будет вызван до того, как worker обработает чанк, worker увидит пустой чанк
			chunkCopy := make([]string, len(lines))
			copy(chunkCopy, lines)
			chunkChan <- chunkCopy

			lines = lines[:0] // Очищаем срез для следующего чанка

			if readCount < chunkSize {
				break // Достигли конца файла, выходим
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case errChan <- fmt.Errorf("scanner error: %w", err):
			default:
			}
		}
	}()

	var wg sync.WaitGroup
	numWorkers := 4 // количество горутин-работников

	// Запускаем горутины-работники
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				// Сортируем текущий чанк
				sorter := NewLineSorter(chunk, cfg)
				sorter.Sort()

				// Создаем временный файл
				tmpFile, err := os.CreateTemp("", "sort-chunk-*.txt")
				if err != nil {
					select {
					case errChan <- fmt.Errorf("failed to create temp file: %w", err):
					default:
					}
					return
				}

				// Пишем отсортированные строки во временный файл
				writer := bufio.NewWriter(tmpFile)
				for _, line := range chunk {
					fmt.Fprintln(writer, line)
				}
				if err := writer.Flush(); err != nil {
					select {
					case errChan <- fmt.Errorf("failed to flush writer to temp file %s: %w", tmpFile.Name(), err):
					default:
					}
					tmpFile.Close()
					os.Remove(tmpFile.Name()) // Удаляем поврежденный файл
					return
				}
				if err := tmpFile.Close(); err != nil {
					select {
					case errChan <- fmt.Errorf("failed to close temp file %s: %w", tmpFile.Name(), err):
					default:
					}
					os.Remove(tmpFile.Name()) // Удаляем поврежденный файл
					return
				}
				resultChan <- tmpFile.Name()
			}
		}()
	}

	// Горутина, которая будет ждать завершения всех работников и закрывать resultChan
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var collectedTempFiles []string
	var firstErr error

	// Собираем имена временных файлов и проверяем ошибки
	for {
		select {
		case tmpFileName, ok := <-resultChan:
			if !ok { // Канал закрыт, все работники завершили работу
				return collectedTempFiles, firstErr
			}
			collectedTempFiles = append(collectedTempFiles, tmpFileName)
		case err := <-errChan:
			if firstErr == nil { // Сохраняем только первую ошибку
				firstErr = err
			}
		}
	}
}

// heapItem представляет элемент в куче для слияния
type heapItem struct {
	line    string // Строка из файла
	fileIdx int    // Индекс файла, из которого прочитана строка
}

// minHeap реализует heap.Interface для heapItem
type minHeap struct {
	items  []*heapItem
	sorter *LineSorter
}

// Len возвращает количество элементов в куче
func (h *minHeap) Len() int {
	return len(h.items)
}

// Less сравнивает элементы в куче
func (h *minHeap) Less(i, j int) bool {
	return h.sorter.compareLines(h.items[i].line, h.items[j].line)
}

// Swap меняет местами элементы в куче
func (h *minHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

// Push добавляет элемент в кучу
func (h *minHeap) Push(x any) {
	h.items = append(h.items, x.(*heapItem))
}

// Pop удаляет и возвращает элемент из кучи
func (h *minHeap) Pop() any {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // удаляем элемент из памяти
	h.items = old[0 : n-1]
	return item
}

// mergeChunks сливает отсортированные временные файлы в один
func mergeChunks(files []string, outputFile string, cfg Config) error {
	scanners := make([]*bufio.Scanner, len(files)) // Слайс сканеров для каждого файла
	fileHandles := make([]*os.File, len(files))    // Слайс для открытия каждого файла

	for i, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		fileHandles[i] = file
		scanners[i] = bufio.NewScanner(file)
	}

	defer func() {
		for _, f := range fileHandles {
			f.Close()
		}
	}()

	// сейчас у нас открыто i временных файлов и i сканеров для них

	// Инициализация кучи
	h := &minHeap{
		items:  make([]*heapItem, 0, len(files)), // слайс для хранения первых строк из каждого файла
		sorter: NewLineSorter(nil, cfg),          // сортировщик
	}
	for i, scanner := range scanners {
		if scanner.Scan() {
			heap.Push(h, &heapItem{line: scanner.Text(), fileIdx: i})
		}
	}

	// Открываем выходной файл для записи
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	var lastWritten string
	for h.Len() > 0 {
		item := heap.Pop(h).(*heapItem)

		if !cfg.U || item.line != lastWritten { // учитываем флаг U для записи только уникальных строк
			fmt.Fprintln(writer, item.line)
			lastWritten = item.line
		}

		// Читаем следующую строку из того же файла и добавляем в кучу
		if scanners[item.fileIdx].Scan() {
			heap.Push(h, &heapItem{line: scanners[item.fileIdx].Text(), fileIdx: item.fileIdx})
		}
	}

	return writer.Flush()
}

// LineSorter содержит логику сортировки
type LineSorter struct {
	lines []string
	cfg   Config
}

// IsSorted проверяет, отсортирован ли ввод (из файла или STDIN) в соответствии с конфигурацией
// читает ввод построчно, не загружая весь файл в память
func IsSorted(reader io.Reader, cfg Config) (bool, error) {
	scanner := bufio.NewScanner(reader)
	var previousLine string
	firstLine := true

	sorter := NewLineSorter(nil, cfg)

	for scanner.Scan() {
		currentLine := scanner.Text()
		if firstLine {
			previousLine = currentLine
			firstLine = false
			continue
		}

		if sorter.compareLines(currentLine, previousLine) {
			return false, nil
		}
		previousLine = currentLine
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading input for sorting check: %w", err)
	}

	return true, nil
}

// compareLines сравнивает строки в соответствии с конфигурацией
func (s *LineSorter) compareLines(lineA, lineB string) bool {
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

// getCompareValue возвращает значение для сравнения
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

// NewLineSorter создает новый экземпляр LineSorter
func NewLineSorter(lines []string, cfg Config) *LineSorter {
	return &LineSorter{lines: lines, cfg: cfg}
}

// Sort выполняет сортировку строк
func (s *LineSorter) Sort() {
	sort.SliceStable(s.lines, s.less)
}

// less определяет порядок сортировки
func (s *LineSorter) less(i, j int) bool {
	lineA, lineB := s.lines[i], s.lines[j]
	return s.compareLines(lineA, lineB)
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
