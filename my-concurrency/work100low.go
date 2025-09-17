package main

import (
	"fmt"
	"time"
)

// worker 函数就是我们的“工人”，它会从 jobs 通道里接收任务，并把结果发送到 results 通道
func worker(id int, jobs <-chan string, results chan<- string) {
	// 这个 for range 循环会一直从 jobs 通道里接收任务，直到通道被 close
	for j := range jobs {
		fmt.Printf("工人 %d: 开始处理 %s\n", id, j)
		time.Sleep(1 * time.Second) // 模拟一个耗时操作
		fmt.Printf("工人 %d: 完成 %s\n", id, j)

		// 把处理结果发送到 results 通道
		results <- fmt.Sprintf("结果: %s 完成", j)
	}
}

func main() {
	// 1. 创建任务通道和结果通道
	jobs := make(chan string, 100)
	results := make(chan string, 100)

	// 2. 启动 5 个“工人”，每个工人都是一个 Goroutine
	for w := 1; w <= 5; w++ {
		go worker(w, jobs, results)
	}

	// 3. 把 100 个任务发送到任务通道里
	for j := 0; j < 100; j++ {
		jobs <- fmt.Sprintf("下载任务%d", j)
	}
	// 关闭任务通道，表示没有新任务了
	close(jobs)

	// 4. 从结果通道里接收所有结果，确保所有任务都完成
	for a := 0; a < 100; a++ {
		<-results
	}

	fmt.Println("所有任务都已完成。")
}
