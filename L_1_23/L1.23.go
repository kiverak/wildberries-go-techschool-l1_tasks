package main

import "fmt"

//Удалить i-ый элемент из слайса. Продемонстрируйте корректное удаление без утечки памяти.
//
//Подсказка: можно сдвинуть хвост слайса на место удаляемого элемента (copy(slice[i:], slice[i+1:])) и уменьшить длину слайса на 1.

// removeByIndex удаляет элемент из среза по индексу i
func removeByIndex[T any](slice []T, i int) []T {
	if i < 0 || i >= len(slice) {
		return slice // Возвращаем исходный срез, если индекс некорректен
	}

	copy(slice[i:], slice[i+1:])

	// Обнуляем последний элемент среза
	var zero T
	slice[len(slice)-1] = zero
	// Обрезаем срез
	return slice[:len(slice)-1]
}

func main() {
	dataInt := []int{0, 1, 2, 3, 4, 5, 6}
	fmt.Printf("Original int slice: %v, len=%d, cap=%d\n", dataInt, len(dataInt), cap(dataInt))

	// Удаляем элемент с индексом 3
	dataInt = removeByIndex(dataInt, 3)
	fmt.Printf("After removing [3]: %v, len=%d, cap=%d\n\n", dataInt, len(dataInt), cap(dataInt))
}
