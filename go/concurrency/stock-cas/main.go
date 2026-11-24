// 学习目标：把「先检查再扣减」改写成 CAS 循环，消除 check-then-act 竞态。
// 运行（竞态检测）：go run -race examples/go/concurrency/stock-cas/main.go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	initialStock = 100
	goroutines   = 1000
)

func main() {
	var stock atomic.Int64
	stock.Store(initialStock)
	var sold atomic.Int64
	var retries atomic.Int64 // 统计 CAS 失败重试次数

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Go(func() {
			for {
				old := stock.Load()
				if old <= 0 {
					return // 没库存，放弃
				}
				if stock.CompareAndSwap(old, old-1) {
					sold.Add(1) // 抢到了
					return
				}
				retries.Add(1) // 别人先改了，重读重试
			}
		})
	}

	wg.Wait()
	fmt.Printf("sold=%d stock=%d retries=%d\n", sold.Load(), stock.Load(), retries.Load())
}
