package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// StockData 股票数据结构
type StockData struct {
	Symbol    string    // 股票代码
	Price     float64   // 当前价格
	Volume    int64     // 成交量
	Timestamp time.Time // 时间戳
	Exchange  string    // 交易所
}

// RingBuffer 高性能环形缓冲区 - 这是我们的共享内存核心
type RingBuffer struct {
	buffer   []StockData   // 数据缓冲区
	size     int           // 缓冲区大小
	writePos int64         // 写入位置（使用atomic操作）
	readPos  int64         // 读取位置（使用atomic操作）
	mutex    sync.RWMutex  // 读写锁保护共享内存
	pool     *sync.Pool    // 内存池优化
}

// NewRingBuffer 创建新的环形缓冲区
func NewRingBuffer(size int) *RingBuffer {
	rb := &RingBuffer{
		buffer: make([]StockData, size),
		size:   size,
		pool: &sync.Pool{
			New: func() interface{} {
				return &StockData{}
			},
		},
	}
	return rb
}

// Write 写入数据到共享内存（生产者使用）
func (rb *RingBuffer) Write(data StockData) bool {
	// 获取写锁
	rb.mutex.Lock()
	defer rb.mutex.Unlock()
	
	currentWrite := atomic.LoadInt64(&rb.writePos)
	currentRead := atomic.LoadInt64(&rb.readPos)
	
	// 检查缓冲区是否已满
	if (currentWrite+1)%int64(rb.size) == currentRead {
		return false // 缓冲区已满
	}
	
	// 写入数据
	rb.buffer[currentWrite] = data
	atomic.StoreInt64(&rb.writePos, (currentWrite+1)%int64(rb.size))
	
	return true
}

// Read 从共享内存读取数据（消费者使用）
func (rb *RingBuffer) Read() (*StockData, bool) {
	// 获取读锁
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	currentRead := atomic.LoadInt64(&rb.readPos)
	currentWrite := atomic.LoadInt64(&rb.writePos)
	
	// 检查是否有数据可读
	if currentRead == currentWrite {
		return nil, false // 没有数据
	}
	
	// 从内存池获取对象
	data := rb.pool.Get().(*StockData)
	*data = rb.buffer[currentRead]
	
	atomic.StoreInt64(&rb.readPos, (currentRead+1)%int64(rb.size))
	
	return data, true
}

// ReturnToPool 将对象返回到内存池
func (rb *RingBuffer) ReturnToPool(data *StockData) {
	rb.pool.Put(data)
}

// GetStats 获取缓冲区统计信息
func (rb *RingBuffer) GetStats() (int64, int64, int) {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	
	writePos := atomic.LoadInt64(&rb.writePos)
	readPos := atomic.LoadInt64(&rb.readPos)
	
	var used int
	if writePos >= readPos {
		used = int(writePos - readPos)
	} else {
		used = int(int64(rb.size) - readPos + writePos)
	}
	
	return writePos, readPos, used
}

// StockExchange 股票交易所模拟器（数据生产者）
type StockExchange struct {
	name     string
	symbols  []string
	buffer   *RingBuffer
	produced int64 // 原子计数器
}

// NewStockExchange 创建新的交易所
func NewStockExchange(name string, symbols []string, buffer *RingBuffer) *StockExchange {
	return &StockExchange{
		name:    name,
		symbols: symbols,
		buffer:  buffer,
	}
}

// Start 启动交易所数据生成
func (se *StockExchange) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	ticker := time.NewTicker(time.Microsecond * 100) // 高频数据生成
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🏢 交易所 %s 停止运行\n", se.name)
			return
		case <-ticker.C:
			// 生成随机股票数据
			symbol := se.symbols[rand.Intn(len(se.symbols))]
			data := StockData{
				Symbol:    symbol,
				Price:     100 + rand.Float64()*50, // 100-150之间的随机价格
				Volume:    int64(rand.Intn(10000) + 1000),
				Timestamp: time.Now(),
				Exchange:  se.name,
			}
			
			// 写入共享内存
			if se.buffer.Write(data) {
				atomic.AddInt64(&se.produced, 1)
			}
		}
	}
}

// GetProduced 获取生产的数据量
func (se *StockExchange) GetProduced() int64 {
	return atomic.LoadInt64(&se.produced)
}

// PriceMonitor 价格监控器（消费者1）
type PriceMonitor struct {
	name      string
	buffer    *RingBuffer
	processed int64
	alerts    int64
}

