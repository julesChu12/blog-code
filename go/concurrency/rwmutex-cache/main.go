// 学习目标：观察 RWMutex 的读者并行、写者独占，以及“写者优先”。
//
// 场景：配置缓存。读多写少，读者可以同时读；写者更新时需要独占。
//
// 运行：
//
//	go run examples/go/concurrency/rwmutex-cache/main.go readers
//	go run examples/go/concurrency/rwmutex-cache/main.go writers
//	go run examples/go/concurrency/rwmutex-cache/main.go preference
package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	readWork  = 50 * time.Millisecond // 模拟一次读操作
	writeWork = 50 * time.Millisecond // 模拟一次写操作
)

type cache struct {
	rw    sync.RWMutex
	value string
}

// Get 用读锁：多个读者可以同时进入。
func (c *cache) Get() string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	time.Sleep(readWork) // 模拟读取配置、反序列化等
	return c.value
}

// Set 用写锁：写者独占，读和写都不能并发。
func (c *cache) Set(v string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	time.Sleep(writeWork) // 模拟更新配置、写回存储等
	c.value = v
}

func main() {
	mode := "readers"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	c := &cache{value: "v1"}
	start := time.Now()

	switch mode {
	case "readers":
		// 5 个读者同时读：总耗时应当接近一次读操作，而不是 5 次相加。
		var wg sync.WaitGroup
		for range 5 {
			wg.Go(func() { c.Get() })
		}
		wg.Wait()
		fmt.Printf("mode=readers readers=5 elapsed=%s value=%s\n",
			time.Since(start).Round(10*time.Millisecond), c.value)

	case "writers":
		// 3 个写者互相排斥：总耗时应当是 3 次写操作相加。
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() { c.Set("v2") })
		}
		wg.Wait()
		fmt.Printf("mode=writers writers=3 elapsed=%s value=%s\n",
			time.Since(start).Round(10*time.Millisecond), c.value)

	case "preference":
		runPreference(c)

	default:
		fmt.Println("usage: go run examples/go/concurrency/rwmutex-cache/main.go readers|writers|preference")
	}
}

// runPreference 演示“写者优先”：
// 写者一旦开始等待，新的 RLock 也会被挡住，直到写者完成。
func runPreference(c *cache) {
	start := time.Now()
	logf := func(format string, args ...any) {
		// 一次 Printf 输出整行，避免多个 goroutine 的前缀和内容交错。
		elapsed := time.Since(start).Round(10 * time.Millisecond)
		fmt.Printf("%-6s %s\n", elapsed, fmt.Sprintf(format, args...))
	}

	var wg sync.WaitGroup

	// 读者 1 先拿到读锁，并持有 50ms。
	wg.Go(func() {
		c.rw.RLock()
		logf("reader-1 获得读锁")
		time.Sleep(readWork)
		c.rw.RUnlock()
		logf("reader-1 释放读锁")
	})

	time.Sleep(10 * time.Millisecond)

	// 写者在 reader-1 还没释放时到达，开始等待。
	wg.Go(func() {
		logf("writer 等待写锁")
		c.rw.Lock()
		logf("writer 获得写锁")
		time.Sleep(writeWork)
		c.rw.Unlock()
		logf("writer 释放写锁")
	})

	time.Sleep(10 * time.Millisecond)

	// 新的读者此时到达；因为有写者在等，它必须排队。
	wg.Go(func() {
		logf("reader-2 尝试读锁")
		c.rw.RLock()
		logf("reader-2 获得读锁")
		c.rw.RUnlock()
		logf("reader-2 释放读锁")
	})

	wg.Wait()
}
