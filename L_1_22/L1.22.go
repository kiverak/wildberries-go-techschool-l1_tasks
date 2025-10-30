package main

import (
	"fmt"
	"math/big"
)

//Разработать программу, которая перемножает, делит, складывает, вычитает две числовых переменных a, b, значения
//которых > 2^20 (больше 1 миллион).
//
//Комментарий: в Go тип int справится с такими числами, но обратите внимание на возможное переполнение для ещё больших
//значений. Для очень больших чисел можно использовать math/big.

// add выполняет сложение двух чисел типа *big.Int.
func add(a, b *big.Int) *big.Int {
	result := new(big.Int)
	return result.Add(a, b)
}

// subtract выполняет вычитание одного числа типа *big.Int из другого.
func subtract(a, b *big.Int) *big.Int {
	result := new(big.Int)
	return result.Sub(a, b)
}

// multiply выполняет умножение двух чисел типа *big.Int.
func multiply(a, b *big.Int) *big.Int {
	result := new(big.Int)
	return result.Mul(a, b)
}

// divide выполняет целочисленное деление одного числа типа *big.Int на другое.
func divide(a, b *big.Int) *big.Int {
	result := new(big.Int)
	return result.Div(a, b)
}

func main() {
	// Инициализируем два больших числа.
	a := big.NewInt(1048577) // 2^20 + 1
	b := big.NewInt(2097153) // 2^21 + 1

	fmt.Printf("a = %s\n", a.String())
	fmt.Printf("b = %s\n\n", b.String())

	// Сложение
	addResult := add(a, b)
	fmt.Printf("a + b = %s\n", addResult.String())

	// Вычитание
	subResult := subtract(a, b)
	fmt.Printf("a - b = %s\n", subResult.String())

	// Умножение
	mulResult := multiply(a, b)
	fmt.Printf("a * b = %s\n", mulResult.String())

	// Деление
	divResult := divide(b, a)
	fmt.Printf("b / a = %s\n", divResult.String())
}