// Start 启动价格监控
func (pm *PriceMonitor) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	ticker := time.NewTicker(time.Microsecond * 50)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("📊 价格监控器停止运行\n")
			return
		case <-ticker.C:
			// 从共享内存读取数据
			if data, ok := pm.buffer.Read(); ok {
				atomic.AddInt64(&pm.processed, 1)
				
				// 价格异常检测
				if data.Price > 140 {
					atomic.AddInt64(&pm.alerts, 1)
					// 这里可以触发实际的告警逻辑
				}
				
				// 返回到内存池
				pm.buffer.ReturnToPool(data)
			}
		}
	}
}

// GetStats 获取监控统计
func (pm *PriceMonitor) GetStats() (int64, int64) {
	return atomic.LoadInt64(&pm.processed), atomic.LoadInt64(&pm.alerts)
}

// TechnicalAnalyzer 技术分析器（消费者2）
type TechnicalAnalyzer struct {
	name         string
	buffer       *RingBuffer
	processed    int64
	priceHistory map[string][]float64 // 价格历史数据
	mutex        sync.RWMutex
}

// NewTechnicalAnalyzer 创建技术分析器
func NewTechnicalAnalyzer(buffer *RingBuffer) *TechnicalAnalyzer {
	return &TechnicalAnalyzer{
		name:         "技术分析器",
		buffer:       buffer,
		priceHistory: make(map[string][]float64),
	}
}

// Start 启动技术分析
func (ta *TechnicalAnalyzer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	ticker := time.NewTicker(time.Microsecond * 80)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("📈 技术分析器停止运行\n")
			return
		case <-ticker.C:
			if data, ok := ta.buffer.Read(); ok {
				atomic.AddInt64(&ta.processed, 1)
				
				// 更新价格历史
				ta.updatePriceHistory(data.Symbol, data.Price)
				
				// 计算移动平均线
				ta.calculateMovingAverage(data.Symbol)
				
				ta.buffer.ReturnToPool(data)
			}
		}
	}
}

// updatePriceHistory 更新价格历史
func (ta *TechnicalAnalyzer) updatePriceHistory(symbol string, price float64) {
	ta.mutex.Lock()
	defer ta.mutex.Unlock()
	
	if _, exists := ta.priceHistory[symbol]; !exists {
		ta.priceHistory[symbol] = make([]float64, 0, 100)
	}
	
	ta.priceHistory[symbol] = append(ta.priceHistory[symbol], price)
	
	// 保持最近100个价格点
	if len(ta.priceHistory[symbol]) > 100 {
		ta.priceHistory[symbol] = ta.priceHistory[symbol][1:]
	}
}

// calculateMovingAverage 计算移动平均线
func (ta *TechnicalAnalyzer) calculateMovingAverage(symbol string) float64 {
	ta.mutex.RLock()
	defer ta.mutex.RUnlock()
	
	prices, exists := ta.priceHistory[symbol]
	if !exists || len(prices) < 20 {
		return 0
	}
	
	// 计算20日移动平均线
	sum := 0.0
	for i := len(prices) - 20; i < len(prices); i++ {
		sum += prices[i]
	}
	
	return sum / 20
}

// GetProcessed 获取处理数量
func (ta *TechnicalAnalyzer) GetProcessed() int64 {
	return atomic.LoadInt64(&ta.processed)
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	buffer    *RingBuffer
	exchanges []*StockExchange
	monitor   *PriceMonitor
	analyzer  *TechnicalAnalyzer
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(buffer *RingBuffer, exchanges []*StockExchange, 
	monitor *PriceMonitor, analyzer *TechnicalAnalyzer) *PerformanceMonitor {
	return &PerformanceMonitor{
		buffer:    buffer,
		exchanges: exchanges,
		monitor:   monitor,
		analyzer:  analyzer,
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	
	startTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("📊 性能监控器停止运行\n")
			return
		case <-ticker.C:
			pm.printStats(time.Since(startTime))
		}
	}
}

