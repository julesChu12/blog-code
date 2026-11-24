// 学习目标：观察切片赋值、参数传递、扩容和指针元素的浅拷贝。
// 运行：go run examples/go/day01/slice-internals/main.go
// 数值观测基于 Go 1.25.8 / darwin arm64；具体扩容容量和机器字长不属于跨平台保证。
// 声明 main 包，使本文件可独立运行。
package main

// 开始分组导入本示例需要的两个标准库包。
import (
	// fmt 用于打印元素、长度、容量和布尔比较结果。
	"fmt"
	// unsafe 仅用于 Sizeof，观察切片描述符与 int 的大小，不读取内部内存布局。
	"unsafe"
	// 结束导入列表。
)

// 接收切片描述符的副本；要求 len(s)>=1 且本例 cap(s)>=3，以观察原数组被修改。
func mutateAndAppend(s []int) {
	// 修改共享数组的首元素；调用者的 a[0] 和 base[0] 也会变为 100。
	s[0] = 100
	// 追加 9 并更新局部 s；本次容量足够，覆盖 base[2]，但调用者 a 的长度仍为 2。
	s = append(s, 9)
	// %v 打印元素，%d 打印整数；期望 s=[100 2 9]、len=3、cap=4。
	fmt.Printf("callee: s=%v len=%d cap=%d\n", s, len(s), cap(s))
	// 结束函数；没有返回 s，调用者不会得到这里更新后的切片长度。
}

// 演示切片赋值和参数传递复制描述符，而不是自动复制底层数组。
func assignmentAndParameter() {
	// 输出实验标题，区分后续各组结果。
	fmt.Println("== assignment and parameter ==")
	// 创建 [1 2 3 4]，len=cap=4，作为共享的底层数据。
	base := []int{1, 2, 3, 4}
	// a 引用前两个元素；len=2、cap=4，底层数组与 base 相同。
	a := base[:2]
	// 复制 a 的描述符给 b；两者的长度、容量相同，并引用同一个数组。
	b := a
	// 仅把 b 的长度缩短为 1，容量仍为 4；a 的长度仍为 2。
	b = b[:1]
	// & 取首元素地址，%t 打印布尔值；期望 len(a)=2、len(b)=1、shared start=true。
	fmt.Printf("len(a)=%d len(b)=%d shared start=%t\n", len(a), len(b), &a[0] == &b[0])
	// 传入 a 的描述符副本；期望共享元素被改写，但 a 自身的长度不变。
	mutateAndAppend(a)
	// 期望 a=[100 2]、len=2、cap=4、base=[100 2 9 4]，展示元素修改与长度修改的区别。
	fmt.Printf("caller: a=%v len=%d cap=%d base=%v\n", a, len(a), cap(a), base)
	// Sizeof(a) 只计算描述符，不含底层数组；int(0) 指定 int 类型。
	// 本机输出 slice bytes=24、int bytes=8，对应描述符的指针、长度、容量三个机器字。
	fmt.Printf("slice bytes=%d int bytes=%d\n\n", unsafe.Sizeof(a), unsafe.Sizeof(int(0)))
	// 结束赋值与传参实验。
}

// 根据 limitCapacity 对比 append 复用数组与分配新数组的情况。
func appendCase(limitCapacity bool) {
	// 每组都新建 [1 2 3 4]，避免上一组的修改污染本组结果。
	base := []int{1, 2, 3, 4}
	// 默认 a=[1 2]、len=2、cap=4，可以再追加两个元素而不扩容。
	a := base[:2]
	// true 时进入限制容量的分支。
	if limitCapacity {
		// 把 a 的容量限制为 2，但仍共享 base；此处不会分配或复制数组。
		a = base[:2:2]
		// 结束分支，此后 a 的容量由本组参数决定。
	}
	// 返回 len=3 的 b；未限制时覆盖 base[2]，限制后容量不足则复制到新数组。
	b := append(a, 9)
	// 把 b[0] 改为 100；是否影响 a 和 base，取决于 b 是否仍共享数组。
	b[0] = 100
	// 打印本组是否限制容量，期望依次为 false 和 true。
	fmt.Printf("limited=%t\n", limitCapacity)
	// false 时 base=[100 2 9 4]，true 时 base=[1 2 3 4]；len、cap 均保持 4。
	fmt.Printf("base=%v len=%d cap=%d\n", base, len(base), cap(base))
	// false 时 a=[100 2]、len=2、cap=4；true 时 a=[1 2]、len=2、cap=2。
	fmt.Printf("a=%v len=%d cap=%d\n", a, len(a), cap(a))
	// 两组 b 都为 [100 2 9]、len=3；本机两组 cap 都为 4，扩容后的具体容量是实现细节。
	fmt.Printf("b=%v len=%d cap=%d\n", b, len(b), cap(b))
	// base/a 始终 true；a/b 在 false 组为 true，在 true 组为 false。
	fmt.Printf("base/a shared start=%t a/b shared start=%t\n\n", &base[0] == &a[0], &a[0] == &b[0])
	// 结束本组 append 实验。
}

