// 学习目标：区分切片本身与底层数组，观察限制容量如何影响 append。
// 运行：go run examples/go/day01/slice-sharing/main.go
// 声明可执行程序所属的 main 包；程序从 main 函数开始执行。
package main

// 导入格式化输出包，使用 Println 和 Printf 展示实验结果。
import "fmt"

// demonstrate 运行一组实验；name 是标题，limitCapacity 决定是否把 a 的容量限制为 2。
func demonstrate(name string, limitCapacity bool) {
	// 创建元素为 [1 2 3 4] 的切片；base 的 len 和 cap 都为 4，每次调用都重新创建。
	base := []int{1, 2, 3, 4}
	// 截取索引 [0, 2)；a=[1 2]、len=2、cap=4，与 base 共享底层数组。
	a := base[:2]
	// 判断是否演示限制容量的情况；true 时执行下面的三下标切片表达式。
	if limitCapacity {
		// low=0、high=2、max=2，使 len=2、cap=2；此时仍共享原数组，并未复制元素。
		a = base[:2:2]
		// 结束容量限制分支；后续 append 会根据 a 的当前容量决定是否分配新数组。
	}

	// 打印本组实验的名称，便于区分两种切片写法。
	fmt.Println(name)
	// %d 输出整数；期望 len 始终为 2，cap 在普通切片时为 4、限制容量时为 2。
	fmt.Printf("before append: len(a)=%d cap(a)=%d\n", len(a), cap(a))

	// 追加 9 并把返回的切片赋给 b；a 的长度仍为 2。
	// cap=4 时复用原数组并覆盖 base[2]；cap=2 时容量不足，复制到新数组后再追加。
	b := append(a, 9)
	// 修改 b 的首元素；复用数组时同步影响 base[0] 和 a[0]，新数组时不影响它们。
	b[0] = 100

	// %v 输出切片元素，\n 换行。
	// 未限制：base=[100 2 9 4]、a=[100 2]、b=[100 2 9]。
	// 已限制：base=[1 2 3 4]、a=[1 2]、b=[100 2 9]。
	fmt.Printf("base=%v\na=%v\nb=%v\n", base, a, b)
	// & 取元素地址，%t 输出布尔值；base 和 a 始终指向同一个首元素，期望 true。
	fmt.Printf("base/a share start: %t\n", &base[0] == &a[0])
	// 比较 a 与 b 的首元素地址；未限制时 true，限制后 false；额外换行分隔实验。
	fmt.Printf("a/b share start: %t\n\n", &a[0] == &b[0])
	// 结束本组实验；函数没有返回值，控制流回到调用处。
}

// main 是程序入口，按顺序运行两种情况，便于直接比较输出。
func main() {
	// 普通切片保留剩余容量；期望 append 复用 base 的底层数组。
	demonstrate("a := base[:2]", false)
	// 三下标切片限制容量；期望本次 append 分配新数组，但限制动作本身不会复制。
	demonstrate("a := base[:2:2]", true)
	// 结束 main，程序正常退出。
}
