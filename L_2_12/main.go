package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

//Реализовать утилиту фильтрации текстового потока (аналог команды grep).
//Программа должна читать входной поток (STDIN или файл) и выводить строки, соответствующие заданному шаблону (подстроке или регулярному выражению).
//Необходимо поддерживать следующие флаги:
//-A N — после каждой найденной строки дополнительно вывести N строк после неё (контекст).
//-B N — вывести N строк до каждой найденной строки.
//-C N — вывести N строк контекста вокруг найденной строки (включает и до, и после; эквивалентно -A N -B N).
//-c — выводить только то количество строк, что совпадающих с шаблоном (т.е. вместо самих строк — число).
//-i — игнорировать регистр.
//-v — инвертировать фильтр: выводить строки, не содержащие шаблон.
//-F — воспринимать шаблон как фиксированную строку, а не регулярное выражение (т.е. выполнять точное совпадение подстроки).
//-n — выводить номер строки перед каждой найденной строкой.
//Программа должна поддерживать сочетания флагов (например, -C 2 -n -i – 2 строки контекста, вывод номеров, без учета регистра и т.д.).
//Результат работы должен максимально соответствовать поведению команды UNIX grep.
//Обязательно учесть пограничные случаи (начало/конец файла для контекста, повторяющиеся совпадения и пр.).
//Код должен быть чистым, отформатированным (gofmt), работать без ситуаций гонки и успешно проходить golint.

// GrepConfig хранит все настройки утилиты, полученные из флагов.
type GrepConfig struct {
	after      int
	before     int
	context    int
	count      bool
	ignoreCase bool
	invert     bool
	fixed      bool
	lineNum    bool
}

func main() {
	// Парсинг флагов из команды
	cfg := parseFlags()

	// Получение шаблона и имени файла из аргументов
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ошибка: не указан шаблон для поиска")
		os.Exit(1)
	}
	pattern := args[0]

	var reader io.Reader
	// читаем из файла...
	if len(args) > 1 {
		file, err := os.Open(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ошибка при открытии файла: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else { // ...или из стандартного ввода
		reader = os.Stdin
	}

	if err := RunGrep(cfg, pattern, reader, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка выполнения: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() GrepConfig {
	var cfg GrepConfig
	flag.IntVar(&cfg.after, "A", 0, "печатать N строк после совпадения")
	flag.IntVar(&cfg.before, "B", 0, "печатать N строк до совпадения")
	flag.IntVar(&cfg.context, "C", 0, "печатать N строк вокруг совпадения")
	flag.BoolVar(&cfg.count, "c", false, "печатать только количество совпадающих строк")
	flag.BoolVar(&cfg.ignoreCase, "i", false, "игнорировать регистр")
	flag.BoolVar(&cfg.invert, "v", false, "инвертировать поиск (печатать несовпадающие строки)")
	flag.BoolVar(&cfg.fixed, "F", false, "фиксированная строка, не регулярное выражение")
	flag.BoolVar(&cfg.lineNum, "n", false, "печатать номер строки")
	flag.Parse()

	// Флаг -C имеет приоритет и устанавливает -A и -B
	if cfg.context > 0 {
		cfg.after = cfg.context
		cfg.before = cfg.context
	}
	return cfg
}

// RunGrep выполняет основную логику фильтрации
func RunGrep(config GrepConfig, pattern string, reader io.Reader, writer io.Writer) error {
	// Подготовка регулярного выражения
	if config.fixed {
		pattern = regexp.QuoteMeta(pattern)
	}
	if config.ignoreCase {
		pattern = "(?i)" + pattern // игнорировать регистр
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("некорректное регулярное выражение: %w", err)
	}

	// Чтение всех строк в память
	scanner := bufio.NewScanner(reader)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения ввода: %w", err)
	}

	// Поиск совпадений
	var matchIndices []int
	for i, line := range lines {
		// Условие совпадения: (match AND NOT invert) OR (NOT match AND invert)
		// Это эквивалентно (match XOR invert)
		if re.MatchString(line) != config.invert {
			matchIndices = append(matchIndices, i)
		}
	}

	// Вывод результата
	if config.count {
		fmt.Fprintln(writer, len(matchIndices))
		return nil
	}

	// Определение строк для вывода с учетом контекста
	// Используем map, чтобы избежать дублирования при пересечении контекста
	linesToPrint := make(map[int]struct{})
	for _, idx := range matchIndices {
		start := idx - config.before
		if start < 0 {
			start = 0
		}
		end := idx + config.after
		if end >= len(lines) {
			end = len(lines) - 1
		}
		for i := start; i <= end; i++ {
			linesToPrint[i] = struct{}{}
		}
	}

	// Печать строк с разделителями
	lastPrinted := -2 // Начальное значение, чтобы не печатать "--" в самом начале
	for i := 0; i < len(lines); i++ {
		if _, ok := linesToPrint[i]; ok {
			if lastPrinted != -2 && i > lastPrinted+1 {
				fmt.Fprintln(writer, "--")
			}
			if config.lineNum {
				fmt.Fprintf(writer, "%d:", i+1)
			}
			fmt.Fprintln(writer, lines[i])
			lastPrinted = i
		}
	}

	return nil
}
