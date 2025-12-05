package main

import (
	"testing"
	"time"
)

// TestOrEmpty тестирует функцию or с пустым списком каналов
func TestOrEmpty(t *testing.T) {
	done := or()

	// Канал должен быть закрытым и готовым к чтению
	select {
	case _, ok := <-done:
		if ok {
			t.Error("канал должен быть закрыт")
		}
	case <-time.After(1 * time.Second):
		t.Error("ожидание истекло, канал не закрыт")
	}
}

// TestOrSingleChannel тестирует функцию or с одним каналом
func TestOrSingleChannel(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(sig(100 * time.Millisecond))
	elapsed := time.Since(start)

	// Должно закрыться примерно через 100ms
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("ожидаемое время ~100ms, получено %v", elapsed)
	}
}

// TestOrMultipleChannels тестирует функцию or с несколькими каналами
func TestOrMultipleChannels(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	tests := []struct {
		name         string
		durations    []time.Duration
		expectedTime time.Duration
		tolerance    time.Duration
	}{
		{
			name:         "fastest channel wins",
			durations:    []time.Duration{2 * time.Second, 500 * time.Millisecond, 100 * time.Millisecond},
			expectedTime: 100 * time.Millisecond,
			tolerance:    50 * time.Millisecond,
		},
		{
			name:         "single fastest in many",
			durations:    []time.Duration{2 * time.Hour, 5 * time.Minute, 1 * time.Second, 1 * time.Hour, 1 * time.Minute},
			expectedTime: 1 * time.Second,
			tolerance:    100 * time.Millisecond,
		},
		{
			name:         "all same duration",
			durations:    []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			expectedTime: 100 * time.Millisecond,
			tolerance:    50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels := make([]<-chan interface{}, len(tt.durations))
			for i, duration := range tt.durations {
				channels[i] = sig(duration)
			}

			start := time.Now()
			<-or(channels...)
			elapsed := time.Since(start)

			// Проверяем, что время выполнения близко к времени самого быстрого канала
			if elapsed < tt.expectedTime-tt.tolerance || elapsed > tt.expectedTime+tt.tolerance {
				t.Errorf("ожидаемое время %v±%v, получено %v", tt.expectedTime, tt.tolerance, elapsed)
			}
		})
	}
}

// TestOrChannelClosure тестирует, что возвращаемый канал корректно закрывается
func TestOrChannelClosure(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	done := or(
		sig(1*time.Second),
		sig(100*time.Millisecond),
		sig(500*time.Millisecond),
	)

	select {
	case val, ok := <-done:
		if ok {
			t.Error("канал должен быть закрыт, но получено значение:", val)
		}
		// ожидаемое поведение - канал закрыт
		return
	case <-time.After(2 * time.Second):
		t.Error("ожидание истекло")
	}

}

// TestOrLargeNumberOfChannels тестирует функцию с большим количеством каналов
func TestOrLargeNumberOfChannels(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	// Создаем 100 каналов, все кроме одного имеют длительность > 1 секунды
	channels := make([]<-chan interface{}, 100)
	for i := 0; i < 100; i++ {
		if i == 50 {
			channels[i] = sig(50 * time.Millisecond)
		} else {
			channels[i] = sig(10 * time.Second)
		}
	}

	start := time.Now()
	<-or(channels...)
	elapsed := time.Since(start)

	// Должно завершиться за время самого быстрого канала (~50ms)
	tolerance := 10 * time.Millisecond
	if elapsed < 50*time.Millisecond || elapsed > 50*time.Millisecond+tolerance {
		t.Errorf("ожидаемое время ~50ms, получено %v", elapsed)
	}
}

// TestOrWithClosedChannel тестирует функцию с уже закрытым каналом
func TestOrWithClosedChannel(t *testing.T) {
	// Создаем уже закрытый канал
	c := make(chan interface{})
	close(c)

	start := time.Now()
	<-or(c)
	elapsed := time.Since(start)

	// Должно завершиться почти мгновенно
	if elapsed > 10*time.Millisecond {
		t.Errorf("функция отработала слишком долго: %v", elapsed)
	}
}
