// 学习目标：对比“不可中断的 Sleep”和“select 监听取消”的退出时机。
// 运行：go run examples/go/concurrency/context-responsive/main.go
package main

import (
	"context"
	"fmt"
	"time"
)

type result struct {
	name      string
	elapsed   time.Duration
	cancelled bool
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan result, 2)
	start := time.Now()

	// worker A：直接 sleep，取消信号无法打断它。
	go func() {
		time.Sleep(2 * time.Second)
		results <- result{"sleep-only", time.Since(start), ctx.Err() != nil}
	}()

	// worker B：用 select 同时等待“工作完成”和“取消信号”。
	go func() {
		select {
		case <-time.After(2 * time.Second):
			results <- result{"select-cancel", time.Since(start), false}
		case <-ctx.Done():
			results <- result{"select-cancel", time.Since(start), true}
		}
	}()

	// 给两个 worker 一点时间进入阻塞，然后取消。
	time.Sleep(50 * time.Millisecond)
	cancel()

	for i := 0; i < 2; i++ {
		r := <-results
		fmt.Printf("%-13s elapsed=%-10s cancelled=%t\n", r.name, r.elapsed.Round(time.Millisecond), r.cancelled)
	}
}
