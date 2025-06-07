package goroutinepool

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Task 代表一个需要执行的任务
type Task struct {
	ID       string
	Function func() error
	Callback func(error)
	Priority int // 0=低, 1=中, 2=高
	Timeout  time.Duration
	Retry    int
}

// Worker 工作协程
type Worker struct {
	ID         int
	TaskChan   chan *Task
	WorkerPool chan chan *Task
	Quit       chan bool
	ctx        context.Context
}

// Pool goroutine池
type Pool struct {
	WorkerPool chan chan *Task
	TaskQueue  chan *Task
	Workers    []*Worker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// 统计信息
	totalTasks     int64
	completedTasks int64
	failedTasks    int64
	activeTasks    int64
}

var (
	globalPool *Pool
	poolOnce   sync.Once
)

// GetPool 获取全局goroutine池
func GetPool() *Pool {
	poolOnce.Do(func() {
		globalPool = NewPool(runtime.NumCPU()*2, 10000)
		globalPool.Start()
	})
	return globalPool
}

// NewPool 创建新的goroutine池
func NewPool(maxWorkers int, maxQueue int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		WorkerPool: make(chan chan *Task, maxWorkers),
		TaskQueue:  make(chan *Task, maxQueue),
		Workers:    make([]*Worker, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 创建工作协程
	for i := 0; i < maxWorkers; i++ {
		worker := &Worker{
			ID:         i + 1,
			TaskChan:   make(chan *Task),
			WorkerPool: pool.WorkerPool,
			Quit:       make(chan bool),
			ctx:        ctx,
		}
		pool.Workers[i] = worker
	}

	return pool
}

// Start 启动goroutine池
func (p *Pool) Start() {
	// 启动分发器
	p.wg.Add(1)
	go p.dispatcher()

	// 启动所有工作协程
	for _, worker := range p.Workers {
		p.wg.Add(1)
		go worker.start(&p.wg)
	}

	// 启动统计收集器
	p.wg.Add(1)
	go p.statsCollector()

	log.Printf("Goroutine池已启动，工作协程数: %d", len(p.Workers))
}

// Stop 停止goroutine池
func (p *Pool) Stop() {
	log.Printf("正在停止goroutine池...")

	// 取消上下文
	p.cancel()

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	// 等待最多30秒
	select {
	case <-done:
		log.Printf("Goroutine池已安全停止")
	case <-time.After(30 * time.Second):
		log.Printf("Goroutine池停止超时，强制退出")
	}
}

// Submit 提交任务到池
func (p *Pool) Submit(task *Task) error {
	// 设置默认值
	if task.Timeout == 0 {
		task.Timeout = 30 * time.Second
	}
	if task.Retry == 0 {
		task.Retry = 3
	}

	atomic.AddInt64(&p.totalTasks, 1)

	select {
	case p.TaskQueue <- task:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		atomic.AddInt64(&p.failedTasks, 1)
		return ErrPoolOverloaded
	}
}

// SubmitFunc 提交简单函数
func (p *Pool) SubmitFunc(fn func() error) error {
	return p.Submit(&Task{
		Function: fn,
		Priority: 1,
	})
}

// SubmitWithCallback 提交带回调的任务
func (p *Pool) SubmitWithCallback(fn func() error, callback func(error)) error {
	return p.Submit(&Task{
		Function: fn,
		Callback: callback,
		Priority: 1,
	})
}

// dispatcher 任务分发器
func (p *Pool) dispatcher() {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.TaskQueue:
			// 获取可用工作协程
			select {
			case workerTaskChan := <-p.WorkerPool:
				// 分发任务给工作协程
				workerTaskChan <- task
			case <-p.ctx.Done():
				return
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// start 启动工作协程
func (w *Worker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// 将当前工作协程注册到池中
		select {
		case w.WorkerPool <- w.TaskChan:
			// 等待任务
			select {
			case task := <-w.TaskChan:
				w.executeTask(task)
			case <-w.ctx.Done():
				return
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
func (w *Worker) executeTask(task *Task) {
	atomic.AddInt64(&GetPool().activeTasks, 1)
	defer atomic.AddInt64(&GetPool().activeTasks, -1)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(w.ctx, task.Timeout)
	defer cancel()

	var err error
	done := make(chan error, 1)

	// 在单独的goroutine中执行任务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- NewTaskPanicError(r)
			}
		}()
		done <- task.Function()
	}()

	// 等待任务完成或超时
	select {
	case err = <-done:
		// 任务正常完成或出错
	case <-ctx.Done():
		err = ctx.Err()
	}

	// 重试逻辑
	if err != nil && task.Retry > 0 {
		task.Retry--
		time.Sleep(time.Second) // 等待1秒后重试
		w.executeTask(task)
		return
	}

	// 更新统计信息
	if err != nil {
		atomic.AddInt64(&GetPool().failedTasks, 1)
	} else {
		atomic.AddInt64(&GetPool().completedTasks, 1)
	}

	// 执行回调
	if task.Callback != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("任务回调发生panic: %v", r)
				}
			}()
			task.Callback(err)
		}()
	}
}

// statsCollector 统计信息收集器
func (p *Pool) statsCollector() {
	defer p.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			total := atomic.LoadInt64(&p.totalTasks)
			completed := atomic.LoadInt64(&p.completedTasks)
			failed := atomic.LoadInt64(&p.failedTasks)
			active := atomic.LoadInt64(&p.activeTasks)

			log.Printf("Goroutine池统计: 总任务=%d, 已完成=%d, 失败=%d, 活跃=%d",
				total, completed, failed, active)

		case <-p.ctx.Done():
			return
		}
	}
}

// GetStats 获取统计信息
func (p *Pool) GetStats() map[string]int64 {
	return map[string]int64{
		"total_tasks":     atomic.LoadInt64(&p.totalTasks),
		"completed_tasks": atomic.LoadInt64(&p.completedTasks),
		"failed_tasks":    atomic.LoadInt64(&p.failedTasks),
		"active_tasks":    atomic.LoadInt64(&p.activeTasks),
		"worker_count":    int64(len(p.Workers)),
	}
}

// 错误定义
var (
	ErrPoolOverloaded = NewPoolError("goroutine pool is overloaded")
)

type PoolError struct {
	Message string
}

func (e *PoolError) Error() string {
	return e.Message
}

func NewPoolError(message string) *PoolError {
	return &PoolError{Message: message}
}

type TaskPanicError struct {
	Panic interface{}
}

func (e *TaskPanicError) Error() string {
	return fmt.Sprintf("task panic: %v", e.Panic)
}

func NewTaskPanicError(panic interface{}) *TaskPanicError {
	return &TaskPanicError{Panic: panic}
}

// 便捷函数
func Submit(fn func() error) error {
	return GetPool().SubmitFunc(fn)
}

func SubmitWithCallback(fn func() error, callback func(error)) error {
	return GetPool().SubmitWithCallback(fn, callback)
}

func Stop() {
	GetPool().Stop()
}
