// 学习目标：验证父取消与子超时的先后顺序如何决定 Err() 和 Cause()。
// 运行：go run examples/go/concurrency/context-cause/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	// 场景一：父先取消，子的 1 秒超时后到期。
	parent, cancelParent := context.WithCancelCause(context.Background())
	child, cancelChild := context.WithTimeoutCause(parent, time.Second, errors.New("child-timeout"))
	defer cancelChild()

	cancelParent(errors.New("parent-cancel"))
	fmt.Printf("父先取消: err=%v cause=%v\n", child.Err(), context.Cause(child))

	time.Sleep(1100 * time.Millisecond) // 等子的定时器真正到期
	fmt.Printf("定时器后: err=%v cause=%v\n", child.Err(), context.Cause(child))

	// 场景二：子先超时，父后取消。
	parent2, cancelParent2 := context.WithCancelCause(context.Background())
	child2, cancelChild2 := context.WithTimeoutCause(parent2, 50*time.Millisecond, errors.New("child-timeout"))
	defer cancelChild2()

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("子先超时: err=%v cause=%v\n", child2.Err(), context.Cause(child2))

	cancelParent2(errors.New("parent-cancel"))
	fmt.Printf("父后取消: err=%v cause=%v\n", child2.Err(), context.Cause(child2))
	fmt.Printf("父自身:   err=%v cause=%v\n", parent2.Err(), context.Cause(parent2))
}
