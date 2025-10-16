package main

import "fmt"

//Реализовать алгоритм быстрой сортировки массива встроенными средствами языка. Можно использовать рекурсию.
//
//Подсказка: напишите функцию quickSort([]int) []int которая сортирует срез целых чисел. Для выбора опорного элемента можно взять середину или первый элемент.

func quickSort(a []int) []int {
	if len(a) < 2 {
		return a
	}

	pivot := a[0] // опорный элемент
	var less []int
	var greater []int

	for _, num := range a[1:] {
		if num <= pivot {
			less = append(less, num)
		} else {
			greater = append(greater, num)
		}
	}

	less = quickSort(less)
	greater = quickSort(greater)

	return append(append(less, pivot), greater...)
}

func main() {
	data := []int{3, 6, 8, 10, 1, 2, 1}
	fmt.Println("Unsorted:", data)
	sorted := quickSort(data)
	fmt.Println("Sorted:  ", sorted)
}
