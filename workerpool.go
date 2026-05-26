// Package nets provides a lightweight goroutine worker pool for concurrent task execution.
// It manages a fixed number of worker goroutines that process tasks from buffered channels.
// Package nets 提供了一个轻量级 goroutine 工作池，用于并发任务执行。
// 它管理固定数量的工作协程，从缓冲通道中处理任务。
package nets

import (
	"context"
	"errors"
	"hash/fnv"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

// Common errors returned by WorkerPool methods.
// WorkerPool 方法返回的常见错误。
var (
	ErrPoolClosed = errors.New("worker pool is closed")
)

// WorkerPoolStats contains statistics about the worker pool's current state.
// WorkerPoolStats 包含工作池当前状态的统计信息。
type WorkerPoolStats struct {
	ActiveWorkers  int32
	PendingTasks   int
	Capacity       int32
	TotalSubmitted int64
	TotalCompleted int64
}

// WorkerPool manages a pool of goroutine workers that execute submitted tasks concurrently.
// It uses done channel pattern to prevent send-on-closed-channel panics.
//
// Features:
//   - Submit(): Round-robin distribution for general tasks
//   - SubmitWithWorker(): Hash binding for ordered execution (same workerId → same worker)
//   - SubmitCtx(): Context-aware submission with cancellation support
//   - HashWorkerId(): Utility to convert string to valid workerId
//
// WorkerPool 管理一组 goroutine 工作协程，并发执行提交的任务。
// 使用 done channel 模式防止 send-on-closed-channel panic。
//
// 特性：
//   - Submit(): 轮询分配，适用于普通任务
//   - SubmitWithWorker(): 哈希绑定，保证顺序执行（相同 workerId → 相同 worker）
//   - SubmitCtx(): 支持 context 取消的提交
//   - HashWorkerId(): 将字符串转换为有效 workerId 的工具方法
type WorkerPool struct {
	maxWorkers     int32
	activeWorkers  int32
	workers        []chan func() // Array of worker channels / worker 通道数组
	done           chan struct{} // Signal channel for shutdown / 关闭信号通道
	wg             sync.WaitGroup
	nextIdx        uint64 // Atomic counter for round-robin / 轮询原子计数器
	totalSubmitted int64
	totalCompleted int64
}

// Singleton instance variables / 单例实例变量
var instanceWorkerPool *WorkerPool
var instanceWorkerPoolOnce = sync.Once{}

// GetInstanceWorkerPool returns the singleton instance of WorkerPool.
// It initializes the pool on first call using configuration from defaultServer.
// If WorkerPoolSize is not configured (<= 0), it defaults to runtime.NumCPU() * 10.
//
// GetInstanceWorkerPool 返回 WorkerPool 的单例实例。
// 它在第一次调用时使用 defaultServer 的配置初始化池。
// 如果 WorkerPoolSize 未配置（<= 0），则默认为 runtime.NumCPU() * 10。
func GetInstanceWorkerPool() *WorkerPool {
	instanceWorkerPoolOnce.Do(func() {
		poolSize := defaultServer.AppConf.WorkerPoolSize
		if poolSize <= 0 {
			poolSize = runtime.NumCPU() * 10
		}
		instanceWorkerPool = NewWorkerPool(poolSize, defaultServer.AppConf.WorkerTaskMaxLen)
	})
	return instanceWorkerPool
}

// NewWorkerPool creates a new WorkerPool with the specified number of workers and task channel capacity.
// It immediately starts all worker goroutines which will process tasks submitted to the pool.
//
// NewWorkerPool 使用指定的工作协程数量和任务通道容量创建一个新的 WorkerPool。
// 它立即启动所有工作协程，这些协程将处理提交到池中的任务。
func NewWorkerPool(poolSize, maxTaskLen int) *WorkerPool {
	pool := &WorkerPool{
		maxWorkers: int32(poolSize),
		workers:    make([]chan func(), poolSize),
		done:       make(chan struct{}),
	}
	pool.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		pool.workers[i] = make(chan func(), maxTaskLen)
		go pool.worker(i)
	}
	return pool
}

