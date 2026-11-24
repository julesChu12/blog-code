package main

import "sync"

func main() {
	services := []string{"user", "order", "inventory"}
	ch := make(chan string)
	var wg sync.WaitGroup
	// ① 为每个 service 启动一个 goroutine，向 ch 发送该 service。
	for _, service := range services {
		wg.Go(func() {
			ch <- service
		})
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	// ② 从 ch 接收 len(services) 次，每次打印收到的值。
	for service := range ch {
		println(service)
	}
}
