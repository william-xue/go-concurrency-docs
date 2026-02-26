package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func calcPowerFlow(ctx context.Context, timePoint int, resultChan chan<- float64) {
	// 模拟分步计算：每步检查 context 是否已取消
	steps := rand.Intn(5) + 1 // 1~5 步，每步 1 秒
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			// 收到取消信号，立即停止计算，释放 CPU
			fmt.Printf("  [断面%02d] 第%d步被取消，停止计算\n", timePoint, i+1)
			return
		case <-time.After(1 * time.Second):
			// 模拟一步计算
		}
	}

	loss := rand.Float64() * 100
	// 写之前也检查一下，避免往已无人读的 channel 写入
	select {
	case resultChan <- loss:
	case <-ctx.Done():
		return
	}
}

func main() {
	totalSnapshots := 96

	// 核心：用 WithTimeout 创建一个 3 秒后自动取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // 兜底：main 退出时确保取消

	results := make(chan float64, totalSnapshots)

	fmt.Println("开始并行计算 96 个断面的潮流（3秒超时，context 级联取消）...")

	for i := 1; i <= totalSnapshots; i++ {
		go calcPowerFlow(ctx, i, results)
	}

	maxLoss := 0.0
	collected := 0

	for collected < totalSnapshots {
		select {
		case loss := <-results:
			collected++
			if loss > maxLoss {
				maxLoss = loss
			}
		case <-ctx.Done():
			fmt.Printf("\n⏰ 超时熔断！已收集 %d/%d 个结果\n", collected, totalSnapshots)
			fmt.Printf("基于已有数据，当前最大损耗为: %.2f\n", maxLoss)
			// cancel() 已被 defer 调用，所有 goroutine 会收到 ctx.Done() 信号
			// 等一小会儿让取消日志打印出来
			time.Sleep(500 * time.Millisecond)
			return
		}
	}

	fmt.Printf("全部计算完成，全天最大损耗为: %.2f\n", maxLoss)
}
