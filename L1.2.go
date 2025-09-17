package main

import (
	"fmt"
	"sync"
)

func CalcSqrtInArray(numbers []int) {
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
