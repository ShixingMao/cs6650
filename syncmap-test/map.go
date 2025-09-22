package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var m sync.Map
	var wg sync.WaitGroup
	start := time.Now()

	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				m.Store(g*1000+i, i) // concurrent write
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Count entries
	count := 0
	m.Range(func(key, value any) bool {
		count++
		return true
	})

	fmt.Println("len(m):", count)
	fmt.Println("Time taken:", elapsed)

}
