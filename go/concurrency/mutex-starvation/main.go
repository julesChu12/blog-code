// 学习目标：观察两件事。
//
//  1. 解锁只唤醒等待者，不把锁交给它；新来的 goroutine 可以先抢到锁（barging）。
//  2. 被唤醒的等待者仍然抢不到锁、等待超过 1ms 后，切到饥饿模式（starving 位）；
//     此后解锁直接交接。
//
// 警告：这里通过 unsafe 读取未导出字段，依赖 Go 1.25.8 的实现布局，
// 只用于学习，不要在生产代码中使用。
//
// 运行：go run examples/go/concurrency/mutex-starvation/main.go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// stateOf 把 sync.Mutex 起始位置的 4 个字节当作 state 读取。
func stateOf(m *sync.Mutex) int32 {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(m)))
}

// describe 把 state 拆成源码里的四个部分。
func describe(s int32) string {
	return fmt.Sprintf("state=%2d locked=%t woken=%t starving=%t waiters=%d",
		s, s&1 != 0, s&2 != 0, s&4 != 0, s>>3)
}

func main() {
	// 单 P：主 goroutine 不让出 CPU 时，被唤醒的等待者不会抢在前面执行，
	// 这样“新来者先抢到锁”和“等待者随后切饥饿模式”都能稳定复现。
	runtime.GOMAXPROCS(1)

	var m sync.Mutex
	m.Lock() // 主 goroutine 先持有锁

	started := make(chan struct{})
	acquired := make(chan struct{})
	allowUnlock := make(chan struct{})
	released := make(chan struct{})

	go func() {
		close(started)
		m.Lock()        // 阻塞，进入等待队列
		close(acquired) // 拿到锁后通知主 goroutine
		<-allowUnlock   // 停住，方便观察持有锁时的状态
		m.Unlock()
		close(released)
	}()

	<-started
	time.Sleep(20 * time.Millisecond) // 等等待者真正进入等待队列
	fmt.Println("等待者入队:    ", describe(stateOf(&m)))

	// 解锁：只会唤醒等待者，并把 woken 置 1，不会把锁直接交给它。
	m.Unlock()
	// 主 goroutine 立刻重新加锁，模拟“新来的 goroutine 抢先拿到锁”。
	m.Lock()
	fmt.Println("新来的先抢到:  ", describe(stateOf(&m)))

	// 被唤醒的等待者在这 5ms 内运行：发现锁被占，且等待已超过 1ms，
	// 于是设置 starving 位并重新排队。
	time.Sleep(5 * time.Millisecond)
	fmt.Println("等待超 1ms:    ", describe(stateOf(&m)))

	// 饥饿模式下 Unlock 直接交接；等待者拿到锁并清掉 starving 位。
	m.Unlock()
	<-acquired
	fmt.Println("等待者拿到:    ", describe(stateOf(&m)))

	close(allowUnlock)
	<-released
	fmt.Println("等待者释放:    ", describe(stateOf(&m)))
}
