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
	go s.handleSignals(sigChan)

	for {
		// Вывод приглашения
		fmt.Fprint(s.writer, s.getPrompt())

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

	user, _ := user.Current()
	var prefix string
	if user.Uid == "0" {
		prefix = "#"
	} else {
		prefix = "$"
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
				continue // ||: выполнять только если предыдущая команда ошибка
			}
		}

		lastExitCode = s.executePipeline(cmd)
	}
}

// parseConditionals разбирает условные операторы && и ||
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
			}
			line = line[i+2:]
			i = 0
		} else {
			i++
		}
	}

	if strings.TrimSpace(line) != "" {
		commands = append(commands, strings.TrimSpace(line))
	}

	return commands, operators
}

// executePipeline выполняет конвейер команд
func (s *Shell) executePipeline(line string) int {
	// Разбиваем по символу |
	parts := strings.Split(line, "|")
	if len(parts) == 1 {
		// Одна команда без конвейера
		return s.executeSingleCommand(strings.TrimSpace(line))
	}

	// Несколько команд в конвейере
	var cmds []*exec.Cmd

	for i, part := range parts {
		part = strings.TrimSpace(part)
		cmd := s.parseCommand(part)
		if cmd == nil {
			return 1
		}

		// Последняя команда - выводим в консоль
		if i == len(parts)-1 {
			cmd.Stdout = s.writer
		}

		cmds = append(cmds, cmd)
	}

	// Подключаем конвейер
	for i := 0; i < len(cmds)-1; i++ {
		r, w, _ := os.Pipe()
		cmds[i].Stdout = w
		cmds[i+1].Stdin = r
	}

	// Запускаем все команды
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(s.writer, "ошибка: %v\n", err)
			return 1
		}
	}

	// Ждем завершения всех команд
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

// parsCommand разбирает строку команды в exec.Cmd
func (s *Shell) parseCommand(line string) *exec.Cmd {
	// Разбиваем на аргументы
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	// Проверяем встроенные команды
	builtin := parts[0]
	switch builtin {
	case "cd", "pwd", "echo", "kill", "ps", "exit":
		// Встроенные команды выполняются отдельно, возвращаем nil
		return nil
	}

	// Внешняя команда
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stderr = os.Stderr
	return cmd
}

// executeSingleCommand выполняет одну команду
func (s *Shell) executeSingleCommand(line string) int {
	if strings.TrimSpace(line) == "" {
		return 0
	}

	// Разбиваем на аргументы
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return 0
	}

	// Выполняем встроенную команду если существует
	if s.executeBuiltin(parts) {
		return 0
	}

	// Выполняем внешнюю команду
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = s.writer
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(s.writer, "ошибка: %v\n", err)
		return 1
	}

	return 0
}

// executeBuiltin выполняет встроенную команду
func (s *Shell) executeBuiltin(parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "cd":
		s.builtinCd(args)
		return true

	case "pwd":
		fmt.Fprintln(s.writer, s.cwd)
		return true

	case "echo":
		fmt.Fprintln(s.writer, strings.Join(args, " "))
		return true

	case "kill":
		s.builtinKill(args)
		return true

	case "ps":
		s.builtinPs()
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

// builtinPs выводит список процессов
func (s *Shell) builtinPs() {
	cmd := exec.Command("ps", "aux")
	cmd.Stdout = s.writer
	cmd.Stderr = s.writer
	cmd.Run()
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
