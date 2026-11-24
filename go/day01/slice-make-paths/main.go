// 学习目标：对比 nil、空字面量、make、append 和 make+copy 的创建路径。
// 运行：go run examples/go/day01/slice-make-paths/main.go
// 逃逸：go build -o /tmp/day01-make-paths -gcflags="-m=2" examples/go/day01/slice-make-paths/main.go
// 汇编：go build -o /tmp/day01-make-paths -gcflags="-S" examples/go/day01/slice-make-paths/main.go
// 下面的 runtime 调用和 K=4 均已在 Go 1.25.8 / darwin arm64 默认优化下核对。
// 它们是实现细节，可能随版本、架构、编译选项变化；输出值本身不能证明栈/堆位置。
// case 8 和 case 9 共用 case8MakeVarLocal，以 n=4 和 n=5 触发两条路径。
// 声明 main 包，使本文件可以独立运行。
package main

// 开始分组导入本实验使用的标准库包。
import (
	// fmt 提供格式化输出，用来展示每组实验结果。
	"fmt"
	// unsafe 用于 SliceData 和地址转换，仅观察地址，不通过地址读写元素。
	"unsafe"
	// 结束导入列表。
)

// 声明包级切片，初始值为 [go slice runtime]、len=cap=3。
// 本机编译器把底层数组静态初始化到可写数据段，无需在初始化时动态分配该数组。
// staticTags 是变量，并非只读常量；元素也不是通过本例递归复制得到的。
var staticTags = []string{"go", "slice", "runtime"}

// data 将 SliceData 的指针转为地址数值，仅用于展示；uintptr 不会让对象保持存活。
// nil 切片返回 0；非 nil 且 cap=0 时只保证非 nil 指针，其具体地址未由语言规定。
func data(s []int) uintptr {
	// SliceData 获取数据指针，unsafe.Pointer 允许转为 uintptr；不得把返回值当作可长期解引用的指针。
	return uintptr(unsafe.Pointer(unsafe.SliceData(s)))
	// 结束地址观测函数。
}

// ---------------------------------------------------------------
// case 1: var s []T —— nil 切片
// 期望：数据地址、长度和容量均为 0；本机汇编无分配调用。
// ---------------------------------------------------------------

// 下一行禁止内联，便于独立检查本函数的逃逸分析和汇编。
//
//go:noinline
func case1Nil() (uintptr, int, int) {
	// 声明切片但不初始化，得到 nil 切片；len=cap=0，没有底层数组。
	var s []int
	// 同时返回数据地址、长度和容量，期望 (0, 0, 0)。
	return data(s), len(s), cap(s)
	// 结束 nil 切片实验函数。
}

// ---------------------------------------------------------------
// case 2: []T{} —— 空字面量
// 语义：非 nil、len=cap=0；本机最终汇编直接取 runtime.zerobase，无分配调用。
// 不能把编译器降低过程中的概念节点当成最终一定执行的调用链。
// ---------------------------------------------------------------

// 下一行禁止内联，保留空字面量创建函数供观察。
//
//go:noinline
func case2EmptyLit() []int {
	// 返回非 nil 的空切片；不包含可索引元素，不能访问 s[0]。
	return []int{}
	// 结束空字面量实验函数。
}

// ---------------------------------------------------------------
// case 3: []T{...} 逃逸 —— newobject
// 本机汇编调用 runtime.newobject 分配 [3]int；不是 runtime.makeslice。
// ---------------------------------------------------------------

// 下一行禁止内联，使返回数组的生命周期跨越清晰的函数边界。
//
//go:noinline
func case3LitEscape(seed int) []int {
	// 返回三个连续整数；seed=10 时为 [10 11 12]、len=cap=3，本机数组逃逸到堆。
	return []int{seed, seed + 1, seed + 2}
	// 结束逃逸字面量实验函数。
}

