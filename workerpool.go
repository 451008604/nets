package nets

import (
	"context"
	"errors"
	"hash/fnv"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	ErrPoolClosed  = errors.New("worker pool is closed")
	ErrInvalidTask = errors.New("nil task")
)

type WorkerPoolStats struct {
	ActiveWorkers  int32
	PendingTasks   int
	Capacity       int32
	TotalSubmitted int64
}

// WorkerPool executes bounded queues. Stop rejects new work and drains accepted tasks.
type WorkerPool struct {
	maxWorkers     int32
	activeWorkers  int32
	workers        []chan func()
	space          []chan struct{}
	done           chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
	mu             sync.RWMutex
	capture        *MsgHandler
	nextIdx        uint64
	totalSubmitted int64
}

func GetInstanceWorkerPool() *WorkerPool { return getDefaultServerManager().WorkerPool() }
func NewWorkerPool(size, capacity uint) *WorkerPool {
	return newWorkerPool(size, capacity, NewMsgHandler())
}
func newWorkerPool(size, capacity uint, capture *MsgHandler) *WorkerPool {
	if size == 0 {
		size = uint(runtime.NumCPU() * 10)
	}
	if capacity == 0 {
		capacity = 10
	}
	p := &WorkerPool{maxWorkers: int32(size), workers: make([]chan func(), size), space: make([]chan struct{}, size), done: make(chan struct{}), capture: capture}
	p.wg.Add(int(size))
	for i := range p.workers {
		p.workers[i] = make(chan func(), capacity)
		p.space[i] = make(chan struct{}, 1)
		go p.worker(i)
	}
	return p
}
func (p *WorkerPool) runTask(task func()) {
	defer p.capture.GetErrCapture(nil)
	task()
}
func (p *WorkerPool) worker(idx int) {
	atomic.AddInt32(&p.activeWorkers, 1)
	defer func() { atomic.AddInt32(&p.activeWorkers, -1); p.wg.Done() }()
	for {
		select {
		case task := <-p.workers[idx]:
			select {
			case p.space[idx] <- struct{}{}:
			default:
			}
			p.runTask(task)
		case <-p.done:
			for {
				select {
				case task := <-p.workers[idx]:
					p.runTask(task)
				default:
					return
				}
			}
		}
	}
}
func (p *WorkerPool) submit(ctx context.Context, task func(), idx int, block bool) error {
	if task == nil {
		return ErrInvalidTask
	}
	for {
		p.mu.RLock()
		if err := ctx.Err(); err != nil {
			p.mu.RUnlock()
			return err
		}
		select {
		case <-p.done:
			p.mu.RUnlock()
			return ErrPoolClosed
		default:
		}
		select {
		case p.workers[idx] <- task:
			atomic.AddInt64(&p.totalSubmitted, 1)
			p.mu.RUnlock()
			return nil
		default:
			p.mu.RUnlock()
		}
		if !block {
			return ErrCapacity
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.done:
			return ErrPoolClosed
		case <-p.space[idx]:
		}
	}
}
func (p *WorkerPool) workerIndex(id int) int {
	idx := id % len(p.workers)
	if idx < 0 {
		idx = -idx
	}
	return idx
}
func (p *WorkerPool) Submit(task func()) error { return p.SubmitCtx(context.Background(), task) }
func (p *WorkerPool) SubmitWithWorker(task func(), id int) error {
	return p.SubmitWithWorkerCtx(context.Background(), task, id)
}
func (p *WorkerPool) SubmitCtx(ctx context.Context, task func()) error {
	return p.submit(ctx, task, p.nextWorker(), true)
}
func (p *WorkerPool) SubmitWithWorkerCtx(ctx context.Context, task func(), id int) error {
	return p.submit(ctx, task, p.workerIndex(id), true)
}
func (p *WorkerPool) TrySubmit(task func()) bool {
	return p.submit(context.Background(), task, p.nextWorker(), false) == nil
}
func (p *WorkerPool) TrySubmitWithWorker(task func(), id int) bool {
	return p.submit(context.Background(), task, p.workerIndex(id), false) == nil
}

// Stop waits for accepted work. Call it from outside the pool's own tasks.
func (p *WorkerPool) Stop() {
	p.stopOnce.Do(func() { p.mu.Lock(); close(p.done); p.mu.Unlock() })
	p.wg.Wait()
}
func (p *WorkerPool) Pending() int {
	total := 0
	for _, ch := range p.workers {
		total += len(ch)
	}
	return total
}
func (p *WorkerPool) Stats() WorkerPoolStats {
	return WorkerPoolStats{atomic.LoadInt32(&p.activeWorkers), p.Pending(), p.maxWorkers, atomic.LoadInt64(&p.totalSubmitted)}
}
func (p *WorkerPool) nextWorker() int {
	return int(atomic.AddUint64(&p.nextIdx, 1) % uint64(len(p.workers)))
}
func (p *WorkerPool) HashWorkerId(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}
