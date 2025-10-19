package main

import (
	"fmt"
	"sync"
)

//Реализовать структуру-счётчик, которая будет инкрементироваться в конкурентной среде (т.е. из нескольких горутин).
//По завершению программы структура должна выводить итоговое значение счётчика.
//Подсказка: вам понадобится механизм синхронизации, например, sync.Mutex или sync/Atomic для безопасного инкремента.

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) increment() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func main() {
	var wg sync.WaitGroup
	counter := Counter{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.increment()
		}()
	}

	wg.Wait()

	fmt.Printf("Final counter value: %d\n", counter.value)
}
