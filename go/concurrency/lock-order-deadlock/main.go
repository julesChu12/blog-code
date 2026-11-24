// 学习目标：复现“交叉加锁”导致的死锁，并对比按账户 id 排序的修复版。
// 复现死锁：go run examples/go/concurrency/lock-order-deadlock/main.go deadlock
// 运行修复版：go run examples/go/concurrency/lock-order-deadlock/main.go fixed
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type account struct {
	id      int
	mu      sync.Mutex
	balance int
}

// transferUnordered 按调用方给的顺序加锁：两条方向相反的转账会互相等待。
func transferUnordered(from, to *account, amount int) {
	from.mu.Lock()
	defer from.mu.Unlock()
	time.Sleep(50 * time.Millisecond) // 放大窗口：让另一边先拿到第一把锁
	to.mu.Lock()
	defer to.mu.Unlock()

	from.balance -= amount
	to.balance += amount
}

// transferOrdered 所有调用方都按账户 id 从小到大加锁，消除循环等待。
func transferOrdered(from, to *account, amount int) {
	first, second := from, to
	if first.id > second.id {
		first, second = second, first
	}

	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	from.balance -= amount
	to.balance += amount
}

func main() {
	mode := "deadlock"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	a := &account{id: 1, balance: 100}
	b := &account{id: 2, balance: 100}

	var wg sync.WaitGroup
	wg.Add(2)

	switch mode {
	case "deadlock":
		fmt.Println("交叉加锁：a->b 先锁 a，b->a 先锁 b")
		go func() { defer wg.Done(); transferUnordered(a, b, 10) }()
		go func() { defer wg.Done(); transferUnordered(b, a, 10) }()
	case "fixed":
		go func() { defer wg.Done(); transferOrdered(a, b, 10) }()
		go func() { defer wg.Done(); transferOrdered(b, a, 10) }()
	default:
		fmt.Println("usage: go run examples/go/concurrency/lock-order-deadlock/main.go deadlock|fixed")
		return
	}

	wg.Wait()
	fmt.Printf("a=%d b=%d total=%d\n", a.balance, b.balance, a.balance+b.balance)
}
