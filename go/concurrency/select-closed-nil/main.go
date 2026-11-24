// 学习目标：确认 select 遇到「已关闭 channel」和「nil channel」时的行为。
// 运行：go run examples/go/concurrency/select-closed-nil/main.go
package main

import (
	"fmt"
	"sync"
)

func main() {
	// 1. 已关闭的 channel：接收分支永远立即就绪，拿到零值和 ok=false。
	closed := make(chan int)
	close(closed)
	select {
	case v, ok := <-closed:
		fmt.Printf("closed channel: v=%d ok=%t\n", v, ok)
	default:
		fmt.Println("closed channel: default（不会执行）")
	}

	// 2. nil channel：该分支永远不就绪，select 只能走 default。
	var nilCh chan int
	select {
	case v := <-nilCh:
		fmt.Println("nil channel: 不会执行", v)
	default:
		fmt.Println("nil channel: default")
	}

	// 3. 工程用法：把不需要的分支变量设为 nil，即可禁用该分支。
	ch := make(chan int, 1)
	ch <- 42
	var disabled chan int // nil：暂时不关心这个分支
	select {
	case v := <-ch:
		fmt.Printf("enabled branch: %d\n", v)
	case <-disabled:
		fmt.Println("disabled branch（不会执行）")
	}

	// 4. close 的广播效果：一次 close，所有阻塞在接收上的 goroutine 同时被唤醒。
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			<-done
			fmt.Printf("worker %d woken\n", i)
		})
	}
	close(done)
	wg.Wait()
}
