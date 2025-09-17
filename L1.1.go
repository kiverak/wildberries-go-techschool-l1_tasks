package main

import "fmt"

// Структура
type Human struct {
	Name string
	Age  int
}

// Метод структуры
func (h *Human) SayHello() {
	fmt.Printf("Привет, меня зовут %s, мне %d лет.\n", h.Name, h.Age)
}

// Дочерняя структура
type Action struct {
	Human    // Встроенная структура
	Activity string
}

// Метод дочерней структуры
func (a *Action) DoSomething() {
	fmt.Printf("%s сейчас %s.\n", a.Name, a.Activity)
}
