package main

import (
	"fmt"
	"time"
)

//Разработать программу, которая будет последовательно отправлять значения в канал, а с другой стороны канала – читать эти значения.
//По истечении N секунд программа должна завершаться.
//Подсказка: используйте time.After или таймер для ограничения времени работы.

func sendMessagesWithTimer(sec int) {
	ch := make(chan int)

	// send messages to channel
	go func() {
		for i := 0; ; i++ {
			ch <- i
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// receive messages from channel
	go func() {
		for val := range ch {
			fmt.Println("Received message:", val)
		}
	}()

	// wait for timer and close channel
	<-time.After(time.Duration(sec) * time.Second)
	fmt.Println("Time is out, stopping program")
	close(ch)
}

func main() {
	sendMessagesWithTimer(3)
}
