package main

import "fmt"

func main() {
	userCh := make(chan string, 1)
	orderCh := make(chan string, 1)

	userCh <- "user result"
	orderCh <- "order result"

	// 用 select 从两个 channel 中接收一个值并打印。
	for {
		select {
		case userResult, ok := <-userCh:
			if !ok {
				userCh = nil // 关闭 channel，避免再次接收。
				continue
			}
			fmt.Println(userResult)
		case orderResult, ok := <-orderCh:
			if !ok {
				orderCh = nil // 关闭 channel，避免再次接收。
				continue
			}
			fmt.Println(orderResult)

		default:
			// 如果两个 channel 都没有值可接收，退出循环。
			return
		}
	}

}
