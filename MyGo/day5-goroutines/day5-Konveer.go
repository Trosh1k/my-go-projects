package main

import (
	"fmt"
	"sync"
)

func generate(out chan<- int, n int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= n; i++ {
		out <- i
	}
	close(out)
}

func square(in <-chan int, out chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range in {
		out <- v * v
	}
	close(out)
}

func printer(in <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range in {
		fmt.Println(v)
	}
}

func main() {
	const n = 10

	ch1 := make(chan int)
	ch2 := make(chan int)

	var wg sync.WaitGroup
	wg.Add(3)

	go generate(ch1, n, &wg)
	go square(ch1, ch2, &wg)
	go printer(ch2, &wg)

	wg.Wait()
	fmt.Println("Конвейер завершён.")
}
