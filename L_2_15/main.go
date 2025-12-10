package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Необходимо реализовать собственный простейший Unix shell.

// Требования
// Ваш интерпретатор командной строки должен поддерживать:

// Встроенные команды:
// – cd <path> – смена текущей директории.
// – pwd – вывод текущей директории.
// – echo <args> – вывод аргументов.
// – kill <pid> – послать сигнал завершения процессу с заданным PID.
// – ps – вывести список запущенных процессов.

// Запуск внешних команд через exec (с помощью системных вызовов fork/exec либо стандартных функций os/exec).
// Конвейеры (pipelines): возможность объединять команды через |, чтобы вывод одной команды направлять на ввод следующей (как в обычном shell).
// Например: ps | grep myprocess | wc -l.
// Обработку завершения: при нажатии Ctrl+D (EOF) шелл должен завершаться; Ctrl+C — прерывание текущей запущенной команды, но без закрыватия самой shell.
// Дополнительно: реализовать парсинг && и || (условное выполнение команд), подстановку переменных окружения $VAR, поддержку редиректов >/< для вывода в
// файл и чтения из файла.
// Основной упор необходимо делать на реализацию базового функционала (exec, builtins, pipelines). Проверять надо как интерактивно, так и скриптом. Код должен
// работать без ситуаций гонки, корректно освобождать ресурсы.
// Совет: используйте пакеты os/exec, bufio (для ввода), strings.Fields (для разбиения командной строки на аргументы) и системные вызовы через syscall,
// если потребуется.

func main() {
	shell := NewShell(bufio.NewReader(os.Stdin), os.Stdout)

	if err := shell.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ошибка shell: %v\n", err)
		os.Exit(1)
	}
}

// Shell представляет интерпретатор команд
type Shell struct {
	reader *bufio.Reader
	writer io.Writer
	cwd    string // Текущая директория
	exit   bool   // Флаг для выхода
}

// NewShell создает новый экземпляр shell'а
func NewShell(reader *bufio.Reader, writer io.Writer) *Shell {
	cwd, _ := os.Getwd()
	return &Shell{
		reader: reader,
		writer: writer,
		cwd:    cwd,
	}
}

// Run запускает интерактивный режим shell'а
func (s *Shell) Run() error {
	// Обработка Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go s.handleSignals(sigChan) // выводит новую строку при получении прерывания Ctrl+C
	defer func() {
		// Очистка ресурсов при выходе из Run
		signal.Stop(sigChan)
		close(sigChan)
	}()

	for {
		// Вывод приглашения
		fmt.Fprint(s.writer, s.getPrompt()) // печатаем текущую директорию

		// Чтение строки команды
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Ctrl+D - выход
				fmt.Fprintln(s.writer)
				return nil
			}
			return err
		}

		// Удаляем символ новой строки
		line = strings.TrimSuffix(line, "\n")

		// Пропускаем пустые строки
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Выполняем команду
		s.executeCommand(line)

		// Проверяем флаг выхода
		if s.exit {
			return nil
		}
	}
}

// getPrompt возвращает приглашение с текущей директорией
func (s *Shell) getPrompt() string {
	dir := s.cwd
	home, _ := os.UserHomeDir()

	// Замена домашней директории на ~
	if strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	currentUser, _ := user.Current()
	var prefix string
	if currentUser.Uid == "0" {
		prefix = "#" // admin
	} else {
		prefix = "$" // user
	}

	return fmt.Sprintf("[%s] %s %s ", getHostname(), dir, prefix)
}

// executeCommand выполняет командную строку
func (s *Shell) executeCommand(line string) {
	// Разбиваем по условным операторам && и ||
	commands, operators := s.parseConditionals(line)

	var lastExitCode int
	for i, cmd := range commands {
		// Проверяем условия выполнения
		if i > 0 {
			if operators[i-1] == "&&" && lastExitCode != 0 {
				break // &&: выполнять только если предыдущая команда успешна
			}
			if operators[i-1] == "||" && lastExitCode == 0 {
				continue // ||: выполнять только если предыдущая команда неуспешна
			}
		}

		lastExitCode = s.executePipeline(cmd)
	}
}

// parseConditionals парсит строку по условным операторам && и ||
func (s *Shell) parseConditionals(line string) ([]string, []string) {
	var commands []string
	var operators []string

	i := 0
	for i < len(line) {
		if i+1 < len(line) && (line[i:i+2] == "&&" || line[i:i+2] == "||") {
			// Найдена условная команда
			cmd := strings.TrimSpace(line[:i])

			if cmd != "" {
				commands = append(commands, cmd)
				operators = append(operators, line[i:i+2])
			} else {
				// Обработка случая, когда перед оператором нет команды
				fmt.Fprintln(s.writer, "syntax error near unexpected token", line[i:i+2])
				return commands, operators // Возвращаем текущий результат с ошибкой
			}

			line = line[i+2:]
			i = 0
		} else {
			i++
		}
	}

	// Проверка, что строка не начинается с оператора
	if len(commands) == 0 && len(operators) > 0 {
		fmt.Fprintln(s.writer, "syntax error near unexpected token", operators[0])
		return commands, operators // Возвращаем текущий результат с ошибкой
	}

	if strings.TrimSpace(line) != "" {
		commands = append(commands, strings.TrimSpace(line))
	}

	return commands, operators
}

