package main

import (
	"fmt"
)

//Поменять местами два числа без использования временной переменной.
//
//Подсказка: примените сложение/вычитание или XOR-обмен.

func swapArithmetic(a, b int) (int, int) {
	a = a + b
	b = a - b
	a = a - b
	return a, b
}

func swapXOR(a, b int) (int, int) {
	a = a ^ b
	b = a ^ b
	a = a ^ b
	return a, b
}

func main() {
	x, y := 10, 20
	fmt.Printf("До арифметического обмена: x = %d, y = %d\n", x, y)
	x, y = swapArithmetic(x, y)
	fmt.Printf("После арифметического обмена: x = %d, y = %d\n\n", x, y)

	a, b := 5, 15
	fmt.Printf("До XOR обмена: a = %d, b = %d\n", a, b)
	a, b = swapXOR(a, b)
	fmt.Printf("После XOR обмена: a = %d, b = %d\n\n", a, b)
}
