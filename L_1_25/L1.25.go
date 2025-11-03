package main

import (
	"fmt"
	"time"
)

//Реализовать собственную функцию sleep(duration) аналогично встроенной функции time.Sleep, которая приостанавливает выполнение текущей горутины.
//
//Важно: в отличие от настоящей time.Sleep, ваша функция должна именно блокировать выполнение (например, через таймер или цикл),
//а не просто вызывать time.Sleep :) — это упражнение.
//
//Можно использовать канал + горутину, или цикл на проверку времени (не лучший способ, но для обучения)

// sleepV1 - реализация с использованием таймера и канала
func sleepV1(d time.Duration) {
	timer := time.NewTimer(d)
	<-timer.C
}

// sleepV2 - реализация через "busy-wait" (активное ожидание)
func sleepV2(d time.Duration) {
	start := time.Now()
	for {
		if time.Since(start) >= d {
			break
		}
	}
}

func main() {
	fmt.Println("\n--- Тестируем sleepV1 ---")
	fmt.Println("Сейчас", time.Now().Format("15:04:05"))
	fmt.Println("Пауза 2 секунды...")
	sleepV1(2 * time.Second)
	fmt.Println("Прошло 2 секунды. Время:", time.Now().Format("15:04:05"))

	fmt.Println("\n--- Тестируем sleepV2 ---")
	fmt.Println("Сейчас", time.Now().Format("15:04:05"))
	fmt.Println("Пауза 2 секунды...")
	sleepV2(2 * time.Second)
	fmt.Println("Прошло 2 секунды. Время:", time.Now().Format("15:04:05"))
}
