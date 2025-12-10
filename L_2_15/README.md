# Unix Shell реализация на Go - L_2_15

Простой Unix shell с поддержкой встроенных команд, внешних команд, конвейеров и обработкой сигналов

## Функциональность

### ✅ Встроенные команды

- **cd \<path\>** - смена текущей директории
  - `cd /tmp` - переход в директорию /tmp
  - `cd ~` - переход в домашнюю директорию
  - `cd` - переход в домашнюю директорию (без аргументов)
  - `cd ..` - переход в родительскую директорию

- **pwd** - вывод текущей директории
  - `pwd` - показывает полный путь текущей папки

- **echo \<args\>** - вывод аргументов
  - `echo hello world` - выводит "hello world"
  - `echo` - выводит пустую строку

- **kill \<pid\> [\<signal\>]** - отправка сигнала процессу
  - `kill 12345` - отправляет SIGTERM процессу с PID 12345
  - `kill 12345 9` - отправляет SIGKILL (сигнал 9)

- **ps** - список запущенных процессов
  - `ps` - выводит список процессов (обертка над системной ps)

- **exit** - выход из shell'а
  - `exit` - завершает работу shell'а

### ✅ Внешние команды

Поддержка запуска любых внешних команд через os/exec:
```bash
ls -la
grep pattern file
find . -name "*.go"
```

### ✅ Конвейеры (Pipelines)

Объединение команд через `|`:
```bash
ps aux | grep myprocess | wc -l
cat file.txt | grep error | sort
```

Как это работает:
1. Разбиваем командную строку по `|`
2. Создаем cmd объекты для каждой команды
3. Подключаем stdout одной команды к stdin следующей через pipe
4. Запускаем все команды параллельно
5. Ждем завершения всех

### ✅ Условное выполнение

- **&&** - выполняет следующую команду только если предыдущая успешна
  ```bash
  cd /tmp && ls
  make && make install
  ```

- **||** - выполняет следующую команду только если предыдущая неудачна
  ```bash
  cd /nonexistent || echo "Directory not found"
  test -f file || echo "File not found"
  ```

### ✅ Обработка сигналов

- **Ctrl+D (EOF)** - выход из shell'а
- **Ctrl+C (SIGINT)** - прерывание текущей команды, но не shell'а

## Архитектура

### Основные компоненты

```
Shell struct
├── reader *bufio.Reader    - чтение команд со stdin
├── writer io.Writer        - вывод результатов
└── cwd string              - текущая директория

Основные методы:
├── Run()                   - главный цикл shell'а
├── executeCommand()        - выполнение команды
├── executePipeline()       - выполнение конвейера
├── executeBuiltin()        - встроенные команды
├── parseConditionals()     - парсинг && и ||
├── parseCommand()          - разбор командной строки
├── getPrompt()             - формирование приглашения
└── handleSignals()         - обработка сигналов
```

## Использование

### Скриптовый режим

```bash
echo "cd /tmp" | go run main.go
echo "pwd" | go run main.go
echo "echo test && echo success" | go run main.go
```

## Примеры команд

### Встроенные команды
```bash
$ echo hello world
hello world

$ pwd
/home/user

$ cd /tmp
$ pwd
/tmp

$ kill 12345
(отправляет сигнал процессу)

$ ps
(список процессов)
```

### Конвейеры
```bash
$ ls | grep test
$ ps aux | wc -l
$ cat file.txt | grep error | sort
$ echo "a\nb\nc" | grep b
```

### Условные операторы
```bash
$ cd /tmp && echo "Success"
Success

$ cd /nonexistent || echo "Failed"
Failed

$ make && make install || echo "Build failed"
```

## Запуск тестов

```bash
# Все тесты
go test -v

# С покрытием
go test -v -cover
```