package main

import (
	"bufio"
	"fmt"
	"os"
)

//Разработать программу, которая переворачивает подаваемую на вход строку.
//Например: при вводе строки «главрыба» вывод должен быть «абырвалг».
//Учтите, что символы могут быть в Unicode (русские буквы, emoji и пр.), то есть просто iterating по байтам может не подойти — нужен срез рун ([]rune).

func main() {
	fmt.Println("Введите строку для переворота:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	reversed := string(runes)
	fmt.Println("Перевернутая строка:")
	fmt.Println(reversed)
}
