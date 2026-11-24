package main

import (
	"context"
	"fmt"
	"time"
)

func query(ctx context.Context) (string, error) {
	// 使用 select 等待：
	// 1. 2 秒后模拟查询成功，返回结果和 nil。
	// 2. ctx 被取消时，返回空字符串和 ctx.Err()。
	select {
	case <-time.After(2 * time.Second):
		return "查询成功", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := query(ctx)
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Println("查询结果:", result)
}
