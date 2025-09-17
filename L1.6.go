package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 1. Остановка по условию (флагу)
func workerByCondition(stop *bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if *stop {
			fmt.Println("Worker: Выход по условию")
			return
		}
		fmt.Println("Worker: Работаю...")
		time.Sleep(500 * time.Millisecond)
	}
}

func stopByCondition() {
	fmt.Println("--- Демонстрация остановки по условию ---")
	var wg sync.WaitGroup
	stop := false

	wg.Add(1)
	go workerByCondition(&stop, &wg)

	time.Sleep(2 * time.Second)
	stop = true // Меняем условие остановки

	wg.Wait()
	fmt.Println("Горутина остановлена.")
}

// 2. Остановка через канал
func workerWithChannel(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-stopCh:
			fmt.Println("Worker: Получен сигнал из канала. Выход.")
			return
		default:
			fmt.Println("Worker: Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func stopWithChannel() {
	fmt.Println("--- Демонстрация остановки через канал ---")
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	wg.Add(1)
	go workerWithChannel(stopCh, &wg)

	time.Sleep(2 * time.Second)
	close(stopCh) // Закрываем канал, горутина получит сигнал и завершится

	wg.Wait()
	fmt.Println("Main: Горутина остановлена.")

}

// 3. Остановка через контекст
func workerWithContext(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done(): // Канал Done() закроется при отмене контекста
			fmt.Println("Worker: Контекст отменен. Выход.")
			return
		default:
			fmt.Println("Worker (context): Работаю...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func stopWithContext() {
	fmt.Println("--- Демонстрация остановки через контекст ---")
	var wg sync.WaitGroup
	// Создаем контекст с функцией отмены
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go workerWithContext(ctx, &wg)

	time.Sleep(2 * time.Second)
	cancel() // Отменяем контекст

	wg.Wait()
	fmt.Println("Main: Горутина остановлена.")
}

// 4. Остановка с помощью runtime.Goexit()
func workerWithGoexit(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Worker: Работаю...")
	time.Sleep(1 * time.Second)
	fmt.Println("Worker: Вызываю runtime.Goexit().")
	runtime.Goexit() // Немедленно завершает текущую горутину
	fmt.Println("Worker (Goexit): Эта строка не будет напечатана.")
}

func stopWithGoexit() {
	fmt.Println("--- Демонстрация остановки с runtime.Goexit() ---")
	var wg sync.WaitGroup
	wg.Add(1)
	go workerWithGoexit(&wg)
	wg.Wait() // Ожидаем завершения горутины
	fmt.Println("Main: Горутина остановлена.")
}

func stoppingGoroutineDemo() {
	stopByCondition()
	fmt.Println()
	stopWithChannel()
	fmt.Println()
	stopWithContext()
	fmt.Println()
	stopWithGoexit()
}
