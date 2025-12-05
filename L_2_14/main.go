package main

import (
	"fmt"
	"sync"
	"time"
)

// Реализовать функцию, которая будет объединять один или более каналов done (каналов сигнала завершения) в один. Возвращаемый канал должен закрываться,
// как только закроется любой из исходных каналов.
//
// В этом примере канал, возвращённый or(...), закроется через ~1 секунду, потому что самый короткий канал sig(1*time.Second) закроется первым.
// Ваша реализация or должна уметь принимать на вход произвольное число каналов и завершаться при сигнале на любом из них.
//
// Подсказка: используйте select в бесконечном цикле для чтения из всех каналов одновременно, либо рекурсивно объединяйте каналы попарно.

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}

func or(channels ...<-chan interface{}) <-chan interface{} {
	orDone := make(chan interface{})

	// Если нет каналов, возвращаем закрытый канал
	if len(channels) == 0 {
		close(orDone)
		return orDone
	}

	var once sync.Once // чтобы гарантировать, что orDone закроется только один раз
	for _, ch := range channels {
		go func(c <-chan interface{}) {
			select {
			case <-c:
				once.Do(func() { close(orDone) })
			case <-orDone: // быстрая отмена ожидания остальных
			}
		}(ch)
	}
	return orDone
}

// or объединяет один или более каналов в один канал
// Возвращаемый канал закрывается, как только закроется любой из исходных каналов
func orRecursive(channels ...<-chan interface{}) <-chan interface{} {
	// Если нет каналов, возвращаем закрытый канал
	if len(channels) == 0 {
		c := make(chan interface{})
		close(c)
		return c
	}

	// Если один канал, возвращаем его как есть
	if len(channels) == 1 {
		return channels[0]
	}

	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		// Используем select для ожидания сигнала с одного из двух каналов:
		// 1. Первый исходный канал
		// 2. Результат объединения оставшихся каналов
		select {
		case <-channels[0]:
		case <-or(channels[1:]...):
		}
	}()
	return orDone
}