// executePipeline выполняет конвейер команд
func (s *Shell) executePipeline(line string) int {
	// Парсим редиректы из всей строки
	line, redirect := parseRedirects(line)

	// Разбиваем по символу |
	parts := strings.Split(line, "|")
	if len(parts) == 1 {
		// Одна команда без конвейера
		singleLine := strings.TrimSpace(line)
		return s.executeSingleCommand(singleLine, redirect)
	}

	// Создаем слайс для хранения команд
	var cmds []*exec.Cmd
	// Создаем слайс для хранения "записывающих" концов труб, чтобы их можно было закрыть позже
	var pipes []io.Closer

	for i, part := range parts {
		part = strings.TrimSpace(part)
		cmd := s.parseCommand(part)
		if cmd == nil {
			// Закрываем все уже открытые трубы перед выходом
			for _, p := range pipes {
				p.Close()
			}
			fmt.Fprintf(s.writer, "ошибка: встроенная команда не может быть частью конвейера: %s\n", part)
			return 1
		}

		// Соединяем команды
		if i > 0 {
			// stdin текущей команды - это stdout предыдущей
			r, w, _ := os.Pipe()
			cmds[i-1].Stdout = w     // Предыдущая пишет в трубу
			cmd.Stdin = r            // Текущая читает из трубы
			pipes = append(pipes, w) // Сохраняем "записывающий" конец для закрытия
		}

		cmds = append(cmds, cmd)
	}

	var cleanupFuncs []func()

	// stdout последней команды
	if redirect.OutputFile != "" {
		// Применяем редирект к последней команде
		cleanup, err := applyRedirects(cmds[len(cmds)-1], redirect)
		cleanupFuncs = append(cleanupFuncs, cleanup)

		if err != nil {
			fmt.Fprintf(s.writer, "ошибка: %v\n", err)
			return 1
		}
	} else {
		// Выводим в консоль
		cmds[len(cmds)-1].Stdout = s.writer
	}

	// Если есть input редирект для первой команды
	if redirect.InputFile != "" && len(cmds) > 0 {
		file, err := os.Open(redirect.InputFile)
		if err != nil {
			fmt.Fprintf(s.writer, "ошибка: не удалось открыть файл для чтения '%s': %v\n", redirect.InputFile, err)
			return 1
		}
		cmds[0].Stdin = file
		cleanupFuncs = append(cleanupFuncs, func() { file.Close() })
	}

	defer func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}()

	// Запускаем все команды
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(s.writer, "ошибка запуска: %v\n", err)
			return 1
		}
	}

	// Закрываем все "записывающие" концы труб в родительском процессе
	for _, p := range pipes {
		p.Close()
	}

	// Ждем завершения всех команд, начиная с последней
	var exitCode int
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
	}

	return exitCode
}

// expandEnv расширяет переменные окружения в строке
// Поддерживает синтаксис $VAR и ${VAR}
func expandEnv(s string) string {
	return os.ExpandEnv(s)
}

// parsCommand разбирает строку команды в exec.Cmd
func (s *Shell) parseCommand(line string) *exec.Cmd {
	// Расширяем переменные окружения
	line = expandEnv(line)

	// Разбиваем на аргументы
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	// Встроенные команды, которые могут быть в конвейере (echo, ps)
	switch command {
	case "echo":
		// echo в конвейере: просто выводит аргументы
		cmd := exec.Command("echo", args...)
		cmd.Stderr = os.Stderr
		return cmd

	case "ps":
		// ps может быть в конвейере
		cmd := exec.Command("ps", args...)
		if len(args) == 0 {
			// Если нет аргументов, используем aux
			cmd = exec.Command("ps", "aux")
		}
		cmd.Stderr = os.Stderr
		return cmd

	// Встроенные команды, которые не могут быть в конвейере
	case "cd", "pwd", "kill", "exit":
		return nil
	}

	// Внешняя команда
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	return cmd
}

// Redirect содержит информацию о редиректах
type Redirect struct {
	InputFile  string // Файл для ввода (<)
	OutputFile string // Файл для вывода (>, >>)
	Append     bool   // Использовать >> вместо >
}

// parseRedirects парсит редиректы из командной строки
// Возвращает команду без редиректов и информацию о редиректах
func parseRedirects(line string) (string, Redirect) {
	redirect := Redirect{}

	// Парсим >>
	parts := strings.Split(line, ">>")
	if len(parts) > 1 {
		redirect.OutputFile = strings.TrimSpace(parts[len(parts)-1])
		redirect.Append = true
		// Восстанавливаем команду без >>
		line = strings.Join(parts[:len(parts)-1], ">>")
	}

	// Парсим >
	if redirect.OutputFile == "" {
		parts := strings.Split(line, ">")
		if len(parts) > 1 {
			redirect.OutputFile = strings.TrimSpace(parts[len(parts)-1])
			redirect.Append = false
			// Восстанавливаем команду без >
			line = strings.Join(parts[:len(parts)-1], ">")
		}
	}

	// Парсим <
	parts = strings.Split(line, "<")
	if len(parts) > 1 {
		redirect.InputFile = strings.TrimSpace(parts[len(parts)-1])
		// Восстанавливаем команду без <
		line = strings.Join(parts[:len(parts)-1], "<")
	}

	return strings.TrimSpace(line), redirect
}

