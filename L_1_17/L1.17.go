package main

import (
	"fmt"
)

//Реализовать алгоритм бинарного поиска встроенными методами языка. Функция должна принимать отсортированный слайс и
//искомый элемент, возвращать индекс элемента или -1, если элемент не найден.
//
//Подсказка: можно реализовать рекурсивно или итеративно, используя цикл for.

func binarySearch(data []int, target int) int {
	left, right := 0, len(data)-1

	for left <= right {
		// Находим середину, избегая возможного переполнения (left+right)/2
		mid := left + (right-left)/2

		if data[mid] == target {
			return mid // Элемент найден
		}

		if data[mid] < target {
			left = mid + 1 // Искать в правой половине
		} else {
			right = mid - 1 // Искать в левой половине
		}
	}

	return -1 // Элемент не найден
}

func main() {
	sorted := []int{1, 1, 2, 3, 6, 8, 10}
	fmt.Println("Sorted:  ", sorted)

	targetExists := 6
	targetMissing := 99

	fmt.Printf("Searching for %d... Index: %d\n", targetExists, binarySearch(sorted, targetExists))
	fmt.Printf("Searching for %d... Index: %d\n", targetMissing, binarySearch(sorted, targetMissing))
}