// ---------------------------------------------------------------
// case 4: []T{...} 不逃逸 —— 栈数组
// 本机分析不逃逸，汇编使用局部数组，无分配调用；其他版本可能进一步优化存储。
// ---------------------------------------------------------------

// 下一行禁止内联，便于比较局部字面量与返回字面量的实现。
//
//go:noinline
func case4LitLocal(seed int) int {
	// 创建三个元素且只在函数内部使用；seed=10 时 s=[10 11 12]。
	s := []int{seed, seed + 1, seed + 2}
	// 初始化累加器为 0。
	total := 0
	// _ 忽略索引，v 依次取得三个整数的值副本。
	for _, v := range s {
		// 累加元素；seed=10 时总和依次为 10、21、33。
		total += v
		// 结束本轮累加，继续读取下一个元素。
	}
	// 只返回整数结果；seed=10 时期望 33，不返回切片。
	return total
	// 结束局部字面量实验函数。
}

// ---------------------------------------------------------------
// case 5: make([]T, 常量) 不逃逸 —— 栈数组
// 本机 walkMakeSlice 对 EscNone + 常量 cap 采用局部数组路径，无 makeslice 调用。
// ---------------------------------------------------------------

// 下一行禁止内联，保留常量长度 make 的局部使用场景。
//
//go:noinline
func case5MakeConstLocal() int {
	// 创建 [0 0 0]，len=cap=3；本机分析不逃逸，汇编使用局部数组。
	s := make([]int, 3)
	// 将首元素改为 7，得到 [7 0 0]。
	s[0] = 7
	// 初始化求和变量为 0。
	total := 0
	// 遍历元素值，忽略索引。
	for _, v := range s {
		// 累加 7、0、0，最终总和为 7。
		total += v
		// 结束本轮累加。
	}
	// 返回 7；局部切片不需要交给调用者。
	return total
	// 结束常量长度局部 make 实验函数。
}

// ---------------------------------------------------------------
// case 6: make([]T, 常量) 逃逸 —— makeslice
// 本例返回非空切片且禁止内联，分析为逃逸；本机汇编调用 runtime.makeslice。
// 不要泛化成所有逃逸 make 都一定调用 makeslice，case 13 的零长度即有不同优化。
// ---------------------------------------------------------------

// 下一行禁止内联，便于观察返回非空 make 结果的分配路径。
//
//go:noinline
func case6MakeConstEscape() []int {
	// 返回 [0 0 0]、len=cap=3；数组在函数返回后仍需要有效，本机分配到堆。
	return make([]int, 3)
	// 结束常量长度逃逸 make 实验函数。
}

// ---------------------------------------------------------------
// case 7: make([]T, 变量) 逃逸 —— makeslice
// ---------------------------------------------------------------

// 下一行禁止内联，防止调用处常量 n=5 被内联带入本函数。
//
//go:noinline
func case7MakeVarEscape(n int) []int {
	// 按运行时 n 创建并返回零值切片，len=cap=n；本例 n=5，负数会 panic。
	// 本机生成 makeslice 调用；n 过大也可能导致运行时错误或分配失败。
	return make([]int, n)
	// 结束变量长度逃逸 make 实验函数。
}

// ---------------------------------------------------------------
// case 8/9: make([]T, 变量) 不逃逸 —— tryStack 分支
// 本机默认 K = VariableMakeThreshold(32 字节) / sizeof(int)(8 字节) = 4。
// 本机汇编对 0<=n<=4 使用栈数组，n>4 调用 makeslice；负数会 panic。
// 这是开启默认优化时的实现细节；求和输出不能证明采用了哪条分配路径。
// ---------------------------------------------------------------