// applyRedirects применяет редиректы к команде
// Возвращает функцию очистки для закрытия открытых файлов
func applyRedirects(cmd *exec.Cmd, redirect Redirect) (func(), error) {
	cleanup := func() {}

	// Применяем stdin редирект (<)
	if redirect.InputFile != "" {
		file, err := os.Open(redirect.InputFile)
		if err != nil {
			return cleanup, fmt.Errorf("не удалось открыть файл для чтения '%s': %v", redirect.InputFile, err)
		}
		cmd.Stdin = file
		oldCleanup := cleanup
		cleanup = func() {
			oldCleanup()
			file.Close()
		}
	}

	// Применяем stdout редирект (>, >>)
	if redirect.OutputFile != "" {
		var file *os.File
		var err error

		if redirect.Append {
			// Открываем файл в режиме добавления
			file, err = os.OpenFile(redirect.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			// Открываем файл в режиме перезаписи
			file, err = os.Create(redirect.OutputFile)
		}

		if err != nil {
			return cleanup, fmt.Errorf("не удалось открыть файл для записи '%s': %v", redirect.OutputFile, err)
		}

		cmd.Stdout = file
		oldCleanup := cleanup
		cleanup = func() {
			oldCleanup()
			file.Close()
		}
	}

	return cleanup, nil
}

// executeSingleCommand выполняет одну команду
func (s *Shell) executeSingleCommand(line string, redirect Redirect) int {
	if strings.TrimSpace(line) == "" {
		return 0
	}

	// Разбиваем на аргументы
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return 0
	}

	// Пытаемся получить exec.Cmd (работает для echo, ps, внешних команд)
	cmd := s.parseCommand(line)
	if cmd != nil {
		// Это внешняя команда или echo/ps
		cmd.Stdout = s.writer
		cmd.Stdin = os.Stdin

		// Применяем редиректы
		cleanup, err := applyRedirects(cmd, redirect)
		defer cleanup()

		if err != nil {
			fmt.Fprintf(s.writer, "ошибка: %v\n", err)
			return 1
		}

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode()
			}
			fmt.Fprintf(s.writer, "ошибка: %v\n", err)
			return 1
		}
		return 0
	}

	// Расширяем переменные для встроенных команд
	line = expandEnv(line)
	parts = strings.Fields(line)

	// Выполняем встроенную команду (cd, pwd, kill, exit)
	if s.executeBuiltin(parts) {
		return 0
	}

	// Неизвестная команда
	fmt.Fprintf(s.writer, "команда не найдена: %s\n", parts[0])
	return 1
}

// executeBuiltin выполняет встроенную команду (только те, что не могут быть в конвейере)
func (s *Shell) executeBuiltin(parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "cd":
		// Расширяем переменные в аргументах cd
		if len(args) > 0 {
			args[0] = expandEnv(args[0])
		}
		s.builtinCd(args)
		return true

	case "pwd":
		fmt.Fprintln(s.writer, s.cwd)
		return true

	case "kill":
		s.builtinKill(args)
		return true

	case "exit":
		s.exit = true
		return true

	default:
		return false
	}
}

// builtinCd меняет текущую директорию
func (s *Shell) builtinCd(args []string) {
	if len(args) == 0 {
		// cd без аргументов - переход в домашнюю директорию
		home, _ := os.UserHomeDir()
		args = []string{home}
	}

	path := args[0]

	// Раскрываем ~ на домашнюю директорию
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	// Если относительный путь, делаем его абсолютным
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.cwd, path)
	}

	// Меняем директорию
	if err := os.Chdir(path); err != nil {
		fmt.Fprintf(s.writer, "cd: %v\n", err)
		return
	}

	s.cwd = path
	os.Setenv("PWD", path)
}

// builtinKill отправляет сигнал процессу
func (s *Shell) builtinKill(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(s.writer, "kill: требуется PID")
		return
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(s.writer, "kill: неверный PID: %s\n", args[0])
		return
	}

	// По умолчанию SIGTERM
	signal := syscall.SIGTERM
	if len(args) > 1 {
		if sig, err := strconv.Atoi(args[1]); err == nil {
			signal = syscall.Signal(sig)
		}
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(s.writer, "kill: процесс не найден: %v\n", err)
		return
	}

	if err := proc.Signal(signal); err != nil {
		fmt.Fprintf(s.writer, "kill: ошибка: %v\n", err)
	}
}

// handleSignals обрабатывает сигналы
func (s *Shell) handleSignals(sigChan <-chan os.Signal) {
	for range sigChan {
		// Ctrl+C - просто переходим на новую строку
		fmt.Fprintln(s.writer)
	}
}

// getHostname возвращает имя хоста
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}
