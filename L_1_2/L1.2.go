package main

import (
	"fmt"
	"sync"
)

//Написать программу, которая конкурентно рассчитает значения квадратов чисел, взятых из массива [2,4,6,8,10], и выведет результаты в stdout.
//Подсказка: запусти несколько горутин, каждая из которых возводит число в квадрат.

func calcSqrtInArray(numbers []int) {
	sqrtArr := make([]int, len(numbers))

	var wg sync.WaitGroup
	for idx, num := range numbers {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sqrtArr[idx] = n * n
		}(num)
	}

	wg.Wait()

	for _, square := range sqrtArr {
		fmt.Printf("%d ", square)
	}
}

func main() {

	numbers := []int{2, 4, 6, 8, 10}
	calcSqrtInArray(numbers)
}
