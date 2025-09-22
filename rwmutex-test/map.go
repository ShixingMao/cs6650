package main

import (
	"fmt"
	"sync"
	"time"
)

type Container struct {
	mu sync.RWMutex
	m  map[int]int
}

func (c *Container) write(key, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}

func (c *Container) length() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

func main() {
	c := Container{
		m: make(map[int]int),
	}
	var wg sync.WaitGroup

	start := time.Now()

	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				c.write(g*1000+i, i)
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Println("len(m):", c.length())
	fmt.Println("Time taken:", elapsed)
}
