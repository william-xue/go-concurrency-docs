package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func calcPowerFlow(ctx context.Context, timePoint int) (float64, error) {
	steps := rand.Intn(5) + 1
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
	return rand.Float64() * 100, nil
}

func main() {
	totalSnapshots := 96
	maxWorkers := 10

	// 【核心】signal.NotifyContext：收到 Ctrl+C / kill 时自动取消 context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 在信号 context 基础上再叠加超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	fmt.Printf("开始并行计算 %d 个断面（Ctrl+C 优雅关闭）...\n", totalSnapshots)
	fmt.Println("提示：按 Ctrl+C 可随时优雅停止")

	type result struct {
		timePoint int
		loss      float64
		err       error
	}

	results := make(chan result, totalSnapshots)
	sem := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup
	for i := 1; i <= totalSnapshots; i++ {
		wg.Add(1)
		go func(tp int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			loss, err := calcPowerFlow(ctx, tp)
			results <- result{tp, loss, err}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集阶段
	var (
		maxLoss   float64
		succeeded int
		cancelled int
	)

	for r := range results {
		if r.err != nil {
			cancelled++
			continue
		}
		succeeded++
		if r.loss > maxLoss {
			maxLoss = r.loss
		}
	}

	// 【核心】无论怎么退出，都走到这里打印最终状态
	reason := "全部完成"
	if ctx.Err() == context.DeadlineExceeded {
		reason = "超时退出"
	} else if ctx.Err() == context.Canceled {
		reason = "收到终止信号 (Ctrl+C)"
	}

	fmt.Printf("\n🏁 退出原因: %s\n", reason)
	fmt.Printf("   成功: %d | 取消: %d | 总计: %d\n", succeeded, cancelled, totalSnapshots)
	fmt.Printf("   最大损耗: %.2f\n", maxLoss)

	os.Exit(0)
}
