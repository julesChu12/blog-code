# Go 示例

配合博客 [开发与阅读](https://blog.yunmengspace.com/) 使用。

- `day01/`：Slice 专题实验
- `map-lifecycle/`：Map 专题实验
- `concurrency/`：Context、Channel、数据竞争与锁专题实验

每个目录下是一个独立的 `main.go`，可以直接运行：

```sh
go run concurrency/lock-scope/main.go wide
go run concurrency/context-lifecycle/main.go
```

这些文件由博客源码仓库的 `examples/go/` 自动同步而来，请优先修改源码仓库，而不是直接改这里。