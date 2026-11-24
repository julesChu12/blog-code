// 学习目标：对比「先检查再原子减」和「CAS 循环」在库存扣减上的差异。
// 运行：go run examples/go/concurrency/stock-oversell/main.go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	initialStock = 100
	goroutines   = 1000
)

// naive 只保证单次 Add 原子，检查与扣减之间仍有空窗，会超卖。
func naive(stock, sold *atomic.Int64, start *sync.WaitGroup) {
	start.Wait()
	if stock.Load() > 0 {
		runtime.Gosched() // 放大检查与扣减之间的窗口，模拟真实业务处理
		stock.Add(-1)
		sold.Add(1)
	}
}

// cas 把「检查 + 扣减」合成一个不可分割的 CAS 操作，冲突时重试。
func cas(stock, sold *atomic.Int64, start *sync.WaitGroup) {
	start.Wait()
	for {
		old := stock.Load()
		if old <= 0 {
			return
		}
		if stock.CompareAndSwap(old, old-1) {
			sold.Add(1)
			return
		}
	}
}

func run(name string, fn func(*atomic.Int64, *atomic.Int64, *sync.WaitGroup)) {
	var stock atomic.Int64
	stock.Store(initialStock)
	var sold atomic.Int64

	// 用 ready 做发令枪，尽量让所有 goroutine 同时开始抢库存。
	var ready sync.WaitGroup
	ready.Add(goroutines)
	var done sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		done.Go(func() {
			ready.Done()
			fn(&stock, &sold, &ready)
		})
	}

	done.Wait()
	fmt.Printf("%-6s sold=%-4d stock=%d\n", name, sold.Load(), stock.Load())
}

func main() {
	run("naive", naive)
	run("cas", cas)
}
