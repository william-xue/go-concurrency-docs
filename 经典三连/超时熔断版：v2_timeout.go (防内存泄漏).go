package main

import (
	"fmt"
	"math/rand"
	"time"
)

func calcPowerFlow(timePoint int, resultChan chan<- float64) {
	// 模拟耗时：0~4秒，故意让部分任务超过3秒超时线
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	loss := rand.Float64() * 100

	// 注意：超时后 main 不再读 channel，但 buffer 能兜住，goroutine 不会永久阻塞
	resultChan <- loss
}

func main() {
	totalSnapshots := 96
	results := make(chan float64, totalSnapshots)

	fmt.Println("开始并行计算 96 个断面的潮流（3秒超时）...")

	for i := 1; i <= totalSnapshots; i++ {
		go calcPowerFlow(i, results)
	}

	// 超时倒计时：3秒后触发
	timeout := time.After(3 * time.Second)

	maxLoss := 0.0
	collected := 0

	for collected < totalSnapshots {
		select {
		case loss := <-results:
			// 正常收到结果
			collected++
			if loss > maxLoss {
				maxLoss = loss
			}
		case <-timeout:
			// 超时熔断：不再等剩余任务
			fmt.Printf("⏰ 超时熔断！已收集 %d/%d 个结果\n", collected, totalSnapshots)
			fmt.Printf("基于已有数据，当前最大损耗为: %.2f\n", maxLoss)
			return
		}
	}

	// 全部按时完成（不太可能走到这，但逻辑完整）
	fmt.Printf("全部计算完成，全天最大损耗为: %.2f\n", maxLoss)
}
