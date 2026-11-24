package main

import (
	"fmt"
	"time"
)

func main() {
	resultCh := make(chan string)

	go func() {
		defer func() {
			recover()
			fmt.Println("goroutine 结束")
		}()
		// 模拟一次耗时的下游查询。
		time.Sleep(2 * time.Second)
		fmt.Println("开始查询")
		resultCh <- "查询成功"
		fmt.Println("查询完毕")
	}()

	// 使用 select：收到结果时打印结果，等待 1 秒后打印“查询超时”。
	// 超时可以使用 <-time.After(1 * time.Second)。
	select {
	case result := <-resultCh:
		fmt.Println(result)
	case <-time.After(1 * time.Second):
		fmt.Println("查询超时")
	}
	time.Sleep(3 * time.Second)
}
