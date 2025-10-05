package main

import (
	"fmt"
)

//Реализовать пересечение двух неупорядоченных множеств (например, двух слайсов) — т.е. вывести элементы, присутствующие и в первом, и во втором.
//
//Пример:
//A = {1,2,3}
//B = {2,3,4}
//Пересечение = {2,3}

func intersection(a, b []int) []int {
	setA := make(map[int]bool)
	for _, item := range a {
		setA[item] = true
	}

	var result []int

	for _, item := range b {
		if _, found := setA[item]; found {
			result = append(result, item)
			// Удаляем элемент из setA на случай, если в слайсе B есть дубликаты.
			delete(setA, item)
		}
	}

	return result
}

func main() {
	A := []int{1, 2, 3, 5, 8, 1}
	B := []int{2, 3, 4, 8, 9, 3}

	result := intersection(A, B)

	fmt.Printf("Множество A: %v\n", A)
	fmt.Printf("Множество B: %v\n", B)
	fmt.Printf("Пересечение: %v\n", result) // [2 3 8]

	C := []int{1, 2, 3, 5, 8, 1}
	D := []int{2, 3, 4, 8, 9, 3}

	result1 := intersection(C, D)

	fmt.Printf("\nМножество A: %v\n", C)
	fmt.Printf("Множество B: %v\n", D)
	fmt.Printf("Пересечение: %v\n", result1) // [2 3 8]

	E := []int{10, 20, 30}
	F := []int{40, 50, 60}
	result2 := intersection(E, F)
	fmt.Printf("\nМножество C: %v\n", E)
	fmt.Printf("Множество D: %v\n", F)
	fmt.Printf("Пересечение: %v\n", result2) // []
}
