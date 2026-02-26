# Go 并发学习仓库

Go 并发编程教学资料集合，包含 467 个示例文件和 54 个测试。

## 技术栈

Go 1.16+（concurrency-master）/ Go 1.20+（concurrency-programming-via-go-code-master）

## 项目结构

| 目录 | 说明 | go.mod |
|------|------|--------|
| concurrency-master/ | 主教程库，23个主题模块（goroutine/channel/mutex/atomic/patterns等） | ✅ Go 1.16，无外部依赖 |
| concurrency-programming-via-go-code-master/ | 《Go并发编程》书籍示例，21章 | ✅ Go 1.20，含 etcd/redis/testify |
| my-concurrency/ | 个人练习文件 | ❌ 无独立模块 |
| 交易所/ | 交易所相关笔记和示例 | ❌ |
| 经典三连/ | 并发经典模式骨架 | ❌ |

子目录也有独立 go.mod：`s3/`、`patterns/circuit-breaker/`、`patterns/rate-limiting/`、`mutexes/crypto-reader/`、`mutexes/distributed-db/` 等。运行示例前先 cd 到对应目录。

## 常用命令

```bash
go run main.go                    # 运行示例（先 cd 到示例目录）
go test -race -v ./...            # 测试（并发代码必须加 -race）
go test -bench . ./...            # 基准测试
go run -race main.go              # 竞态检测单文件
```

## 注意事项

- 多个独立 go.mod：修改前确认当前在哪个模块目录下
- 部分示例故意包含竞态条件/死锁，用于教学演示
- 教学仓库，非生产代码——不需要生产级错误处理
- 并发测试必须加 `-race` flag