// worker is the goroutine that continuously processes tasks from its dedicated channel.
// It recovers from panics in task functions to prevent worker death.
//
// worker 是持续从其专属通道处理任务的 goroutine。
// 它从任务函数的 panic 中恢复，防止 worker 死亡。
func (p *WorkerPool) worker(idx int) {
	atomic.AddInt32(&p.activeWorkers, 1)
	defer func() {
		if r := recover(); r != nil {
			// Worker recovered from fatal panic / Worker 从致命 panic 中恢复
		}
		atomic.AddInt32(&p.activeWorkers, -1)
		p.wg.Done()
	}()

	for task := range p.workers[idx] {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Task panicked, continue processing next task / 任务 panic，继续处理下一个
				}
			}()
			task()
			atomic.AddInt64(&p.totalCompleted, 1)
		}()
	}
}

// Submit adds a task to the pool for execution using round-robin distribution.
// It blocks if the target worker's channel is full.
// Returns ErrPoolClosed if the pool has been closed.
// Priority: p.done > channel send
//
// Submit 使用轮询分配将任务添加到池中执行。
// 如果目标 worker 的通道已满则阻塞。
// 如果池已关闭则返回 ErrPoolClosed。
// 优先级：p.done > channel send
func (p *WorkerPool) Submit(task func()) error {
	// Non-blocking check closed first / 先非阻塞检查关闭状态
	select {
	case <-p.done:
		return ErrPoolClosed
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[p.nextWorker()] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return nil
	case <-p.done:
		return ErrPoolClosed
	}
}

// SubmitWithWorker adds a task bound to a specific worker.
// Tasks with the same workerId are guaranteed to execute on the same worker in FIFO order.
// The workerId will be mapped to a valid worker index internally using modulo operation.
// Priority: p.done > channel send
//
// Use HashWorkerId() to convert string (e.g., connection ID) to workerId.
//
// SubmitWithWorker 添加绑定到特定 worker 的任务。
// 相同 workerId 的任务保证在同一 worker 上按 FIFO 顺序执行。
// workerId 会在内部通过取模运算映射到有效的 worker 索引。
// 优先级：p.done > channel send
//
// 使用 HashWorkerId() 将字符串（如连接 ID）转换为 workerId。
func (p *WorkerPool) SubmitWithWorker(task func(), workerId int) error {
	idx := workerId % len(p.workers)
	if idx < 0 {
		idx = -idx
	}
	// Non-blocking check closed first / 先非阻塞检查关闭状态
	select {
	case <-p.done:
		return ErrPoolClosed
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[idx] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return nil
	case <-p.done:
		return ErrPoolClosed
	}
}

// SubmitCtx adds a task to the pool with context cancellation support using round-robin distribution.
// It blocks if the target worker's channel is full, but can be canceled via the context.
// Priority: ctx.Done() > p.done > channel send
//
// SubmitCtx 使用轮询分配将任务添加到池中，支持 context 取消。
// 如果目标 worker 的通道已满则阻塞，但可以通过 context 取消。
// 优先级：ctx.Done() > p.done > channel send
func (p *WorkerPool) SubmitCtx(ctx context.Context, task func()) error {
	// Non-blocking check cancellation signals first / 先非阻塞检查取消信号
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return ErrPoolClosed
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[p.nextWorker()] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return ErrPoolClosed
	}
}

// SubmitWithWorkerCtx adds a task bound to a specific worker with context cancellation support.
// Priority: ctx.Done() > p.done > channel send
//
// SubmitWithWorkerCtx 添加绑定到特定 worker 的任务，支持 context 取消。
// 优先级：ctx.Done() > p.done > channel send
func (p *WorkerPool) SubmitWithWorkerCtx(ctx context.Context, task func(), workerId int) error {
	idx := workerId % len(p.workers)
	if idx < 0 {
		idx = -idx
	}
	// Non-blocking check cancellation signals first / 先非阻塞检查取消信号
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return ErrPoolClosed
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[idx] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return ErrPoolClosed
	}
}

