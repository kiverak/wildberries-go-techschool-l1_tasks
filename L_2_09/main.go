package main

//Распаковка строки
//Написать функцию Go, осуществляющую примитивную распаковку строки, содержащей повторяющиеся символы/руны.
//
//Примеры работы функции:
//
//Вход: "a4bc2d5e"
//Выход: "aaaabccddddde"
//
//Вход: "abcd"
//Выход: "abcd" (нет цифр — ничего не меняется)
//
//Вход: "45"
//Выход: "" (некорректная строка, т.к. в строке только цифры — функция должна вернуть ошибку)
//
//Вход: ""
//Выход: "" (пустая строка -> пустая строка)
//
//Дополнительное задание
//Поддерживать escape-последовательности вида \:
//
//Вход: "qwe\4\5"
//Выход: "qwe45" (4 и 5 не трактуются как числа, т.к. экранированы)
//
//Вход: "qwe\45"
//Выход: "qwe44444" (\4 экранирует 4, поэтому распаковывается только 5)
//
//Требования к реализации
//Функция должна корректно обрабатывать ошибочные случаи (возвращать ошибку, например, через error), и проходить unit-тесты.
//
//Код должен быть статически анализируем (vet, golint).

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string format")

// UnpackString распаковывает строку, содержащую повторяющиеся символы
func UnpackString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var result strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		// Если символ — это цифра, но он не экранирован, это ошибка
		if unicode.IsDigit(char) {
			return "", ErrInvalidString
		}

		// Если символ — это escape-символ '\'
		if char == '\\' {
			i++
			if i >= len(runes) {
				return "", ErrInvalidString // Строка не может заканчиваться на '\'
			}
			char = runes[i]
		}

		// Определяем, сколько раз нужно повторить символ
		repeatCount := 1
		if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
			i++
			// Преобразуем руну-цифру в число
			repeatCount = int(runes[i] - '0')
		}

		// Добавляем символ в результат нужное количество раз.
		result.WriteString(strings.Repeat(string(char), repeatCount))
	}

	return result.String(), nil
}
