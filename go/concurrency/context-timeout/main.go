package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 等待 ctx 的取消信号，然后打印 ctx.Err()。
	<-ctx.Done()
	fmt.Println(ctx.Err())
}
