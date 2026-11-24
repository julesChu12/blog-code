// 学习目标：对比只在函数内使用的数组与需要在函数返回后继续使用的数组。
// 运行：go run examples/go/day01/slice-escape/main.go
// 分析：go build -o /tmp/day01-escape -gcflags="-m=2" examples/go/day01/slice-escape/main.go
// 声明 main 包，使本文件可以作为独立程序运行。
package main

// 声明包级切片，初始为 nil；随后持有返回的切片，使其底层数组在函数返回后仍可访问。
var retained []int

// localSum 只返回整数，不把 s 或底层数组交给外部；seed=1 时期望返回 10。
// 下一行是编译器指令，禁止内联以保留函数边界；指令必须紧邻函数声明。
//
//go:noinline
func localSum(seed int) int {
	// 创建 len=cap=4 的零值切片；本机逃逸分析为 does not escape。
	// 不逃逸表示无需因生命周期放到堆上；实际存储仍可能被编译器优化。
	s := make([]int, 4)
	// 依次取索引 0、1、2、3；单个 range 变量得到的是索引，而非元素值。
	for i := range s {
		// 按索引填充；seed=1 时 s 依次变为 [1 2 3 4]。
		s[i] = seed + i
		// 结束当前循环体，继续下一个索引，直到遍历完 4 个元素。
	}
	// 初始化整数累加器，期望后续保存所有元素的和。
	total := 0
	// _ 忽略索引，value 是当前元素的值副本；本例只读取，不修改元素。
	for _, value := range s {
		// 将当前元素加到总和；seed=1 时依次得到 1、3、6、10。
		total += value
		// 结束本轮累加，继续下一个元素。
	}
	// 只返回计算结果 4*seed+6；不返回切片，因此本例不需要让数组延长生命周期。
	return total
	// 结束 localSum 的定义。
}

// returnedSlice 返回切片，使调用者能在函数返回后读取底层数组。
// 下一行禁止内联；本机分析期望 make([]int, 4) escapes to heap。
//
//go:noinline
func returnedSlice(seed int) []int {
	// 创建 4 个零值元素；本例的底层数组需要存活到调用者使用它之后。
	s := make([]int, 4)
	// 只设置首元素；seed=7 时切片为 [7 0 0 0]。
	s[0] = seed
	// 返回切片描述符的副本，不是复制 4 个元素；返回值仍引用原来的底层数组。
	return s
	// 结束 returnedSlice 的定义；返回的数组仍可由 retained 引用。
}

// 程序入口：用数值输出确认行为，用 -gcflags="-m=2" 确认逃逸结论。
func main() {

	var s = make([]int, 2, 4)
	for i := range s {
		println(s[i])
	}
	println(len(s), cap(s))
	// 调用局部求和，期望打印 10；内置 println 用于演示，本机输出到标准错误流。
	println(localSum(1))
	// 把返回的 [7 0 0 0] 保存到包级变量，保持对其底层数组的引用。
	retained = returnedSlice(7)
	// 读取函数返回后仍有效的首元素，期望打印 7；输出本身不能证明分配位置。
	println(retained[0])
	retained[2] = 42
	// 读取函数返回后仍有效的第三个元素，期望打印 42；输出本身不能证明分配位置。
	println(retained[2])
	// 结束 main，程序正常退出。
}
