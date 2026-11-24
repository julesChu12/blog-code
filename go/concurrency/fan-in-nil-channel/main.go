// 学习目标：在实际的 fan-in 场景里理解「把分支设为 nil 来禁用」。
// 三个下游各自返回一串结果，汇总到一个 channel；某个下游关闭后，
// 必须把它的分支置 nil，否则 select 会一直从已关闭 channel 读到零值。
// 运行：go run examples/go/concurrency/fan-in-nil-channel/main.go
package main

import (
	"fmt"
	"sync"
	"time"
)

// producer 模拟一个下游服务：每隔 delay 返回一个结果，发完 N 条后关闭。
func producer(name string, n int, delay time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := 1; i <= n; i++ {
			time.Sleep(delay)
			ch <- fmt.Sprintf("%s-%d", name, i)
		}
	}()
	return ch
}

// merge 把三个上游合并成一个输出流。
func merge(a, b, c <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for a != nil || b != nil || c != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil // 关闭后禁用该分支
					continue
				}
				out <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				out <- v
			case v, ok := <-c:
				if !ok {
					c = nil
					continue
				}
				out <- v
			}
		}
	}()
	return out
}

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Go(func() {
		for v := range merge(
			producer("A", 2, 30*time.Millisecond),
			producer("B", 3, 20*time.Millisecond),
			producer("C", 1, 50*time.Millisecond),
		) {
			fmt.Printf("%-6s got %s\n", time.Since(start).Round(10*time.Millisecond), v)
		}
	})
	wg.Wait()
	fmt.Println("all sources closed")
}
