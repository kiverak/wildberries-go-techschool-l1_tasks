package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

func writeDataToMapConcurrent() {
	const goroutines = 100 // число горутин
	const perG = 1000      // число записей на одну горутину

	var m sync.Map // concurrent map
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				key := strings.Join([]string{"k-", strconv.Itoa(g), strconv.Itoa(i)}, "-") // ключ вида k-55-555
				m.Store(key, g*i)
			}
		}(g)
	}

	wg.Wait()

	// Подсчёт размера мапы через Range, т.к. у sync.Map нет len()
	count := 0
	m.Range(func(key, value any) bool {
		count++
		return true
	})
	fmt.Println("len:", count) // должно быть 100.000
}

// запуск детектора гонок: go run -race main.go
// для запуска на Windows установить w64devkit
