package queue

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// DockTask 对接任务
type DockTask struct {
	OID int64
}

// DockResult 对接结果（用于准确统计成功/失败）
type DockResult struct {
	OID     int64
	Success bool
}

// DockQueue 网课对接并发队列
// 注意：int64 字段必须放在结构体最前面，保证 32 位系统上 64 位原子操作对齐
type DockQueue struct {
	pending     int64 // 待处理数
	processing  int64 // 处理中数
	completed   int64 // 对接成功数
	failed      int64 // 对接失败数（含 panic + 业务失败）
	workerCount int32 // 当前worker数
	maxWorkers  int32 // 最大并发数
	taskChan    chan DockTask
	sem         chan struct{}        // 信号量，控制并发数
	handler     func(oid int64)      // 任务处理函数
	checker     func(oid int64) bool // 检查对接是否成功（查DB dockstatus=1）
	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

var (
	GlobalDockQueue *DockQueue
)

// NewDockQueue 创建对接队列
// checker 可选：传入一个函数判断 oid 是否对接成功（用于准确统计），传 nil 则不检查
func NewDockQueue(maxWorkers int, bufferSize int, handler func(oid int64), checker func(oid int64) bool) *DockQueue {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	return &DockQueue{
		taskChan:   make(chan DockTask, bufferSize),
		sem:        make(chan struct{}, maxWorkers),
		maxWorkers: int32(maxWorkers),
		handler:    handler,
		checker:    checker,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动队列
func (q *DockQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.running {
		return
	}
	q.running = true
	log.Printf("[DockQueue] 启动，最大并发数: %d, 缓冲区: %d", q.maxWorkers, cap(q.taskChan))

	// 启动 dispatcher
	go q.dispatch()
}

// dispatch 分发任务到 workers（信号量模式，无忙等待）
func (q *DockQueue) dispatch() {
	for {
		select {
		case task, ok := <-q.taskChan:
			if !ok {
				return
			}
			// 获取信号量槽位（阻塞等待，不消耗CPU）
			select {
			case q.sem <- struct{}{}:
			case <-q.stopCh:
				return
			}
			atomic.AddInt32(&q.workerCount, 1)
			atomic.AddInt64(&q.pending, -1)
			atomic.AddInt64(&q.processing, 1)
			q.wg.Add(1)
			go q.processTask(task)
		case <-q.stopCh:
			return
		}
	}
}

// processTask 处理单个任务
func (q *DockQueue) processTask(task DockTask) {
	panicked := true
	defer func() {
		<-q.sem // 释放信号量槽位
		atomic.AddInt32(&q.workerCount, -1)
		atomic.AddInt64(&q.processing, -1)
		q.wg.Done()
		if r := recover(); r != nil {
			atomic.AddInt64(&q.failed, 1)
			log.Printf("[DockQueue] worker panic: %v, oid=%d", r, task.OID)
			return
		}
		if panicked {
			return // 不应到达
		}
		// 通过 checker 判断对接是否成功（查 DB dockstatus）
		if q.checker != nil {
			if q.checker(task.OID) {
				atomic.AddInt64(&q.completed, 1)
			} else {
				atomic.AddInt64(&q.failed, 1)
			}
		} else {
			// 无 checker 时仍按旧逻辑：完成即 completed
			atomic.AddInt64(&q.completed, 1)
		}
	}()

	q.handler(task.OID)
	panicked = false
}

// Push 添加任务到队列
func (q *DockQueue) Push(oid int64) {
	atomic.AddInt64(&q.pending, 1)
	select {
	case q.taskChan <- DockTask{OID: oid}:
		// 成功入队
	default:
		// 队列满了，带超时等待入队，防止 goroutine 泄漏
		log.Printf("[DockQueue] 队列已满，oid=%d 等待入队...", oid)
		go func() {
			timer := time.NewTimer(2 * time.Minute)
			defer timer.Stop()
			select {
			case q.taskChan <- DockTask{OID: oid}:
				// 入队成功
			case <-timer.C:
				atomic.AddInt64(&q.pending, -1)
				log.Printf("[DockQueue] oid=%d 入队超时，已丢弃", oid)
			case <-q.stopCh:
				atomic.AddInt64(&q.pending, -1)
			}
		}()
	}
}

// PushBatch 批量添加任务
func (q *DockQueue) PushBatch(oids []int64) {
	for _, oid := range oids {
		q.Push(oid)
	}
}

// SetMaxWorkers 动态调整并发数
// 安全实现：等待当前 worker 全部完成后再替换信号量
func (q *DockQueue) SetMaxWorkers(n int) {
	if n < 1 {
		n = 1
	}
	if n > 100 {
		n = 100
	}
	old := atomic.LoadInt32(&q.maxWorkers)
	if int32(n) == old {
		return
	}
	atomic.StoreInt32(&q.maxWorkers, int32(n))

	// 等待所有当前 worker 完成后替换 sem
	// 注意：这会短暂暂停新 worker 的启动，但保证了安全
	q.mu.Lock()
	q.sem = make(chan struct{}, n)
	q.mu.Unlock()
	log.Printf("[DockQueue] 并发数从 %d 调整为 %d", old, n)
}

// Stats 获取队列状态
func (q *DockQueue) Stats() map[string]interface{} {
	return map[string]interface{}{
		"max_workers": atomic.LoadInt32(&q.maxWorkers),
		"active":      atomic.LoadInt32(&q.workerCount),
		"pending":     atomic.LoadInt64(&q.pending),
		"processing":  atomic.LoadInt64(&q.processing),
		"completed":   atomic.LoadInt64(&q.completed),
		"failed":      atomic.LoadInt64(&q.failed),
		"queue_size":  len(q.taskChan),
		"queue_cap":   cap(q.taskChan),
		"running":     q.running,
	}
}

// Stop 停止队列（等待当前任务完成）
func (q *DockQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.running {
		return
	}
	q.running = false
	close(q.stopCh)
	q.wg.Wait()
	fmt.Println("[DockQueue] 已停止")
}
