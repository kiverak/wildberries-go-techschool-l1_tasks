package main

import (
	"fmt"
	"strings"
)

//Разработать программу, которая проверяет, что все символы в строке встречаются один раз (т.е. строка состоит из уникальных символов).
//Вывод: true, если все символы уникальны, false, если есть повторения. Проверка должна быть регистронезависимой,
//т.е. символы в разных регистрах считать одинаковыми.
//Например: "abcd" -> true, "abCdefAaf" -> false (повторяются a/A), "aabcd" -> false.
//Подумайте, какой структурой данных удобно воспользоваться для проверки условия.

// areCharsUnique проверяет, что все символы в строке уникальны (регистронезависимо)
func areCharsUnique(s string) bool {
	lowerStr := strings.ToLower(s)

	// Создаем map для хранения встреченных символов.
	// Ключ - rune (символ), значение - пустая структура для экономии памяти.
	seen := make(map[rune]struct{})

	for _, char := range lowerStr {
		if _, ok := seen[char]; ok {
			return false
		}
		seen[char] = struct{}{}
	}

	return true
}

func main() {
	testCases := []string{"abcd", "abCdefAaf", "aabcd", "qwerty", "QWERTYq"}

	for _, tc := range testCases {
		fmt.Printf("Строка: \"%s\" -> Уникальные символы: %v\n", tc, areCharsUnique(tc))
	}
}
