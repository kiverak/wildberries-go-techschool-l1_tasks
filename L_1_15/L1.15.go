package main

import (
	"fmt"
	"strings"
)

//Рассмотреть следующий код и ответить на вопросы: к каким негативным последствиям он может привести и как это исправить?
//
//Приведите корректный пример реализации.
//
//var justString string
//
//func someFunc() {
//	v := createHugeString(1 &lt;&lt; 10)
//	justString = v[:100]
//}
//
//func main() {
//	someFunc()
//}
//Вопрос: что происходит с переменной justString?

//Переменная justString содержит подстроку из первых 100 символов большой строки, но может удерживать в памяти всю исходную
//большую строку, потому что в Go срез строки разделяет внутренний буфер с оригиналом. Это приводит к неоправданному расходу
//памяти (утечке). Чтобы избежать этого, нужно явно создать копию подстроки.

func createHugeString(size int) string {
	var builder strings.Builder
	builder.Grow(size)
	for i := 0; i < size; i++ {
		builder.WriteString("a")
	}
	return builder.String()
}

var justString string

func someFunc() {
	v := createHugeString(1 << 10)
	justString = strings.Clone(v[:100])

	v = ""
	fmt.Print(justString)
}

func main() {
	someFunc()
}
