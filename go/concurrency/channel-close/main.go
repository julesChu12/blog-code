// 学习目标：验证 close(channel) 的语义：缓冲区剩余数据、重复关闭、关闭 nil、向已关闭发送。
// 运行：go run examples/go/concurrency/channel-close/main.go
package main

import "fmt"

// try 执行 f，并捕获 panic，用于对比不同非法操作的运行时错误。
func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%-24s panic: %v\n", name, r)
		}
	}()
	f()
	fmt.Printf("%-24s no panic\n", name)
}

func main() {
	// 1. 有缓冲 channel 关闭后，缓冲区里的数据仍然可以读出来。
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	for i := 1; i <= 4; i++ {
		v, ok := <-ch
		fmt.Printf("receive #%d: value=%d ok=%t\n", i, v, ok)
	}

	// 2. 缓冲区读完后，range 会正常结束。
	ch2 := make(chan int, 2)
	ch2 <- 10
	ch2 <- 20
	close(ch2)
	for v := range ch2 {
		fmt.Printf("range: %d\n", v)
	}

	// 3. 非法操作：关闭 nil、重复关闭、向已关闭 channel 发送。
	try("close nil channel", func() {
		var nilCh chan int
		close(nilCh)
	})

	closed := make(chan int)
	close(closed)
	try("close closed channel", func() {
		close(closed)
	})
	try("send on closed channel", func() {
		closed <- 1
	})

	// 4. 阻塞中的发送者被 close 唤醒后 panic，而不是在 close 当场 panic。
	unbuffered := make(chan int)
	result := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Sprint(r)
			}
		}()
		unbuffered <- 1
	}()

	close(unbuffered)
	fmt.Printf("%-24s panic: %s\n", "blocked sender", <-result)
}
