package main

import (
	"fmt"
)

func worker(id int, jobs <-chan int) {
	for j := range jobs {
		fmt.Printf("Worker %d: %d\n", id, j)
	}
}

func PrintingWorkers(workerCount int) {
	if workerCount <= 0 {
		fmt.Println("Error: Invalid worker count")
		return
	}

	dataChannel := make(chan int)

	for i := 1; i <= workerCount; i++ {
		go worker(i, dataChannel)
		fmt.Printf("Worker #%d launched\n", i)
	}

	fmt.Println("--- Sending data to channel ---")
	for i := 0; ; i++ {
		dataChannel <- i
	}
}
