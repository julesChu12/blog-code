// 学习目标：对比“锁的范围”对并发度的影响，并演示 Mutex 不可重入。
//
// 场景：10 个 worker 各自完成一次 50ms 的慢操作（模拟 RPC、日志、磁盘 I/O），
// 然后把共享计数器 n 加一。
//
//	wide      : 慢操作也在锁内，10 个 worker 串行，总耗时约 10 × 50ms。
//	narrow    : 只有 n++ 在锁内，慢操作并行，总耗时约 50ms。
//	recursive : 同一个 goroutine 重复 Lock 同一把锁，触发运行时死锁检测。
//
// 运行：
//
//	go run examples/go/concurrency/lock-scope/main.go wide
//	go run examples/go/concurrency/lock-scope/main.go narrow
//	go run examples/go/concurrency/lock-scope/main.go recursive
//
// 预期（Go 1.25.8、macOS arm64，耗时会有少量调度开销）：
//
//	mode=wide   workers=10 count=10 elapsed=510ms
//	mode=narrow workers=10 count=10 elapsed=50ms
//	recursive → fatal error: all goroutines are asleep - deadlock!
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	// workers 是并发执行慢操作的 goroutine 数量。
	workers = 10
	// slowWork 模拟一次不碰共享状态的慢操作；Sleep 不占 CPU，但会占用墙钟时间。
	slowWork = 50 * time.Millisecond
)

// stats 是多个 goroutine 共享的状态。
type stats struct {
	// mu 保护 n，相当于“记账本的那支笔”。
	mu sync.Mutex
	// n 是共享计数器；所有读写都必须经过 mu。
	n int
}

// incWide 把慢操作也放进临界区，锁的范围过大。
//
// 时间线（W 表示一个 worker）：
//
//	W1 [Lock][------ 50ms ------][n++][Unlock]
//	W2        [等待锁....................][Lock][------ 50ms ------]...
//	W3                 [等待锁.....................]...
//
// 锁被占用的 50ms 里，其他 worker 只能排队，10 个 worker 约 500ms。
func (s *stats) incWide() {
	s.mu.Lock()         // 拿笔
	defer s.mu.Unlock() // 函数返回前还笔，保证异常路径也会解锁

	time.Sleep(slowWork) // 慢操作在锁内：其他 worker 无法进入临界区
	s.n++                // 共享状态修改；这一步才真正需要锁
}

// incNarrow 只把“修改共享状态”放进临界区，锁的范围刚好覆盖不变量。
//
// 时间线：
//
//	W1 [------ 50ms ------][Lock][n++][Unlock]
//	W2 [------ 50ms ------][Lock][n++][Unlock]
//	W3 [------ 50ms ------][Lock][n++][Unlock]
//
// 10 个 50ms 的慢操作重叠执行，只有最后的 n++ 需要短暂排队，总耗时约 50ms。
func (s *stats) incNarrow() {
	time.Sleep(slowWork) // 慢操作在锁外：多个 worker 可以同时进行

	s.mu.Lock()   // 只在读写 n 时占用锁
	s.n++         // 读 n、加一、写回，三步必须整体互斥
	s.mu.Unlock() // 尽早释放，减少其他 worker 的等待
}

// incRecursive 演示 Mutex 不可重入。
//
// 同一个 goroutine 已经持有锁，又调用自己再次 Lock 同一把锁：
// Go 的 Mutex 不记录 owner，不会提示“你已经拿过了”，只会一直等自己，
// 直到所有 goroutine 都阻塞，运行时输出 deadlock。
func (s *stats) incRecursive() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.incRecursive() // 第二次 Lock 永远等不到 Unlock
}

// scenario 描述一种模式：执行哪个函数、起多少个 worker、结束后是否打印结果。
type scenario struct {
	run    func()
	count  int
	report bool
}

func main() {
	// 默认跑 wide；也可以通过命令行参数选择模式。
	mode := "wide"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	var s stats

	// 模式到行为的映射：查表一次，之后每次 wg.Go 直接执行 run，
	// 不需要在每个 goroutine 里再判断 mode。
	scenarios := map[string]scenario{
		"wide":      {run: s.incWide, count: workers, report: true},
		"narrow":    {run: s.incNarrow, count: workers, report: true},
		"recursive": {run: s.incRecursive, count: 1, report: false},
	}

	sc, ok := scenarios[mode]
	if !ok {
		fmt.Println("usage: go run examples/go/concurrency/lock-scope/main.go wide|narrow|recursive")
		return
	}

	var wg sync.WaitGroup // 用来等待所有 worker 结束
	start := time.Now()   // 只统计整个并发阶段，不含进程启动

	// Go 1.22+ 可以直接 range 一个整数，等价于 for i := 0; i < sc.count; i++。
	for range sc.count {
		// wg.Go 等价于 wg.Add(1) + go func() + defer wg.Done()，Go 1.25 提供。
		// 直接传入方法值，避免了闭包内部的 mode 分支。
		wg.Go(sc.run)
	}
	wg.Wait() // 等所有 worker 完成；recursive 模式会先被检测为死锁

	if !sc.report {
		return // recursive 模式不会执行到这里
	}

	// elapsed 是墙钟时间；Round 让输出稳定在 10ms 粒度。
	fmt.Printf("mode=%s workers=%d count=%d elapsed=%s\n",
		mode, sc.count, s.n, time.Since(start).Round(10*time.Millisecond))
}
