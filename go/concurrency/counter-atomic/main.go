// 学习目标：用 sync/atomic 的原子加修复 counter 的数据竞争。
// 运行（竞态检测）：go run -race examples/go/concurrency/counter-atomic/main.go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	const goroutines = 100
	const perGoroutine = 1000

	var counter atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Go(func() {
			for j := 0; j < perGoroutine; j++ {
				counter.Add(1)
			}
		})
	}

	wg.Wait()
	fmt.Println("expected:", goroutines*perGoroutine)
	fmt.Println("actual:  ", counter.Load())
}