// printStats 打印统计信息
func (pm *PerformanceMonitor) printStats(duration time.Duration) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("🚀 实时股票数据处理系统 - 运行时间: %.1f秒\n", duration.Seconds())
	fmt.Printf(strings.Repeat("=", 80) + "\n")
	
	// 缓冲区状态
	writePos, readPos, used := pm.buffer.GetStats()
	fmt.Printf("💾 共享内存缓冲区状态:\n")
	fmt.Printf("   写入位置: %d | 读取位置: %d | 使用量: %d/%d (%.1f%%)\n", 
		writePos, readPos, used, len(pm.buffer.buffer), 
		float64(used)/float64(len(pm.buffer.buffer))*100)
	
	// 生产者统计
	fmt.Printf("\n📈 数据生产者统计:\n")
	totalProduced := int64(0)
	for _, exchange := range pm.exchanges {
		produced := exchange.GetProduced()
		totalProduced += produced
		rate := float64(produced) / duration.Seconds()
		fmt.Printf("   %s: %d 条数据 (%.0f 条/秒)\n", exchange.name, produced, rate)
	}
	
	// 消费者统计
	fmt.Printf("\n📊 数据消费者统计:\n")
	monitorProcessed, alerts := pm.monitor.GetStats()
	analyzerProcessed := pm.analyzer.GetProcessed()
	
	fmt.Printf("   价格监控器: %d 条处理 (%.0f 条/秒) | 告警: %d 次\n", 
		monitorProcessed, float64(monitorProcessed)/duration.Seconds(), alerts)
	fmt.Printf("   技术分析器: %d 条处理 (%.0f 条/秒)\n", 
		analyzerProcessed, float64(analyzerProcessed)/duration.Seconds())
	
	// 系统性能
	fmt.Printf("\n⚡ 系统性能指标:\n")
	fmt.Printf("   总生产量: %d 条/秒\n", int64(float64(totalProduced)/duration.Seconds()))
	fmt.Printf("   总消费量: %d 条/秒\n", int64(float64(monitorProcessed+analyzerProcessed)/duration.Seconds()))
	fmt.Printf("   Goroutine数量: %d\n", runtime.NumGoroutine())
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("   内存使用: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("   GC次数: %d\n", m.NumGC)
}

func main() {
	fmt.Printf("🚀 启动高性能实时股票数据处理系统\n")
	fmt.Printf("💡 这个例子展示了Go语言中共享内存的强大应用\n\n")
	
	// 创建共享内存缓冲区
	bufferSize := 10000
	sharedBuffer := NewRingBuffer(bufferSize)
	
	// 股票代码列表
	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN", "META", "NVDA", "NFLX"}
	
	// 创建多个交易所（生产者）
	exchanges := []*StockExchange{
		NewStockExchange("纳斯达克", symbols[:4], sharedBuffer),
		NewStockExchange("纽约证交所", symbols[4:], sharedBuffer),
		NewStockExchange("上海证交所", []string{"000001", "000002", "600000", "600036"}, sharedBuffer),
	}
	
	// 创建消费者
	priceMonitor := &PriceMonitor{name: "价格监控器", buffer: sharedBuffer}
	techAnalyzer := NewTechnicalAnalyzer(sharedBuffer)
	
	// 创建性能监控器
	perfMonitor := NewPerformanceMonitor(sharedBuffer, exchanges, priceMonitor, techAnalyzer)
	
	// 创建上下文和等待组
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	
	// 启动所有生产者
	for _, exchange := range exchanges {
		wg.Add(1)
		go exchange.Start(ctx, &wg)
	}
	
	// 启动所有消费者
	wg.Add(1)
	go priceMonitor.Start(ctx, &wg)
	
	wg.Add(1)
	go techAnalyzer.Start(ctx, &wg)
	
	// 启动性能监控
	go perfMonitor.Start(ctx)
	
	// 运行10秒后优雅关闭
	fmt.Printf("⏰ 系统将运行10秒，然后优雅关闭...\n")
	time.Sleep(10 * time.Second)
	
	fmt.Printf("\n🛑 开始优雅关闭系统...\n")
	cancel()
	wg.Wait()
	
	fmt.Printf("\n✅ 系统已安全关闭！\n")
	fmt.Printf("\n🎯 这个例子展示了:\n")
	fmt.Printf("   ✓ 高性能环形缓冲区作为共享内存\n")
	fmt.Printf("   ✓ 多生产者-多消费者模式\n")
	fmt.Printf("   ✓ 原子操作和读写锁的协同使用\n")
	fmt.Printf("   ✓ 内存池优化减少GC压力\n")
	fmt.Printf("   ✓ 实时性能监控\n")
	fmt.Printf("   ✓ 优雅关闭机制\n")
	fmt.Printf("\n💪 这就是Go语言共享内存的强大威力！\n")
}
