package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	http.DefaultClient.Do(req)
	resultCh := make(chan string)
	var wg sync.WaitGroup

	wg.Go(func() {
		defer fmt.Println("查询 goroutine 结束")
		// 用 select 同时等待发送结果和 ctx.Done()。
		// 如果收到取消信号，打印“查询已取消”。
		select {
		case <-time.After(2 * time.Second):
			resultCh <- "查询成功"
			fmt.Println("查询完毕")
		case <-ctx.Done():
			fmt.Println("查询已取消")
		}

	})

	select {
	case result := <-resultCh:
		fmt.Println(result)
	case <-time.After(1 * time.Second):
		fmt.Println("查询超时，发送取消信号")
		cancel()
	}

	wg.Wait()
}
