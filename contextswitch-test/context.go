package main

import (
	"fmt"
	"runtime"
	"time"
)

func pingPong(rounds int) time.Duration {
	ch := make(chan struct{})
	done := make(chan struct{})

	go func() {
		for i := 0; i < rounds; i++ {
			ch <- struct{}{}
			<-ch
		}
		done <- struct{}{}
	}()

	start := time.Now()
	for i := 0; i < rounds; i++ {
		<-ch
		ch <- struct{}{}
	}
	<-done
	elapsed := time.Since(start)
	return elapsed
}

func main() {
	const rounds = 1_000_000

	// Single OS thread
	runtime.GOMAXPROCS(1)
	elapsed1 := pingPong(rounds)
	avg1 := elapsed1.Seconds() / float64(2*rounds)
	fmt.Printf("GOMAXPROCS(1):   total=%v, avg switch=%.2fus\n", elapsed1, avg1*1e6)

	// Multiple OS threads (default)
	runtime.GOMAXPROCS(runtime.NumCPU())
	elapsed2 := pingPong(rounds)
	avg2 := elapsed2.Seconds() / float64(2*rounds)
	fmt.Printf("GOMAXPROCS(N):   total=%v, avg switch=%.2fus\n", elapsed2, avg2*1e6)
}
