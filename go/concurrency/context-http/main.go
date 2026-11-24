package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("下游：收到请求")
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprintln(w, "查询成功")
			fmt.Println("下游：查询完成")
		case <-r.Context().Done():
			fmt.Println("下游：请求已取消")
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 将下面的 NewRequest 改为 NewRequestWithContext，并传入 ctx。
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}
	fmt.Println("请求结果:", string(body))
}
