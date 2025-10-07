package main

import (
	"fmt"
)

//Имеется последовательность строк: ("cat", "cat", "dog", "cat", "tree"). Создать для неё собственное множество.
//
//Ожидается: получить набор уникальных слов. Для примера, множество = {"cat", "dog", "tree"}.

type MyStruct struct {
	uniqueSet map[string]struct{}
}

// NewMyStruct создаёт структуру с уникальными значениями
func NewMyStruct(sequence []string) *MyStruct {
	ms := &MyStruct{
		uniqueSet: make(map[string]struct{}),
	}
	for _, item := range sequence {
		ms.uniqueSet[item] = struct{}{}
	}
	return ms
}

// Add добавляет элемент в множество
func (ms *MyStruct) Add(item string) {
	ms.uniqueSet[item] = struct{}{}
}

// Remove удаляет элемент из множества
func (ms *MyStruct) Remove(item string) bool {
	_, exists := ms.uniqueSet[item]
	if exists {
		delete(ms.uniqueSet, item)
		return true
	}
	return false
}

// Len возвращает количество уникальных элементов
func (ms *MyStruct) Len() int {
	return len(ms.uniqueSet)
}

// GetUniqueItems возвращает срез уникальных элементов
func (ms *MyStruct) GetUniqueItems() []string {
	items := make([]string, 0, len(ms.uniqueSet))
	for item := range ms.uniqueSet {
		items = append(items, item)
	}
	return items
}

func main() {
	sequence := []string{"cat", "cat", "dog", "cat", "tree"}
	myStruct := NewMyStruct(sequence)

	myStruct.Add("cat")
	myStruct.Add("dog")
	myStruct.Add("god")

	fmt.Printf("myStruct: %v\n", myStruct.GetUniqueItems())
	fmt.Println("myStruct length: ", myStruct.Len())

	myStruct.Remove("cat")
	fmt.Println("cat removed from myStruct")

	fmt.Printf("myStruct: %v\n", myStruct.GetUniqueItems())
	fmt.Println("myStruct length:", myStruct.Len())
}
