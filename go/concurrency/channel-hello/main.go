package main

import "fmt"

func main() {
	ch := make(chan string)

	go func() {
		// ① 把 "hello" 发送到 ch。
		ch <- "hello"
	}()

	// ② 从 ch 接收一个值，赋给 message。
	message := <-ch
	fmt.Println(message)
	close(ch)
}
