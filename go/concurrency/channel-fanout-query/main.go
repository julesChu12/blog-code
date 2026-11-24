// 学习目标：并发查询三个下游，每个最多等 80ms，用 channel 汇总结果。
// 运行：go run examples/go/concurrency/channel-fanout-query/main.go
package main

import (
	"fmt"
	"sync"
	"time"
)

type result struct {
	service string
	value   string
	err     error
}

// query 模拟一次下游查询：latency 内返回成功，否则在 timeout 后放弃。
func query(service string, latency, timeout time.Duration) result {
	select {
	case <-time.After(latency):
		return result{service: service, value: service + " 查询成功"}
	case <-time.After(timeout):
		return result{service: service, err: fmt.Errorf("超过 %s", timeout)}
	}
}

func main() {
	downstreams := []struct {
		name    string
		latency time.Duration
	}{
		{"user", 20 * time.Millisecond},
		{"stock", 200 * time.Millisecond}, // 慢下游
		{"coupon", 40 * time.Millisecond},
	}

	results := make(chan result)
	var wg sync.WaitGroup
	start := time.Now()

	for _, d := range downstreams {
		wg.Go(func() {
			results <- query(d.name, d.latency, 80*time.Millisecond)
		})
	}

	// 所有发送者结束后关闭 results，range 才能正常退出。
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		elapsed := time.Since(start).Round(10 * time.Millisecond)
		if r.err != nil {
			fmt.Printf("%-6s %-6s %v\n", elapsed, r.service, r.err)
			continue
		}
		fmt.Printf("%-6s %-6s %s\n", elapsed, r.service, r.value)
	}
	fmt.Println("全部下游结束")
}