// 观察小容量、大容量和一次追加多个元素的扩容结果，不把具体倍率当成语言规则。
func capacityGrowth() {
	// 打印扩容实验标题。
	fmt.Println("== capacity growth ==")
	// _ 忽略索引，oldCap 依次为 2、512、1024，用于比较不同规模。
	for _, oldCap := range []int{2, 512, 1024} {
		// 不指定第三个参数时 cap=len=oldCap，全部元素初始为 0，容量已满。
		a := make([]int, oldCap)
		// 再追加一个 0，必然超过原容量；b 长度为 oldCap+1，引用扩容后的数组。
		b := append(a, 0)
		// 本机三个 (old cap, new len, new cap) 为 (2,3,4)、(512,513,848)、(1024,1025,1536)。
		// 实际容量还受分配器尺寸对齐影响，不能简单记成永远翻倍或永远乘 1.25。
		fmt.Printf("old cap=%d new len=%d new cap=%d\n", oldCap, len(b), cap(b))
		// 结束当前规模的实验，继续下一个容量。
	}
	// 创建 [0 0]，len=cap=2，用于观察批量追加。
	a := make([]int, 2)
	// ... 将另一个切片的 9 个零展开追加；b 长度为 11，需要一次容纳全部新增元素。
	b := append(a, make([]int, 9)...)
	// 本机期望 old cap=2、new len=11、new cap=12；12 包含分配器对齐造成的余量。
	fmt.Printf("bulk: old cap=%d new len=%d new cap=%d\n\n", cap(a), len(b), cap(b))
	// 结束扩容实验。
}

// 演示扩容只复制元素值；当元素是指针时，不会递归复制指针指向的对象。
func shallowCopy() {
	// 打印浅拷贝实验标题。
	fmt.Println("== shallow copy ==")
	// 定义局部结构体类型；Value 是用于观察共享对象修改的整数成员。
	type Cell struct{ Value int }
	// 创建 Value=1 的 Cell 并取地址；cell 的类型是 *Cell。
	cell := &Cell{Value: 1}
	// 创建指针切片；a 长度和容量都是 1，唯一元素保存 cell 的地址。
	a := []*Cell{cell}
	// 容量限制为 1 后追加第二个指针，必然扩容；b[0] 仍是复制来的同一个 cell 地址。
	b := append(a[:1:1], &Cell{Value: 2})
	// 经由 b[0] 修改共享的 Cell 对象，因此 a[0].Value 也会变为 100。
	b[0].Value = 100
	// &a[0] 与 &b[0] 比较数组槽位地址，期望 false；a[0] 与 b[0] 比较对象地址，期望 true。
	// 最后一个值期望 100，说明新数组没有带来新的 Cell 对象。
	fmt.Printf("same array slot=%t same object=%t a[0].Value=%d\n", &a[0] == &b[0], a[0] == b[0], a[0].Value)
	// 结束浅拷贝实验。
}

// main 按概念顺序运行四组实验。
func main() {
	// 先理解描述符按值复制与数组共享，期望调用者的长度不受局部 append 影响。
	assignmentAndParameter()
	// 打印后续两组 append 实验的公共标题。
	fmt.Println("== append and backing arrays ==")
	// 容量充足，期望复用原数组。
	appendCase(false)
	// 容量受限，期望 append 分配新数组。
	appendCase(true)
	// 展示实际容量，理解扩容结果依赖当前运行时和元素类型。
	capacityGrowth()
	// 最后区分数组是否共享与元素指向的对象是否共享。
	shallowCopy()
	// 结束 main，程序正常退出。
}
