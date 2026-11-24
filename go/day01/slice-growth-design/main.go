// 学习目标：比较扩容频率、累计搬迁元素数和预留空间之间的取舍。
// 运行：go run examples/go/day01/slice-growth-design/main.go
// 声明 main 包，使本文件可独立运行。
package main

// 导入 fmt，用于打印统计结果。
import "fmt"

// stats 保存一次实验的统计数据。
type stats struct {
	// growths 记录容量增长次数，包含从 0 到首次非零容量的分配。
	growths int
	// copied 是扩容时旧元素数量的累加值；单位是元素，不是字节或耗时。
	copied int
	// cap 是最终容量；这是结构体字段，与内置 cap 函数不是同一个标识符。
	cap int
	// 结束统计结构体定义。
}

// 模拟逐个追加 count 个元素，只计算逻辑容量和搬迁数量，不模拟分配器、GC 或运行时间。
// 本例 count>=0；next 接收旧容量和所需长度，必须返回至少 needed 的容量。
func simulate(count int, next func(capacity, needed int) int) stats {
	// 所有统计字段初始为 0，对应尚未分配容量的状态。
	var result stats
	// length 是追加前的长度；每轮模拟追加一个元素，最终长度为 count。
	for length := 0; length < count; length++ {
		// 所需长度超过现有容量时才扩容；容量足够时不需要搬迁旧元素。
		if length+1 > result.cap {
			// 扩容次数加 1，首次从空状态分配也计入。
			result.growths++
			// 假设本轮扩容需搬迁全部 length 个旧元素；首次 length=0，不搬迁旧元素。
			result.copied += length
			// 用策略函数计算新容量，确保能容纳追加后的 length+1 个元素。
			result.cap = next(result.cap, length+1)
			// 结束本轮扩容处理。
		}
		// 结束本轮模拟，length++ 表示又追加了一个元素。
	}
	// 返回统计值的副本，供 report 输出。
	return result
	// 结束模拟函数。
}

// 包级变量保留真实实验的最终切片，使底层数组在 observeGo 返回后仍可使用。
var retained []int

// 实际执行 Go append，通过容量变化统计增长次数和逻辑搬迁量，不直接测量复制指令。
// 下一行禁止内联，保留函数边界；最终结果由 retained 持有以固定本例的逃逸条件。
//
//go:noinline
func observeGo(count int) stats {
	// 从 nil 切片开始，len=cap=0，第一次 append 就需要获得容量。
	var values []int
	// 初始化三个统计字段为 0，与模拟采用相同计数口径。
	var result stats
	// 依次追加 0 到 count-1；count=10000 时最终长度为 10000。
	for i := 0; i < count; i++ {
		// 同时保存追加前的长度和容量，供追加后比较。
		oldLen, oldCap := len(values), cap(values)
		// 追加当前整数并保存返回的新描述符；底层数组可能复用，也可能重新分配。
		values = append(values, i)
		// 容量变化说明本轮获得了更大的底层数组。
		if cap(values) != oldCap {
			// 记录一次容量增长，包含首次从 nil 状态分配。
			result.growths++
			// 把旧长度计作本轮需保留的旧元素数量，不包含新元素。
			result.copied += oldLen
			// 结束本轮扩容计数。
		}
		// 结束当前追加，继续下一个整数。
	}
	// 保存最终切片，使其底层数组在本函数返回后仍被引用。
	retained = values
	// 记录最终容量；它与 count 的差就是尚未使用的槽位数。
	result.cap = cap(values)
	// 返回真实执行得到的统计值。
	return result
	// 结束真实追加实验。
}

// 统一打印名称和统计数据，避免各策略重复实现输出逻辑。
func report(name string, count int, result stats) {
	// %s 打印名称，%d 打印整数；spare=cap-count，表示闲置的元素槽位数。
	fmt.Printf("%s: growths=%d copied=%d cap=%d spare=%d\n", name, result.growths, result.copied, result.cap, result.cap-count)
	// 结束输出函数。
}

// 程序入口：使用相同元素数量对比三种模型和当前 Go 实现。
func main() {
	// 固定追加 10000 个元素，足以展示精确扩容带来的累计搬迁开销。
	const count = 10000
	// 打印公共实验参数，期望 elements=10000。
	fmt.Printf("elements=%d\n", count)
	// 传入匿名策略函数，_ 忽略旧容量；期望 growths=10000、copied=49995000、cap=10000、spare=0。
	report("exact model", count, simulate(count, func(_, needed int) int {
		// 每次只分配刚好够用的容量，下一次又需扩容，搬迁总数为 0+1+...+9999。
		return needed
		// 依次结束匿名函数、simulate 调用和 report 调用。
	}))
	// 翻倍策略；期望 growths=15、copied=16383、cap=16384、spare=6384。
	report("2x model", count, simulate(count, func(capacity, needed int) int {
		// 从 0 直接翻倍仍然是 0，因此首次分配需要单独处理。
		if capacity == 0 {
			// 首次 needed=1，建立初始容量 1。
			return needed
			// 结束初始容量分支。
		}
		// 后续容量翻倍；本例逐个追加，触发增长时 needed=capacity+1，因此能够容纳。
		return capacity * 2
		// 结束匿名策略并输出模拟结果。
	}))
	// 约 1.25 倍策略；期望 growths=38、copied=41344、cap=10349、spare=349。
	report("1.25x model", count, simulate(count, func(capacity, needed int) int {
		// 单独处理初次分配，避免容量从 0 增长后仍为 0。
		if capacity == 0 {
			// 首次分配 1 个槽位，确保能容纳第一个元素。
			return needed
			// 结束初始容量分支。
		}
		// 整数除法向下取整，加 3 实现 ceil(capacity*5/4)，确保小容量也会增长。
		// 本例规模不会溢出；这只是模型，不是 Go 运行时的完整扩容算法。
		return (capacity*5 + 3) / 4
		// 结束匿名策略并输出模拟结果。
	}))
	// 本机 Go 1.25.8 / arm64 为 growths=19、copied=32415、cap=12288、spare=2288。
	// 真实结果可能随版本、架构或编译优化变化；上述统计不能直接用作耗时基准。
	report("Go observed", count, observeGo(count))
	// 结束 main，可据此比较少搬迁元素与少预留空间之间的取舍。
}
