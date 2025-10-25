package main

import "fmt"

//Разработать программу, которая переворачивает порядок слов в строке.
//Пример: входная строка:
//«snow dog sun», выход: «sun dog snow».
//Считайте, что слова разделяются одиночным пробелом. Постарайтесь не использовать дополнительные срезы, а выполнять операцию «на месте».

// reverse переворачивает руну "на месте"
func reverse(r []rune) {
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
}

// reverseWords переворачивает порядок слов в строке
func reverseWords(s string) string {
	runes := []rune(s)

	// переворачиваем весь срез целиком
	// "snow dog sun" -> "nus god wons"
	reverse(runes)

	// переворачиваем каждое слово по-отдельности
	// "nus god wons" -> "sun dog snow"
	start := 0
	for i := 0; i < len(runes); i++ {
		if runes[i] == ' ' {
			// нашли окончание слова, переворачиваем его
			reverse(runes[start:i])
			// передвигаем start на начало следующего слова
			start = i + 1
		}
	}
	// разворачиваем последнее слово
	reverse(runes[start:])

	return string(runes)
}

func main() {
	input := "snow dog sun"
	fmt.Printf("Input:  %s\n", input)
	output := reverseWords(input)
	fmt.Printf("Output: %s\n", output)
}
