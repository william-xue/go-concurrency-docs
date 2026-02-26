package main

import (
	"fmt"
	"sync"
	"time"
)

// worker 函数就是我们的“工人”，现在它需要一个 WaitGroup 来通知老板它完成了任务
func worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
    // 确保在函数退出前，wg.Done() 被调用
	defer wg.Done() 

	// 这个 for range 循环会一直从 jobs 通道里接收任务，直到通道被 close
	for j := range jobs {
		fmt.Printf("工人 %d: 开始处理 %s\n", id, j)
		time.Sleep(1 * time.Second) // 模拟一个耗时操作
		fmt.Printf("工人 %d: 完成 %s\n", id, j)
	}
}

func main() {
	// 1. 创建任务通道
	jobs := make(chan string, 100)
	
	// 2. 创建 WaitGroup
	var wg sync.WaitGroup

	// 3. 启动 5 个“工人”，每个工人都是一个 Goroutine
	for w := 1; w <= 5; w++ {
		// 每次启动一个工人，都把 WaitGroup 的计数器加 1
		wg.Add(1)
		// 注意这里，我们不再需要 results 通道了，因为 WaitGroup 负责同步
		go worker(w, jobs, &wg)
	}

	// 4. 把 100 个任务发送到任务通道里
	for j := 0; j < 100; j++ {
		jobs <- fmt.Sprintf("下载任务%d", j)
	}
	// 关闭任务通道，表示没有新任务了
	close(jobs)

	// 5. 等待所有 Goroutine 完成
	// 这行代码会阻塞，直到所有的 wg.Done() 都被调用
	wg.Wait()
	
	fmt.Println("所有任务都已完成。")
}