// 学习目标：直接观察 sync.Mutex 的 state 位（Go 1.25.8 内部布局），
// 验证 locked、woken、waiter 数量随加锁/解锁/等待如何变化。
//
// 警告：这里通过 unsafe 读取未导出字段，依赖 Go 1.25.8 的实现布局，
// 只用于学习，不要在生产代码中使用。
//
// 运行：go run examples/go/concurrency/mutex-state/main.go
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
// Go 1.25.8 的 sync.Mutex 包装 internal/sync.Mutex，第一个字段就是 state int32。
func stateOf(m *sync.Mutex) int32 {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(m)))
}

// describe 把 state 拆成源码里的四个部分。
func describe(s int32) string {
	return fmt.Sprintf("state=%2d locked=%t woken=%t starving=%t waiters=%d",
		s, s&1 != 0, s&2 != 0, s&4 != 0, s>>3)
}

func main() {
	// 单 P：主 goroutine 不让出 CPU 时，被唤醒的等待者不会抢在前面运行，
	// 这样能稳定观察到“解锁后、等待者运行前”的中间状态。
	runtime.GOMAXPROCS(1)

	var m sync.Mutex
	fmt.Println("初始:        ", describe(stateOf(&m)))

	m.Lock()
	fmt.Println("加锁后:      ", describe(stateOf(&m)))

	started := make(chan struct{})
	acquired := make(chan struct{})
	allowUnlock := make(chan struct{})
	released := make(chan struct{})

	go func() {
		close(started)
		m.Lock()        // 主 goroutine 持有锁，这里会阻塞
		close(acquired) // 通知主 goroutine：已经拿到锁
		<-allowUnlock   // 停在这里，方便观察“持有锁”的状态
		m.Unlock()
		close(released)
	}()

	<-started
	time.Sleep(50 * time.Millisecond) // 让等待者进入 sendq/等待队列
	fmt.Println("有等待者:    ", describe(stateOf(&m)))

	m.Unlock()
	// 解锁后、等待者被调度前：锁位已清除，woken 位被设置。
	fmt.Println("解锁后:      ", describe(stateOf(&m)))

	<-acquired
	fmt.Println("等待者拿到:  ", describe(stateOf(&m)))

	close(allowUnlock)
	<-released
	fmt.Println("等待者释放:  ", describe(stateOf(&m)))
}
