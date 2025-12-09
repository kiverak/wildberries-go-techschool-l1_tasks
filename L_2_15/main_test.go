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
			isBuiltin: true,
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
			isBuiltin: true,
		},
		{
			name:      "unknown",
			command:   []string{"unknowncommand"},
			isBuiltin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	hostname := getHostname()
	if hostname == "" {
		t.Error("getHostname: expected non-empty hostname")
	}
}

// TestShellNewShell тестирует создание нового shell'а
func TestShellNewShell(t *testing.T) {
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

// BenchmarkShellExecution измеряет производительность выполнения команд
func BenchmarkShellExecution(b *testing.B) {
	input := bufio.NewReader(strings.NewReader("echo benchmark\necho benchmark\necho benchmark\nexit\n"))
	output := &bytes.Buffer{}

	shell := NewShell(input, output)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shell.executeCommand("echo test")
	}
}
