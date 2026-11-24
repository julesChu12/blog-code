// 学习目标：演示循环里使用 defer Unlock 导致跨迭代持锁，第二次 Lock 死锁。
// 运行：go run examples/go/concurrency/lock-loop-defer/main.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex

	for i := 0; i < 2; i++ {
		mu.Lock()
		defer mu.Unlock() // 函数返回时才解锁，第一轮结束后锁仍被持有
		fmt.Println("iteration", i)
	}
}
