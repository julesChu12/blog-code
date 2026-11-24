// 学习目标：用 sync.Mutex 保护「两个账户之间的转账」，理解为什么必须固定加锁顺序。
// 运行：go run examples/go/concurrency/transfer-mutex/main.go
// 竞态检测：go run -race examples/go/concurrency/transfer-mutex/main.go
package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

type Account struct {
	id      int
	mu      sync.Mutex
	balance int
}

func NewAccount(id, balance int) *Account {
	return &Account{id: id, balance: balance}
}

// transfer 从 from 转 amount 给 to，余额不足返回 false。
//
// 每笔转账要同时改两个账户，所以持有两把锁。问题在于：
//   - goroutine 1 执行 A->B：先锁 A 再锁 B
//   - goroutine 2 执行 B->A：先锁 B 再锁 A
//
// 两边各持一把、等另一把，就是死锁。
//
// 解法：所有 goroutine 用同一套顺序拿锁——按账户 id 从小到大。
// 这样后拿到第一把锁的人一定排在同一个顺序里，不会形成环。
func transfer(from, to *Account, amount int) bool {
	first, second := from, to
	if first.id > second.id {
		first, second = second, first
	}

	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	if from.balance < amount {
		return false
	}
	from.balance -= amount
	to.balance += amount
	return true
}

func total(accounts []*Account) int {
	sum := 0
	for _, a := range accounts {
		a.mu.Lock()
		sum += a.balance
		a.mu.Unlock()
	}
	return sum
}

func main() {
	const (
		accountCount = 10
		initial      = 1000
		workers      = 100
		perWorker    = 200
	)

	accounts := make([]*Account, accountCount)
	// 初始化账户
	for i := range accounts {
		accounts[i] = NewAccount(i, initial)
	}
	// 计算初始总额
	want := accountCount * initial
	fmt.Println("initial total:", total(accounts))

	var wg sync.WaitGroup
	// 启动 workers 个 goroutine，每个 goroutine 做 perWorker 次随机转账
	for w := range workers {
		wg.Go(func() {
			// rand.NewPCG(seed, 1) 是一个并发安全的随机数生成器，使用 PCG 算法。
			r := rand.New(rand.NewPCG(uint64(w), 1))
			// 每个 goroutine 做 perWorker 次随机转账
			for range perWorker {
				from := accounts[r.IntN(len(accounts))]
				to := accounts[r.IntN(len(accounts))]
				if from == to {
					continue
				}
				// 随机转账 1~50 元
				transfer(from, to, 1+r.IntN(50))
			}
		})
	}
	wg.Wait()

	got := total(accounts)
	fmt.Println("final total:  ", got)
	fmt.Println("invariant:    ", got == want)
	for _, a := range accounts {
		if a.balance < 0 {
			fmt.Printf("negative balance on account %d: %d\n", a.id, a.balance)
		}
	}
}