// 下一行禁止内联，保留 n 为函数参数，以观察变量长度 make 的运行时分支。
//
//go:noinline
func case8MakeVarLocal(n, seed int) int {
	// 创建 len=cap=n 的零值切片；本机虽分析为不逃逸，n>4 时仍会走堆分配分支。
	s := make([]int, n)
	// 按索引从 0 遍历到 n-1；本例 n 非负。
	for i := range s {
		// 填入 seed+i；seed=1 时得到 [1 2 ... n]。
		s[i] = seed + i
		// 结束当前元素赋值，继续下一个索引。
	}
	// 初始化累加器为 0。
	total := 0
	// 忽略索引，逐个读取元素值。
	for _, v := range s {
		// 累加元素；seed=1 时，n=4 得到 10，n=5 得到 15。
		total += v
		// 结束本轮累加。
	}
	// 返回整数而非切片；元素数不同时和也不同，不能声称两组输出相同。
	return total
	// 结束变量长度局部 make 实验函数。
}

// ---------------------------------------------------------------
// case 10: append 增长 —— growslice
// 容量从 2 开始增长；实际新容量还受分配器尺寸对齐影响，不只取决于 nextslicecap。
// ---------------------------------------------------------------

// 下一行禁止内联，便于观察 append 的 growslice 调用与返回数组的逃逸。
//
//go:noinline
func case10Append(n int) ([]int, []int) {
	// 创建长度为 0、容量为 2 的 int 切片；前两次追加无需扩容，本例 n=10。
	s := make([]int, 0, 2)
	// 为容量记录预留 n+1 个槽位：初始容量一次、每次追加后各一次；本例要求 n>=0。
	caps := make([]int, 0, n+1)
	// 先记录初始容量 2，期望 caps=[2]。
	caps = append(caps, cap(s))
	// 依次追加 0 到 n-1；n=10 时执行 10 轮。
	for i := 0; i < n; i++ {
		// 追加 i 并保存返回的描述符；容量不足时才执行 growslice 分支。
		s = append(s, i)
		// 记录本次追加后的容量；caps 已预留足够空间，本例执行时自身不会扩容。
		caps = append(caps, cap(s))
		// 结束本轮追加和记录。
	}
	// 返回两个切片；n=10 时 s=[0 1 2 3 4 5 6 7 8 9]。
	// 本机 caps=[2 2 2 4 4 8 8 8 8 16 16]；具体容量序列不是跨版本保证。
	return s, caps
	// 结束 append 容量记录函数。
}

// ---------------------------------------------------------------
// case 11: make + copy，元素含指针 —— makeslicecopy
// 本机编译器把 make+copy 合并，调用 runtime.makeslicecopy；string 值内部含数据指针。
// ---------------------------------------------------------------

// 下一行禁止内联，保留 make+copy 的独立函数用于观察优化。
//
//go:noinline
func case11MakeCopyString(src []string) []string {
	// 为 len(src) 个字符串值分配独立数组；初始每项是空字符串，len=cap=len(src)。
	dst := make([]string, len(src))
	// 复制 min(len(dst),len(src)) 个元素，本例复制全部；返回的复制数量在此被忽略。
	// 复制的是字符串值，不递归复制字符串字节；字符串不可变，重新赋值 dst[i] 不影响 src[i]。
	copy(dst, src)
	// 返回使用独立数组的目标切片；底层字符串数据可以共享。
	return dst
	// 结束字符串切片复制函数。
}

// ---------------------------------------------------------------
// case 12: make + copy，元素无指针且编译器能证明源、目标长度相等。
// 长度不需要是编译期常量；本机优化为 runtime.mallocgc + runtime.memmove。
// ---------------------------------------------------------------

// 下一行禁止内联，便于与字符串元素的复制路径比较。
//
//go:noinline
func case12MakeCopyInt(src []int) []int {
	// 创建与 src 等长的零值 int 切片；目标数组与源数组独立。
	dst := make([]int, len(src))
	// 将所有整数值复制到 dst；int 不含指针，复制后修改目标元素不影响源元素。
	copy(dst, src)
	// 返回新切片；src=[1 2 3] 时，返回值初始也为 [1 2 3]。
	return dst
	// 结束整数切片复制函数。
}

