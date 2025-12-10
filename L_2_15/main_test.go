package main

import (
	"bufio"
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuiltinEcho тестирует встроенную команду echo
func TestBuiltinEcho(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "simple echo",
			input:    "echo hello world\n",
			contains: "hello world",
		},
		{
			name:     "echo multiple args",
			input:    "echo a b c\n",
			contains: "a b c",
		},
		{
			name:     "echo empty",
			input:    "echo\n",
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := bufio.NewReader(strings.NewReader(tt.input + "exit\n"))
			output := &bytes.Buffer{}

			shell := NewShell(input, output)
			shell.Run()

			result := output.String()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("echo: expected to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

// TestBuiltinPwd тестирует встроенную команду pwd
func TestBuiltinPwd(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader("pwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	cwd, _ := os.Getwd()

	shell.Run()

	result := output.String()
	if !strings.Contains(result, cwd) {
		t.Errorf("pwd: expected to contain %q, got %q", cwd, result)
	}
}

// TestBuiltinCd тестирует встроенную команду cd
func TestBuiltinCd(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	originalCwd, _ := os.Getwd()
	defer os.Chdir(originalCwd)

	input := bufio.NewReader(strings.NewReader("cd " + tmpDir + "\npwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, tmpDir) {
		t.Errorf("cd: expected cwd to contain %q, got output %q", tmpDir, result)
	}
}

// TestBuiltinCdHome тестирует cd без аргументов (в домашнюю папку)
func TestBuiltinCdHome(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader("cd\npwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	home, _ := os.UserHomeDir()
	result := output.String()
	if !strings.Contains(result, home) {
		t.Errorf("cd: expected home directory, got %q", result)
	}
}

// TestBuiltinCdTilde тестирует cd с тильдой
func TestBuiltinCdTilde(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader("cd ~\npwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	home, _ := os.UserHomeDir()
	result := output.String()
	if !strings.Contains(result, home) {
		t.Errorf("cd ~: expected home directory, got %q", result)
	}
}

// TestParseConditionals тестирует парсинг условных операторов
func TestParseConditionals(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		numCmds int
		numOps  int
		firstOp string
	}{
		{
			name:    "single command",
			input:   "echo test",
			numCmds: 1,
			numOps:  0,
		},
		{
			name:    "with &&",
			input:   "echo a && echo b",
			numCmds: 2,
			numOps:  1,
			firstOp: "&&",
		},
		{
			name:    "with ||",
			input:   "echo a || echo b",
			numCmds: 2,
			numOps:  1,
			firstOp: "||",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			shell := NewShell(nil, nil)
			cmds, ops := shell.parseConditionals(tt.input)

			if len(cmds) != tt.numCmds {
				t.Errorf("parseConditionals: expected %d commands, got %d", tt.numCmds, len(cmds))
			}

			if len(ops) != tt.numOps {
				t.Errorf("parseConditionals: expected %d operators, got %d", tt.numOps, len(ops))
			}

			if tt.numOps > 0 && ops[0] != tt.firstOp {
				t.Errorf("parseConditionals: expected first operator %q, got %q", tt.firstOp, ops[0])
			}
		})
	}
}

// TestExternalCommand тестирует выполнение внешней команды
func TestExternalCommand(t *testing.T) {
	t.Parallel()
	// Используем команду которая есть на всех платформах
	input := bufio.NewReader(strings.NewReader("echo test\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "test") {
		t.Errorf("external command: expected to contain 'test', got %q", result)
	}
}

// TestGetPrompt тестирует формирование приглашения
func TestGetPrompt(t *testing.T) {
	t.Parallel()
	shell := NewShell(nil, nil)
	prompt := shell.getPrompt()

	// Проверяем что в приглашении есть необходимые элементы
	if !strings.Contains(prompt, "[") || !strings.Contains(prompt, "]") {
		t.Errorf("getPrompt: expected brackets in prompt, got %q", prompt)
	}

	if !strings.Contains(prompt, "$") && !strings.Contains(prompt, "#") {
		t.Errorf("getPrompt: expected $ or # in prompt, got %q", prompt)
	}
}

// TestExecuteBuiltin тестирует выполнение встроенных команд
func TestExecuteBuiltin(t *testing.T) {
	tests := []struct {
		name      string
		command   []string
		isBuiltin bool
	}{
		{
			name:      "echo",
			command:   []string{"echo", "test"},
			isBuiltin: false, // echo теперь работает как внешняя команда (может быть в конвейере)
		},
		{
			name:      "pwd",
			command:   []string{"pwd"},
			isBuiltin: true,
		},
		{
			name:      "cd",
			command:   []string{"cd", "/"},
			isBuiltin: true,
		},
		{
			name:      "kill",
			command:   []string{"kill", "12345"},
			isBuiltin: true,
		},
		{
			name:      "ps",
			command:   []string{"ps"},
			isBuiltin: false, // ps теперь работает как внешняя команда (может быть в конвейере)
		},
		{
			name:      "unknown",
			command:   []string{"unknowncommand"},
			isBuiltin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := &bytes.Buffer{}
			shell := NewShell(nil, output)

			result := shell.executeBuiltin(tt.command)
			if result != tt.isBuiltin {
				t.Errorf("executeBuiltin: expected %v, got %v", tt.isBuiltin, result)
			}
		})
	}
}

// TestGetHostname тестирует получение имени хоста
func TestGetHostname(t *testing.T) {
	t.Parallel()
	hostname := getHostname()
	if hostname == "" {
		t.Error("getHostname: expected non-empty hostname")
	}
}

// TestShellNewShell тестирует создание нового shell'а
func TestShellNewShell(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader(""))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)

	if shell == nil {
		t.Error("NewShell: expected non-nil shell")
	}

	if shell.cwd == "" {
		t.Error("NewShell: expected non-empty cwd")
	}

	cwd, _ := os.Getwd()
	if shell.cwd != cwd {
		t.Errorf("NewShell: expected cwd to be %q, got %q", cwd, shell.cwd)
	}
}

// TestUserInfo тестирует получение информации о пользователе
func TestUserInfo(t *testing.T) {
	t.Parallel()
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if currentUser == nil {
		t.Error("user.Current: expected non-nil user")
	}

	if currentUser.Username == "" {
		t.Error("user.Current: expected non-empty username")
	}
}

// TestCdWithRelativePath тестирует cd с относительным путем
func TestCdWithRelativePath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	originalCwd, _ := os.Getwd()
	defer os.Chdir(originalCwd)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	input := bufio.NewReader(strings.NewReader("cd " + tmpDir + "\ncd subdir\npwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, subDir) {
		t.Errorf("cd relative: expected %q in output, got %q", subDir, result)
	}
}

// TestEchoPipeline тестирует echo в конвейере
func TestEchoPipeline(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader("echo hello world | cat\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "hello world") {
		t.Errorf("echo pipeline: expected 'hello world' in output, got %q", result)
	}
}

// TestPsPipeline тестирует ps в конвейере
func TestPsPipeline(t *testing.T) {
	t.Parallel()

	// Пропускаем тест на Windows
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("ps command not available on Windows")
	}

	input := bufio.NewReader(strings.NewReader("ps | head -1\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	// ps должна вывести строки, даже если они отфильтрованы head
	if len(result) == 0 {
		t.Errorf("ps pipeline: expected some output, got empty")
	}
}

// TestComplexPipeline тестирует сложный конвейер с echo
func TestComplexPipeline(t *testing.T) {
	t.Parallel()
	input := bufio.NewReader(strings.NewReader("echo -e 'line1\\nline2\\nline3' | wc -l\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	// Вывод должен содержать результат wc -l
	if len(result) == 0 {
		t.Errorf("complex pipeline: expected output, got empty")
	}
}

// TestEnvVarExpansion тестирует подстановку переменных окружения
func TestEnvVarExpansion(t *testing.T) {
	t.Parallel()
	// Устанавливаем переменную окружения
	os.Setenv("TEST_VAR", "hello_world")
	defer os.Unsetenv("TEST_VAR")

	input := bufio.NewReader(strings.NewReader("echo $TEST_VAR\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "hello_world") {
		t.Errorf("env var expansion: expected 'hello_world' in output, got %q", result)
	}
}

// TestEnvVarExpansionBraces тестирует подстановку ${VAR}
func TestEnvVarExpansionBraces(t *testing.T) {
	t.Parallel()
	os.Setenv("TEST_BRACES", "test_value")
	defer os.Unsetenv("TEST_BRACES")

	input := bufio.NewReader(strings.NewReader("echo ${TEST_BRACES}\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "test_value") {
		t.Errorf("env var braces expansion: expected 'test_value' in output, got %q", result)
	}
}

// TestEnvVarInPipeline тестирует переменные в конвейере
func TestEnvVarInPipeline(t *testing.T) {
	t.Parallel()
	os.Setenv("PIPE_TEST", "pipeline_var")
	defer os.Unsetenv("PIPE_TEST")

	input := bufio.NewReader(strings.NewReader("echo $PIPE_TEST | cat\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "pipeline_var") {
		t.Errorf("env var in pipeline: expected 'pipeline_var' in output, got %q", result)
	}
}

// TestEnvVarInCd тестирует переменные в команде cd
func TestEnvVarInCd(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	os.Setenv("TEST_DIR", tmpDir)
	defer os.Unsetenv("TEST_DIR")

	originalCwd, _ := os.Getwd()
	defer os.Chdir(originalCwd)

	input := bufio.NewReader(strings.NewReader("cd $TEST_DIR\npwd\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, tmpDir) {
		t.Errorf("env var in cd: expected %q in output, got %q", tmpDir, result)
	}
}

// TestOutputRedirect тестирует редирект > для вывода в файл
func TestOutputRedirect(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	input := bufio.NewReader(strings.NewReader("echo hello world > " + outputFile + "\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	// Проверяем, что файл создан и содержит правильный текст
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("не удалось прочитать файл: %v", err)
	}

	if !strings.Contains(string(content), "hello world") {
		t.Errorf("output redirect: expected 'hello world' in file, got %q", string(content))
	}
}

// TestInputRedirect тестирует редирект < для чтения из файла
func TestInputRedirect(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")

	// Создаем файл с тестовым содержимым
	err := os.WriteFile(inputFile, []byte("test content\n"), 0644)
	if err != nil {
		t.Fatalf("не удалось создать файл: %v", err)
	}

	input := bufio.NewReader(strings.NewReader("cat < " + inputFile + "\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	result := output.String()
	if !strings.Contains(result, "test content") {
		t.Errorf("input redirect: expected 'test content' in output, got %q", result)
	}
}

// TestAppendRedirect тестирует редирект >> для добавления в файл
func TestAppendRedirect(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "append.txt")

	// Создаем файл с начальным содержимым
	err := os.WriteFile(outputFile, []byte("first line\n"), 0644)
	if err != nil {
		t.Fatalf("не удалось создать файл: %v", err)
	}

	// Добавляем строку через >>
	input := bufio.NewReader(strings.NewReader("echo second line >> " + outputFile + "\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	// Проверяем содержимое файла
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("не удалось прочитать файл: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "first line") {
		t.Errorf("append redirect: expected 'first line' in file, got %q", contentStr)
	}
	if !strings.Contains(contentStr, "second line") {
		t.Errorf("append redirect: expected 'second line' in file, got %q", contentStr)
	}
}

// TestPipeRedirect тестирует комбинацию pipeline и redirect
func TestPipeRedirect(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "pipe_output.txt")

	input := bufio.NewReader(strings.NewReader("echo line1 | cat > " + outputFile + "\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)
	shell.Run()

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("не удалось прочитать файл: %v", err)
	}

	if !strings.Contains(string(content), "line1") {
		t.Errorf("pipe redirect: expected 'line1' in file, got %q", string(content))
	}
}

// TestParseRedirects тестирует парсинг редиректов
func TestParseRedirects(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCmd    string
		expectedInput  string
		expectedOutput string
		expectedAppend bool
	}{
		{
			name:           "output redirect",
			input:          "echo hello > output.txt",
			expectedCmd:    "echo hello",
			expectedOutput: "output.txt",
		},
		{
			name:          "input redirect",
			input:         "cat < input.txt",
			expectedCmd:   "cat",
			expectedInput: "input.txt",
		},
		{
			name:           "append redirect",
			input:          "echo line >> output.txt",
			expectedCmd:    "echo line",
			expectedOutput: "output.txt",
			expectedAppend: true,
		},
		{
			name:        "no redirect",
			input:       "ls -la",
			expectedCmd: "ls -la",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, redirect := parseRedirects(tt.input)

			if cmd != tt.expectedCmd {
				t.Errorf("parseRedirects: expected cmd %q, got %q", tt.expectedCmd, cmd)
			}
			if redirect.InputFile != tt.expectedInput {
				t.Errorf("parseRedirects: expected input %q, got %q", tt.expectedInput, redirect.InputFile)
			}
			if redirect.OutputFile != tt.expectedOutput {
				t.Errorf("parseRedirects: expected output %q, got %q", tt.expectedOutput, redirect.OutputFile)
			}
			if redirect.Append != tt.expectedAppend {
				t.Errorf("parseRedirects: expected append %v, got %v", tt.expectedAppend, redirect.Append)
			}
		})
	}
}
