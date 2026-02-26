package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func calcPowerFlow(ctx context.Context, timePoint int, resultChan chan<- float64) {
	steps := rand.Intn(5) + 1
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("  [断面%02d] 被取消\n", timePoint)
			return
		case <-time.After(1 * time.Second):
		}
	}

	loss := rand.Float64() * 100
	select {
	case resultChan <- loss:
	case <-ctx.Done():
		return
	}
}

func main() {
	totalSnapshots := 96
	maxWorkers := 10 // 【核心】同时最多跑 10 个 goroutine

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := make(chan float64, totalSnapshots)

	fmt.Printf("开始并行计算 %d 个断面（最多 %d 个并发 worker，5秒超时）...\n", totalSnapshots, maxWorkers)

	// 【核心】信号量模式：用 buffered channel 当令牌桶
	// 容量 = maxWorkers，满了就阻塞，天然限流
	sem := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup
	for i := 1; i <= totalSnapshots; i++ {
		wg.Add(1)
		go func(tp int) {
			defer wg.Done()
			sem <- struct{}{} // 取令牌（满了就排队）
			defer func() { <-sem }() // 还令牌

			calcPowerFlow(ctx, tp, results)
		}(i)
	}

	// 单独 goroutine 等所有 worker 完成后关闭 results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	maxLoss := 0.0
	collected := 0

	for {
		select {
		case loss, ok := <-results:
			if !ok {
				// channel 已关闭，所有任务完成
				fmt.Printf("\n✅ 全部完成！收集 %d/%d 个结果\n", collected, totalSnapshots)
				fmt.Printf("全天最大损耗为: %.2f\n", maxLoss)
				return
			}
			collected++
			if loss > maxLoss {
				maxLoss = loss
			}
		case <-ctx.Done():
			fmt.Printf("\n⏰ 超时熔断！已收集 %d/%d 个结果\n", collected, totalSnapshots)
			fmt.Printf("基于已有数据，当前最大损耗为: %.2f\n", maxLoss)
			return
		}
	}
}