// ---------------------------------------------------------------
// case 13: make([]T, 0) 且逃逸
// 语义：非 nil、len=cap=0；本机最终汇编直接取 runtime.zerobase，没有 makeslice 调用。
// 即使逃逸分析标记为 escapes to heap，也不能据此断定实际分配了非零大小的堆对象。
// ---------------------------------------------------------------

// 下一行禁止内联，单独观察零长度 make 的最终生成代码。
//
//go:noinline
func case13MakeZeroEscape() []int {
	// 返回非 nil 的空切片；与 var s []int 的 nil 状态不同，但二者 len 和 cap 都为 0。
	return make([]int, 0)
	// 结束零长度 make 实验函数。
}

// 程序入口：依次打印实验结果；%v 展示值，%d 展示整数，\n 表示换行。
func main() {
	// 打印 nil 切片实验标题。
	fmt.Println("=== case 1: var s []int ===")
	// 用多返回值接收地址 p、长度 n、容量 c，期望三者均为 0。
	p, n, c := case1Nil()
	// 打印 nil 切片的三个观测值，期望 data=0 len=0 cap=0。
	fmt.Printf("data=%d len=%d cap=%d (data==0 表示 nil)\n\n", p, n, c)

	// 将空字面量和零长度 make 放在一起，比较它们的地址观测值。
	fmt.Println("=== case 2: []int{} 与 case 13: make([]int, 0) ===")
	// 第一次获取非 nil 的空字面量切片，len=cap=0。
	e1 := case2EmptyLit()
	// 再次获取空字面量切片，用来观察两次调用是否返回相同地址。
	e2 := case2EmptyLit()
	// 获取 make 创建的非 nil 空切片，与字面量形式比较。
	z := case13MakeZeroEscape()
	// 本机期望 true，因为最终代码都直接使用 zerobase；这不是空切片的地址唯一性保证。
	fmt.Printf("两次 []int{} 的 data 相同: %v\n", data(e1) == data(e2))
	// 本机期望 true；明确标注实现观察，避免把地址相等当作跨版本语言规则。
	fmt.Printf("[]int{} 与 make([]int,0) 的 data 相同: %v (本机使用 zerobase，非语言保证)\n\n", data(e1) == data(z))

	// 打印返回非空字面量的实验标题。
	fmt.Println("=== case 3: []int{...} 逃逸 ===")
	// 以 10 为起点创建三个元素，期望 l=[10 11 12]。
	l := case3LitEscape(10)
	// 期望值=[10 11 12]、len=3、cap=3；该字面量没有额外预留槽位。
	fmt.Printf("值=%v len=%d cap=%d (cap==len，字面量没有余量)\n", l, len(l), cap(l))
	// 打印数据地址，具体数值可能每次运行不同；不能仅凭数值判断栈或堆。
	fmt.Printf("data=%d\n\n", data(l))

	// 打印仅在函数内部使用字面量的实验标题。
	fmt.Println("=== case 4: []int{...} 不逃逸 ===")
	// 计算 10+11+12，期望 sum=33；是否分配由编译结果验证。
	fmt.Printf("sum=%d\n\n", case4LitLocal(10))

	// 打印常量长度 make 在函数内使用的实验标题。
	fmt.Println("=== case 5: make([]int, 3) 不逃逸 ===")
	// 对 [7 0 0] 求和，期望 sum=7。
	fmt.Printf("sum=%d\n\n", case5MakeConstLocal())

	// 打印常量长度 make 返回切片的实验标题。
	fmt.Println("=== case 6: make([]int, 3) 逃逸 ===")
	// 获取返回的三个零值整数，期望 m6=[0 0 0]。
	m6 := case6MakeConstEscape()
	// 期望值=[0 0 0]、len=3、cap=3。
	fmt.Printf("值=%v len=%d cap=%d\n\n", m6, len(m6), cap(m6))

	// 打印变量长度 make 返回切片的实验标题。
	fmt.Println("=== case 7: make([]int, n) 逃逸 ===")
	// 传入 n=5；禁止内联使被调函数内的 n 仍是运行时参数。
	m7 := case7MakeVarEscape(5)
	// 期望 len=5、cap=5；make 未指定容量参数时，容量等于长度。
	fmt.Printf("len=%d cap=%d (make 不带 cap 时 cap==len)\n\n", len(m7), cap(m7))

	// 打印变量长度局部 make 的两组实验；K=4 仅对应本机默认优化配置。
	fmt.Println("=== case 8/9: make([]int, n) 不逃逸（本机 Go 1.25.8 / arm64 默认 K=4）===")
	// n=4、seed=1，期望 sum=10；本机汇编对应栈数组分支，其他环境应重新核对。
	fmt.Printf("n=4 sum=%d (本机默认优化走栈分支)\n", case8MakeVarLocal(4, 1))
	// n=5、seed=1，期望 sum=15；本机汇编对应 makeslice 分支。
	fmt.Printf("n=5 sum=%d (本机默认优化走 makeslice 分支)\n", case8MakeVarLocal(5, 1))
	// 两组和不同，输出不能证明分配位置；用 Printf 控制换行，避免 Println 冗余换行的 vet 报错。
	fmt.Printf("两组求和分别为 10 和 15；分配路径需查看当前环境的编译结果。\n\n")

	// 打印 append 扩容实验标题。
	fmt.Println("=== case 10: append 增长 ===")
	// 接收最终切片和容量序列；caps 包含初始容量及 10 次追加后的容量，共 11 项。
	s10, caps := case10Append(10)
	// 期望最终切片=[0 1 2 3 4 5 6 7 8 9]。
	fmt.Printf("最终切片=%v\n", s10)
	// 本机期望 [2 2 2 4 4 8 8 8 8 16 16]；未改变的容量说明该轮追加没有扩容。
	fmt.Printf("cap 变化=%v\n\n", caps)

	// 打印 string 元素的复制实验标题；string 值内部含指针。
	fmt.Println("=== case 11: make+copy，元素含指针 ===")
	// 创建源切片 [a b c]，作为复制输入。
	src11 := []string{"a", "b", "c"}
	// 复制到独立的元素数组，初始 dst11=[a b c]。
	dst11 := case11MakeCopyString(src11)
	// 替换目标数组中的首个字符串值；不修改源数组，也不是修改字符串内部字节。
	dst11[0] = "changed"
	// 期望 dst=[changed b c]、src=[a b c]，说明元素槽位独立。
	fmt.Printf("dst=%v src=%v (改 dst 不影响 src)\n\n", dst11, src11)

	// 打印无指针 int 元素的复制实验标题。
	fmt.Println("=== case 12: make+copy，元素无指针 ===")
	// 创建源整数切片 [1 2 3]。
	src12 := []int{1, 2, 3}
	// 获取独立数组中的副本，初始 dst12=[1 2 3]。
	dst12 := case12MakeCopyInt(src12)
	// 只修改目标数组的首元素，不会改变 src12[0]。
	dst12[0] = 100
	// 期望 dst=[100 2 3]、src=[1 2 3]。
	fmt.Printf("dst=%v src=%v\n\n", dst12, src12)

	// 打印包级字面量实验标题，本机底层数组在数据段静态初始化。
	fmt.Println("=== 包级字面量（静态数据段）===")
	// 输出包级切片的元素、长度、容量和地址；这是跨多行的一次 Printf 调用。
	fmt.Printf("staticTags=%v len=%d cap=%d data=%d\n",
		// 前三个参数依次为 [go slice runtime]、3、3，与格式串前三个占位符对应。
		staticTags, len(staticTags), cap(staticTags),
		// 最后一个参数把数据指针转成地址数值；地址不固定，这里的右括号同时结束 Printf 调用。
		uintptr(unsafe.Pointer(unsafe.SliceData(staticTags))))
	// 结束 main，程序正常退出。
}
