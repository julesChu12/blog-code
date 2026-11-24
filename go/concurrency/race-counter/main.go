// 学习目标：观察多个 goroutine 对同一个变量做 counter++ 时的数据竞争。
// 运行（普通）：go run examples/go/concurrency/race-counter/main.go
// 运行（竞态检测）：go run -race examples/go/concurrency/race-counter/main.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	const goroutines = 100
	const perGoroutine = 1000

	var counter int
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Go(func() {
			for j := 0; j < perGoroutine; j++ {
				counter++ // 读 counter、加一、写回；三步不是原子的
			}
		})
	}

	wg.Wait()
	fmt.Println("expected:", goroutines*perGoroutine)
	fmt.Println("actual:  ", counter)
}
