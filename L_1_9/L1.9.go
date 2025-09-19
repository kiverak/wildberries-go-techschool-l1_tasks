package main

import "fmt"

//Разработать конвейер чисел. Даны два канала: в первый пишутся числа x из массива, во второй – результат операции x*2.
//После этого данные из второго канала должны выводиться в stdout. То есть, организуйте конвейер из двух этапов с горутинами:
//генерация чисел и их обработка. Убедитесь, что чтение из второго канала корректно завершается.

// generator принимает срез чисел и отправляет их в канал.
func generator(nums []int, out chan<- int) {
	defer close(out)
	for _, n := range nums {
		out <- n
	}
}

// processor читает числа из входного канала, умножает их на 2 и отправляет результат в выходной канал.
func processor(in <-chan int, out chan<- int) {
	defer close(out)
	for n := range in {
		out <- n * 2
	}
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ch1 := make(chan int)
	ch2 := make(chan int)

	go generator(nums, ch1)
	go processor(ch1, ch2)

	// Основная горутина читает данные из второго канала, пока он не будет закрыт.
	for result := range ch2 {
		fmt.Println(result)
	}
}