// TrySubmit attempts to add a task without blocking using round-robin distribution.
// Returns true if the task was submitted successfully, false if the channel is full or the pool is closed.
// Priority: p.done > channel send > default
//
// TrySubmit 尝试使用轮询分配非阻塞地添加任务。
// 如果任务提交成功则返回 true，如果通道已满或池已关闭则返回 false。
// 优先级：p.done > channel send > default
func (p *WorkerPool) TrySubmit(task func()) bool {
	// Non-blocking check closed first / 先非阻塞检查关闭状态
	select {
	case <-p.done:
		return false
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[p.nextWorker()] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return true
	case <-p.done:
		return false
	default:
		return false
	}
}

// TrySubmitWithWorker attempts to add a task bound to a specific worker without blocking.
// The workerId will be mapped to a valid worker index internally using modulo operation.
// Priority: p.done > channel send > default
//
// TrySubmitWithWorker 尝试非阻塞地添加绑定到特定 worker 的任务。
// workerId 会在内部通过取模运算映射到有效的 worker 索引。
// 优先级：p.done > channel send > default
func (p *WorkerPool) TrySubmitWithWorker(task func(), workerId int) bool {
	idx := workerId % len(p.workers)
	if idx < 0 {
		idx = -idx
	}
	// Non-blocking check closed first / 先非阻塞检查关闭状态
	select {
	case <-p.done:
		return false
	default:
	}
	// Then try to send task / 再尝试发送任务
	select {
	case p.workers[idx] <- task:
		atomic.AddInt64(&p.totalSubmitted, 1)
		return true
	case <-p.done:
		return false
	default:
		return false
	}
}

// Stop gracefully shuts down the worker pool. It prevents new task submissions,
// waits for all workers to finish their current tasks, and then returns.
// Subsequent calls to Stop are no-ops.
//
// Stop 优雅地关闭工作池。它阻止新任务提交，等待所有工作协程完成当前任务，然后返回。
// 后续对 Stop 的调用是无操作。
func (p *WorkerPool) Stop() {
	select {
	case <-p.done:
		return // Already stopped / 已经停止
	default:
		close(p.done)
	}
	// Close all worker channels after signaling done
	// 在发送 done 信号后关闭所有 worker 通道
	for _, ch := range p.workers {
		close(ch)
	}
	p.wg.Wait()
}

// Pending returns the total number of tasks waiting in all worker queues.
//
// Pending 返回所有 worker 队列中等待的任务总数。
func (p *WorkerPool) Pending() int {
	total := 0
	for _, ch := range p.workers {
		total += len(ch)
	}
	return total
}

// Stats returns a snapshot of the pool's current statistics.
//
// Stats 返回池当前统计信息的快照。
func (p *WorkerPool) Stats() WorkerPoolStats {
	return WorkerPoolStats{
		ActiveWorkers:  atomic.LoadInt32(&p.activeWorkers),
		PendingTasks:   p.Pending(),
		Capacity:       p.maxWorkers,
		TotalSubmitted: atomic.LoadInt64(&p.totalSubmitted),
		TotalCompleted: atomic.LoadInt64(&p.totalCompleted),
	}
}

// nextWorker returns the next worker index using round-robin distribution.
// Uses atomic counter for thread-safe concurrent access.
//
// nextWorker 使用轮询分配返回下一个 worker 索引。
// 使用原子计数器保证线程安全的并发访问。
func (p *WorkerPool) nextWorker() int {
	idx := atomic.AddUint64(&p.nextIdx, 1)
	return int(idx % uint64(len(p.workers)))
}

// HashWorkerId converts a string identifier to an integer.
// Use this to generate workerId from connection ID or other string keys.
// The returned value can be used with SubmitWithWorker, which will map it to a valid worker internally.
//
// HashWorkerId 使用 FNV 哈希将字符串标识符转换为整数。
// 用于从连接 ID 或其他字符串键生成 workerId。
// 返回值可用于 SubmitWithWorker，后者会在内部将其映射到有效的 worker。
func (p *WorkerPool) HashWorkerId(key string) int {
	if n, err := strconv.ParseInt(key, 16, 0); n > 0 && err == nil {
		return int(n)
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32())
}
