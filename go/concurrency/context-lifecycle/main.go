package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	wallet, cancelWallet := context.WithCancel(parent)
	defer cancelWallet()
	stock, cancelStock := context.WithCancel(parent)
	defer cancelStock()

	observed := make(chan struct{})
	allowCleanup := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() {
		<-wallet.Done()
		close(observed)
		<-allowCleanup // 模拟收到取消后仍需完成清理
		fmt.Println("钱包任务：清理完成")
	})

	cancelStock()
	fmt.Println("库存取消:", stock.Err())
	fmt.Println("父节点仍有效:", parent.Err() == nil)
	fmt.Println("钱包仍有效:", wallet.Err() == nil)

	cancelParent()
	<-observed
	fmt.Println("钱包任务：已观察到取消，清理尚未完成")
	close(allowCleanup)
	wg.Wait()
	fmt.Println("所有任务已结束")

	cause := errors.New("库存查询超时")
	timed, cancelTimed := context.WithTimeoutCause(context.Background(), 10*time.Millisecond, cause)
	defer cancelTimed()
	<-timed.Done()
	fmt.Println("超时状态:", timed.Err())
	fmt.Println("超时原因:", context.Cause(timed))
}
