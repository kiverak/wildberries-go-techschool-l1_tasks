package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

/*
	To launch the program input: go run L1.4.go <worker number>
	e.g.: go run L1.4.go 5
	To stop the program press: Ctrl+C
*/

func workerWithCancelling(ctx context.Context, wg *sync.WaitGroup, id int, jobs <-chan int) {
	defer wg.Done()
	fmt.Printf("Worker #%d launched\n", id)

	for {
		select {
		// Cancel signal received
		case <-ctx.Done():
			fmt.Printf("Worker #%d received cancel signal\n", id)
			return

		// Worker gets new data from channel
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Channel is closed, worker #%d is stopping\n", id)
				return
			}
			fmt.Printf("Worker #%d finished job: %d\n", id, job)
		}
	}
}

func main1() {
	workerCount, err := strconv.Atoi(os.Args[1])
	if err != nil || workerCount <= 0 {
		fmt.Println("Error: Invalid worker count")
		return
	}

	// Create context with cancel function
	ctx, cancel := context.WithCancel(context.Background())

	// Create channel & wait group
	jobs := make(chan int)
	var wg sync.WaitGroup

	// Launch workers
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go workerWithCancelling(ctx, &wg, i, jobs)
	}

	// Launch goroutine for cancelling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		// Block while waiting for signal
		<-sigChan
		cancel()
	}()

	// Sending data to channel
producerLoop:
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Producer stop sending data to channel")
			break producerLoop
		case jobs <- i:
		}
	}

	close(jobs)
	wg.Wait()
}
