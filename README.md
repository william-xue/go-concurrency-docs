# Go 并发编程学习资料集合

这是一个全面的Go并发编程学习资料和示例代码集合，包含了从基础概念到高级模式的完整知识体系。

## 📚 内容概览

### 🎯 核心主题

- **Goroutines** - Go协程的基础使用和高级技巧
- **Channels** - 通道的各种使用模式和最佳实践
- **Mutexes** - 互斥锁和读写锁的应用场景
- **Atomics** - 原子操作的性能优化
- **Patterns** - 并发编程的设计模式
- **WaitGroups** - 等待组的正确使用方法
- **Select** - 多路复用的强大功能
- **Context** - 上下文管理和取消机制

### 📁 项目结构

```
├── concurrency-master/          # 主要的并发编程教程和示例
│   ├── go-routines/            # Goroutines 相关示例
│   ├── channels/               # Channels 使用模式
│   ├── mutexes/                # 互斥锁和同步原语
│   ├── atomics/                # 原子操作示例
│   ├── patterns/               # 并发设计模式
│   ├── waitgroups/             # WaitGroup 使用方法
│   ├── select/                 # Select 语句示例
│   └── ...
├── concurrency-programming-via-go-code-master/  # 额外的并发编程示例
├── my-concurrency/             # 个人练习代码
└── golang_shared_memory_*      # 共享内存相关资料
```

## 🚀 快速开始

1. **克隆仓库**
   ```bash
   git clone https://github.com/william-xue/go-concurrency-docs.git
   cd go-concurrency-docs
   ```

2. **运行示例**
   ```bash
   cd concurrency-master/go-routines/simple
   go run main.go
   ```

3. **学习路径建议**
   - 从 `intro/` 开始了解基础概念
   - 学习 `go-routines/` 掌握协程使用
   - 深入 `channels/` 理解通道机制
   - 探索 `patterns/` 学习设计模式
   - 实践 `mutexes/` 和 `atomics/` 优化性能

## 📖 学习资源

### 视频教程
- [Concurrency in Go #1 -- Introduction to Concurrency](https://youtu.be/_uQgGS_VIXM)
- [Concurrency in Go #2 -- WaitGroups Part 1](https://youtu.be/srb6fbioEY4)
- [Concurrency in Go #3 -- WaitGroups Part 2](https://youtu.be/zAMUKb6fCO0)
- 更多视频请查看 `concurrency-master/README.md`

### 重点主题

#### 1. Goroutines 基础
- 协程的创建和生命周期
- 协程泄漏的预防
- 协程间的通信机制

#### 2. Channels 模式
- 缓冲和非缓冲通道
- 通道的关闭和检测
- Fan-in/Fan-out 模式
- Pipeline 模式

#### 3. 同步原语
- Mutex 和 RWMutex
- 原子操作的使用场景
- WaitGroup 的正确用法
- Once 的单例模式

#### 4. 高级模式
- Context 的传播和取消
- 错误处理策略
- 超时和截止时间
- 优雅关闭

## 🛠 实践建议

1. **从简单开始** - 先理解基础概念再进入复杂模式
2. **动手实践** - 每个示例都要亲自运行和修改
3. **性能测试** - 使用 benchmark 测试不同方案的性能
4. **避免陷阱** - 注意死锁、竞态条件等常见问题

## 📝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个学习资料集合！

## 📄 许可证

本项目遵循各子项目的原始许可证。详情请查看各目录下的 LICENSE 文件。

---

**Happy Coding! 🎉**

让我们一起掌握Go并发编程的精髓！