package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 这是一个独立的计算任务
func calcPowerFlow(timePoint int, resultChan chan<- float64) {
	// 模拟大规模矩阵求解的耗时操作
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second) 
	
	// 算出一个断面的损耗结果
	loss := rand.Float64() * 100 
	
	// 【灵魂操作】：算完直接把结果扔进传送带，不用加锁！
	resultChan <- loss 
}

func main() {
	totalSnapshots := 96
	// 创建一条容量为 96 的传送带 (Channel)
	results := make(chan float64, totalSnapshots)

	fmt.Println("开始并行计算 96 个断面的潮流...")

	// 1. 扇出 (Fan-out)：瞬间砸出 96 个 Goroutine
	for i := 1; i <= totalSnapshots; i++ {
		go calcPowerFlow(i, results)
	}

	maxLoss := 0.0
	// 2. 扇入 (Fan-in)：站在传送带尽头，收齐 96 个结果
	for i := 1; i <= totalSnapshots; i++ {
		// <-results 会天然阻塞，直到传送带上有数据过来
		currentLoss := <-results 
		
		// 汇总逻辑完全在单线程(main)里，绝对安全
		if currentLoss > maxLoss {
			maxLoss = currentLoss
		}
	}

	fmt.Printf("全部计算完成，全天最大损耗为: %.2f\n", maxLoss)
}