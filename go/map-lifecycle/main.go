package main

import "fmt"

func main() {
	// hint 表示预期条目数量；创建时 map 仍然为空。
	items := make(map[int]int, 8)
	fmt.Println("创建后长度:", len(items))

	// 插入 8 个不同的 key，记录操作前后的可观察状态。
	for key := 1; key <= 8; key++ {
		items[key] = key * 10
	}
	fmt.Println("插入 8 项后长度:", len(items))

	// 新增第 9 个 key；具体增长路径取决于 Go 版本和编译条件。
	items[9] = 90
	fmt.Println("插入第 9 项后长度:", len(items))

	// 更新已存在的 key，不会增加 map 元素数量。
	items[3] = 300
	value, ok := items[3]
	fmt.Println("更新后 key 3:", value, ok, "长度:", len(items))

	// 查找不存在的 key，得到 value 的零值和 false。
	value, ok = items[99]
	fmt.Println("不存在的 key 99:", value, ok)

	// 删除指定 key，之后该 key 不再存在。
	delete(items, 4)
	_, ok = items[4]
	fmt.Println("删除 key 4:", ok, "长度:", len(items))

	// 赋值后的 alias 与 items 引用同一张 map。
	alias := items
	alias[1] = 11
	fmt.Println("通过 alias 更新 key 1:", items[1])

	// 复制条目到新 map；int value 是值复制，后续修改互不影响。
	snapshot := make(map[int]int, len(items))
	for key, value := range items {
		snapshot[key] = value
	}
	snapshot[1] = 999
	fmt.Println("原 map 与快照的 key 1:", items[1], snapshot[1])

	// 只汇总 value，不依赖 map 的遍历顺序。
	sum := 0
	for _, value := range items {
		sum += value
	}
	fmt.Println("剩余 value 总和:", sum)
}
