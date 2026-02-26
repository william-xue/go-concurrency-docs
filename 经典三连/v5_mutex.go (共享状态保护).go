package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// 【核心】共享状态：多个 goroutine 要同时读写这个结构
type PowerStats struct {
	mu       sync.RWMutex // 读写锁：多读单写，比 Mutex 性能好
	maxLoss  float64
	total    float64
	count    int
	failures int
}

// 写操作：独占锁
func (s *PowerStats) Record(loss float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	s.total += loss
	if loss > s.maxLoss {
		s.maxLoss = loss
	}
}

func (s *PowerStats) RecordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures++
}

// 读操作：共享锁（多个 goroutine 可同时读）
func (s *PowerStats) Snapshot() (max, avg float64, count, failures int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	avg = 0
	if s.count > 0 {
		avg = s.total / float64(s.count)
	}
	return s.maxLoss, avg, s.count, s.failures
}

func calcPowerFlow(ctx context.Context, timePoint int, stats *PowerStats) {
	steps := rand.Intn(5) + 1
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			stats.RecordFailure()
			return
		case <-time.After(1 * time.Second):
		}
	}
	loss := rand.Float64() * 100
	stats.Record(loss)
}

func main() {
	totalSnapshots := 96
	maxWorkers := 10

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := &PowerStats{}
	sem := make(chan struct{}, maxWorkers)

	fmt.Printf("开始并行计算 %d 个断面（RWMutex 保护共享统计）...\n", totalSnapshots)

	// 【新增】实时监控：每秒打印一次进度（读锁，不阻塞写入）
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				max, avg, count, failures := stats.Snapshot()
				fmt.Printf("  📊 进度: %d 完成, %d 失败, 当前最大=%.1f, 平均=%.1f\n",
					count, failures, max, avg)
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 1; i <= totalSnapshots; i++ {
		wg.Add(1)
		go func(tp int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			calcPowerFlow(ctx, tp, stats)
		}(i)
	}

	wg.Wait()
	cancel() // 停止监控 goroutine

	max, avg, count, failures := stats.Snapshot()
	fmt.Printf("\n📋 最终报告:\n")
	fmt.Printf("   完成: %d | 取消: %d | 总计: %d\n", count, failures, totalSnapshots)
	fmt.Printf("   最大损耗: %.2f | 平均损耗: %.2f\n", max, avg)
}
