package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func unbufferedWrite(filename string, n int) time.Duration {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	start := time.Now()
	for i := 0; i < n; i++ {
		_, err := f.Write([]byte(fmt.Sprintf("line %d\n", i)))
		if err != nil {
			panic(err)
		}
	}
	return time.Since(start)
}

func bufferedWrite(filename string, n int) time.Duration {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	start := time.Now()
	for i := 0; i < n; i++ {
		_, err := w.WriteString(fmt.Sprintf("line %d\n", i))
		if err != nil {
			panic(err)
		}
	}
	w.Flush()
	return time.Since(start)
}

func main() {
	const n = 100000
	unbufTime := unbufferedWrite("unbuffered.txt", n)
	bufTime := bufferedWrite("buffered.txt", n)

	fmt.Println("Unbuffered write time:", unbufTime)
	fmt.Println("Buffered write time:", bufTime)
}
