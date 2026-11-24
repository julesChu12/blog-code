// 学习目标：用 sync.Mutex 把 counter++ 变成互斥的临界区。
// 运行（竞态检测）：go run -race examples/go/concurrency/counter-mutex/main.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	const goroutines = 100
	const perGoroutine = 1000

	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Go(func() {
			for j := 0; j < perGoroutine; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	fmt.Println("expected:", goroutines*perGoroutine)
	fmt.Println("actual:  ", counter)
}
