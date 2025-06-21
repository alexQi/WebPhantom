// Package scheduler 提供基于优先级队列和动态弹性伸缩的任务调度功能
package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"noctua/internal/queue"
	"sync"
	"sync/atomic"
	"time"
)

// SchedulerStatus 定义调度器的状态
type SchedulerStatus struct {
	Running        bool                   `json:"running"`        // 是否正在运行
	Paused         bool                   `json:"paused"`         // 是否暂停
	Config         Config                 `json:"config"`         // 当前配置
	QueueDetails   map[string]QueueStatus `json:"queueDetails"`   // 各队列详情
	TotalWorkers   int                    `json:"totalWorkers"`   // 总 worker 数
	ActiveWorkers  int                    `json:"activeWorkers"`  // 活跃 worker 数
	QueueDepth     int                    `json:"queueDepth"`     // 总队列深度
	ProcessedTasks int                    `json:"processedTasks"` // 已处理任务数
	FailedTasks    int                    `json:"failedTasks"`    // 失败任务数
	PendingTasks   int                    `json:"pendingTasks"`   // taskIndex 中的任务数
}

// QueueStatus 定义队列状态
type QueueStatus struct {
	Depth         int `json:"depth"`         // 当前队列深度
	Workers       int `json:"workers"`       // worker 总数
	ActiveWorkers int `json:"activeWorkers"` // 活跃 worker 数
	QPS           int `json:"qps"`           // 当前速率限制
}

type Config struct {
	MaxWorkersPerQueue int
	WorkerIdleTimeout  time.Duration
	AutoScaleInterval  time.Duration
	MaxQueueDepth      int
	BaseRetryDelay     time.Duration
	DefaultQPS         int
}

type Metrics struct {
	QueueDepths    *sync.Map
	ActiveWorkers  *sync.Map
	ProcessedTasks *sync.Map
	FailedTasks    *sync.Map
}

func (m *Metrics) Reset() {
	m.QueueDepths.Clear()
	m.ActiveWorkers.Clear()
	m.ProcessedTasks.Clear()
	m.FailedTasks.Clear()
}

type Scheduler struct {
	queues       *sync.Map // map[string]*queues.PriorityQueue
	queueLocks   *sync.Map // map[string]*sync.RWMutex
	workers      *sync.Map // map[string][]*worker
	handlerFuncs *sync.Map // map[string]TaskHandler
	taskIndex    *sync.Map // 新增：每个队列的任务 ID 索引
	queueQPS     *sync.Map // 新增：map[string]int，存储每个队列的 QPS
	metrics      *Metrics
	config       Config
	mainCtx      context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	dispatcherWg sync.WaitGroup
	wg           sync.WaitGroup
	mu           sync.Mutex
	isPaused     atomic.Bool // 新增：暂停状态
}

type TaskHandler func(*Task) error

func New(mainCtx context.Context, cfg Config) *Scheduler {
	if cfg.MaxWorkersPerQueue == 0 {
		cfg.MaxWorkersPerQueue = 20
	}
	if cfg.WorkerIdleTimeout == 0 {
		cfg.WorkerIdleTimeout = 30
	}
	if cfg.AutoScaleInterval == 0 {
		cfg.AutoScaleInterval = 1
	}
	if cfg.MaxQueueDepth == 0 {
		cfg.MaxQueueDepth = 1000
	}
	if cfg.BaseRetryDelay == 0 {
		cfg.BaseRetryDelay = 1
	}
	if cfg.DefaultQPS == 0 {
		cfg.DefaultQPS = 1
	}
	ctx, cancel := context.WithCancel(mainCtx)
	s := &Scheduler{
		mainCtx:      mainCtx,
		ctx:          ctx,
		cancel:       cancel,
		config:       cfg,
		isPaused:     atomic.Bool{}, // 初始化为 false
		queues:       new(sync.Map),
		queueLocks:   new(sync.Map),
		workers:      new(sync.Map),
		handlerFuncs: new(sync.Map),
		taskIndex:    new(sync.Map),
		queueQPS:     new(sync.Map),
		metrics: &Metrics{
			QueueDepths:    new(sync.Map),
			ActiveWorkers:  new(sync.Map),
			ProcessedTasks: new(sync.Map),
			FailedTasks:    new(sync.Map),
		},
	}
	// 停止调度器，同意状态
	s.cancel()
	return s
}

func (s *Scheduler) Context() context.Context {
	return s.ctx
}

// Pause 暂停调度器
func (s *Scheduler) Pause() {
	s.isPaused.Store(true)
}

// Resume 恢复调度器
func (s *Scheduler) Resume() {
	s.isPaused.Store(false)
}

func (s *Scheduler) recursionDeleteTask(task *Task) {
	if len(task.Children) > 0 {
		for _, subTask := range task.Children {
			s.recursionDeleteTask(subTask)
		}
	}
	s.taskIndex.Delete(task.ID)
}

func (s *Scheduler) processTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			s.mu.Lock()
			s.recordFailed(task)
			s.mu.Unlock()
		}
	}()

	handlerVal, _ := s.handlerFuncs.Load(task.QueueKey)
	if handlerVal == nil {
		s.mu.Lock()
		s.recordFailed(task)
		s.mu.Unlock()
		return
	}
	handler := handlerVal.(TaskHandler)

	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		err := handler(task)
		errCh <- err
	}()

	var err error
	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = ctx.Err()
	}

	if err == nil {
		task.IsFinished = true
		s.taskIndex.Store(task.ID, task)
		s.recordSuccess(task)
		return
	}

	if task.CurrentRetry < task.MaxRetries {
		task.CurrentRetry++
		go s.retryTask(task)
	} else {
		s.recordFailed(task)
	}
}

func (s *Scheduler) recordSuccess(task *Task) {
	s.updateMetric(s.metrics.ProcessedTasks, task.QueueKey, 1)
}

func (s *Scheduler) recordFailed(task *Task) {
	s.updateMetric(s.metrics.FailedTasks, task.QueueKey, 1)
}

func (s *Scheduler) allSubTasksInactive(task *Task) bool {
	if len(task.Children) == 0 {
		return true // 无子任务，直接认为不活跃
	}
	for _, subTask := range task.Children {
		if subTask.IsActive {
			return false
		}
	}
	return true
}

func (s *Scheduler) updateParentTask(task *Task) {
	if parentTask, err := s.getTaskByID(task.ParentTaskID); err == nil {
		allFinished := true
		for _, child := range parentTask.Children {
			if !child.IsFinished {
				allFinished = false
				break
			}
		}
		if allFinished {
			parentTask.IsFinished = true
			s.taskIndex.Store(parentTask.ID, parentTask)
		}
	}
}

// initQueue 初始化队列（线程安全）
func (s *Scheduler) initQueue(queueKey string) {
	if _, loaded := s.queues.LoadOrStore(queueKey, queue.NewPriorityQueue[*TaskItem]()); !loaded {
		s.queueLocks.Store(queueKey, new(sync.RWMutex))
		s.metrics.QueueDepths.Store(queueKey, 0)
		heap.Init(s.getQueue(queueKey))

		// 启动队列的专属分发goroutine
		s.dispatcherWg.Add(1)
		go s.dispatchLoop(queueKey)
	}
}

func (s *Scheduler) getQueue(queueKey string) *queue.PriorityQueue[*TaskItem] {
	q, _ := s.queues.Load(queueKey)
	return q.(*queue.PriorityQueue[*TaskItem])
}

func (s *Scheduler) getQueueLock(queueKey string) *sync.RWMutex {
	l, _ := s.queueLocks.Load(queueKey)
	return l.(*sync.RWMutex)
}

func (s *Scheduler) RegisterHandler(funcsMapKey string, handler TaskHandler) {
	s.handlerFuncs.Store(funcsMapKey, handler)
}

// 分发循环负责将任务分配给worker
func (s *Scheduler) dispatchLoop(queueKey string) {
	defer s.dispatcherWg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if !s.isPaused.Load() {
				s.dispatchTasks(queueKey)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) dispatchTasks(queueKey string) {
	q := s.getQueue(queueKey)
	if q.Len() == 0 {
		return
	}
	workers, _ := s.workers.Load(queueKey)
	if workers == nil {
		return
	}

	lock := s.getQueueLock(queueKey)
	lock.Lock()
	defer lock.Unlock()

	availableWorkers := workers.([]*worker)
	for _, w := range availableWorkers {
		if q.Len() == 0 {
			break
		}
		if atomic.CompareAndSwapInt32(&w.active, 0, 1) {
			item := heap.Pop(q).(*TaskItem)
			s.updateQueueDepth(queueKey, -1)
			w.taskChan <- item.Task
		}
	}
}

// 提交任务并返回任务ID
func (s *Scheduler) SubmitTask(task Task) (string, error) {
	if s.ctx.Err() != nil {
		return "", fmt.Errorf("Scheduler has been stopped...")
	}
	s.initQueue(task.QueueKey)
	if depth, _ := s.metrics.QueueDepths.Load(task.QueueKey); depth.(int) >= s.config.MaxQueueDepth {
		return "", fmt.Errorf("queue %s is full", task.QueueKey)
	}
	item := &TaskItem{
		Task:       &task,
		EnqueuedAt: time.Now(),
	}
	lock := s.getQueueLock(task.QueueKey)
	lock.Lock()
	heap.Push(s.getQueue(task.QueueKey), item)
	depthCount := depthInc(task.QueueKey, s.metrics.QueueDepths)
	s.metrics.QueueDepths.Store(task.QueueKey, depthCount)
	lock.Unlock()

	s.taskIndex.Store(task.ID, &task)
	if task.ParentTaskID != "" {
		s.markParentHasSubTask(task.ParentTaskID, &task)
	}
	return task.ID, nil
}

func (s *Scheduler) markParentHasSubTask(parentTaskID string, subTask *Task) {
	parentTask, ok := s.taskIndex.Load(parentTaskID)
	if !ok {
		return
	}
	parent := parentTask.(*Task)
	parent.Children = append(parent.Children, subTask)
	parent.IsActive = true
	s.taskIndex.Store(parentTaskID, parent)
}

func (s *Scheduler) getTaskByID(taskID string) (*Task, error) {
	task, ok := s.taskIndex.Load(taskID)
	if !ok {
		return nil, fmt.Errorf("Not found in TaskIndex: %s", taskID)
	}
	return task.(*Task), nil
}

func (s *Scheduler) autoScaler() {
	ticker := time.NewTicker(s.config.AutoScaleInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !s.isPaused.Load() {
				s.queues.Range(func(key, _ interface{}) bool {
					queueKey := key.(string)
					s.adjustWorkers(queueKey)
					return true
				})
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) adjustWorkers(queueKey string) {
	currentWorkers := s.workerCount(queueKey)
	idealWorkers := s.calculateIdealWorkers(queueKey)

	if idealWorkers > currentWorkers {
		s.scaleUp(queueKey, idealWorkers-currentWorkers)
	} else if idealWorkers < currentWorkers {
		s.scaleDown(queueKey, currentWorkers-idealWorkers)
	}
}

func (s *Scheduler) workerCount(queueKey string) int {
	workers, _ := s.workers.Load(queueKey)
	if workers == nil {
		return 0
	}
	return len(workers.([]*worker))
}

func (s *Scheduler) calculateIdealWorkers(queueKey string) int {
	depthVal, _ := s.metrics.QueueDepths.Load(queueKey)
	depth := depthVal.(int)
	if depth == 0 {
		return 0
	}
	return int(math.Ceil(math.Sqrt(float64(depth))))
}

// SetQueueQPS 设置特定队列的 QPS
func (s *Scheduler) SetQueueQPS(queueKey string, qps int) {
	if qps <= 0 {
		qps = s.config.DefaultQPS // 防止无效值
	}
	s.queueQPS.Store(queueKey, qps)
}

// GetQueueQPS 获取队列的 QPS，默认为 DefaultQPS
func (s *Scheduler) GetQueueQPS(queueKey string) int {
	if qps, ok := s.queueQPS.Load(queueKey); ok {
		return qps.(int)
	}
	return s.config.DefaultQPS
}

func (s *Scheduler) scaleUp(queueKey string, count int) {
	qps := s.GetQueueQPS(queueKey) // 使用队列特定的 QPS
	for i := 0; i < count; i++ {
		w := newWorker(queueKey, s.config.WorkerIdleTimeout*time.Second, qps)
		actual, _ := s.workers.LoadOrStore(queueKey, []*worker{})
		workers := append(actual.([]*worker), w)
		s.workers.Store(queueKey, workers)
		s.wg.Add(1)
		go s.runWorker(w)
	}
}

func (s *Scheduler) scaleDown(queueKey string, count int) {
	actual, _ := s.workers.Load(queueKey)
	if actual == nil {
		return
	}

	workers := actual.([]*worker)
	var retain []*worker
	toRemove := count
	for _, w := range workers {
		if toRemove > 0 && !w.IsActive() && time.Since(w.lastActive) > w.idleTimeout {
			w.stop()
			toRemove--
		} else {
			retain = append(retain, w)
		}
	}
	s.workers.Store(queueKey, retain)
}

func (s *Scheduler) runWorker(w *worker) {
	defer s.wg.Done()
	for {
		select {
		case task := <-w.taskChan:
			// 使用 worker 的限流器
			if !s.isPaused.Load() {
				err := w.limiter.Wait(s.ctx)
				if err != nil {
					return
				}
				atomic.StoreInt32(&w.active, 1)
				s.processTask(task)
				atomic.StoreInt32(&w.active, 0)
				w.lastActive = time.Now()
			}
		case <-w.quitChan:
			return
			// 检查 worker 是否空闲超过 idleTimeout
		case <-time.After(w.idleTimeout):
			// 如果 worker 已经空闲超过了指定时间
			if time.Since(w.lastActive) > w.idleTimeout && !s.isPaused.Load() {
				// 关闭该 worker
				s.scaleDown(w.queue, 1) // 调用 scaleDown 停止该 worker
				return
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) updateMetric(m *sync.Map, key string, delta int) {
	val, _ := m.LoadOrStore(key, 0)
	newVal := val.(int) + delta
	m.Store(key, newVal)
}

func (s *Scheduler) updateQueueDepth(queueKey string, delta int) {
	val, _ := s.metrics.QueueDepths.LoadOrStore(queueKey, 0)
	newVal := val.(int) + delta
	s.metrics.QueueDepths.Store(queueKey, newVal)
}

func (s *Scheduler) retryTask(task *Task) {
	delay := s.config.BaseRetryDelay * time.Second * time.Duration(math.Pow(2, float64(task.CurrentRetry)))
	time.Sleep(delay)
	if _, err := s.SubmitTask(*task); err != nil {
		s.recordFailed(task)
	}
}

func (s *Scheduler) taskStateChecker() {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !s.isPaused.Load() {
				s.checkTaskStates()
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) checkTaskStates() {
	s.taskIndex.Range(func(key, value interface{}) bool {
		task := value.(*Task)
		if task.IsFinished && s.allSubTasksInactive(task) {
			task.IsActive = false
			s.taskIndex.Store(task.ID, task)
			if len(task.ParentTaskID) == 0 {
				s.recursionDeleteTask(task)
			} else {
				s.updateParentTask(task)
			}
		}
		return true
	})
}

func (s *Scheduler) IsPaused() bool {
	return s.isPaused.Load()
}

func (s *Scheduler) Status() *SchedulerStatus {
	queueDepth, processed, failed, _ := s.GetTaskStatistics()

	totalWorkers := 0
	activeWorkers := 0
	queueDetails := make(map[string]QueueStatus)
	s.queues.Range(func(key, value interface{}) bool {
		queueKey := key.(string)
		depth, _ := s.metrics.QueueDepths.LoadOrStore(queueKey, 0)

		workers, _ := s.workers.Load(queueKey)
		workerCount := 0
		active := 0
		if workers != nil {
			ws := workers.([]*worker)
			workerCount = len(ws)
			for _, w := range ws {
				if w.IsActive() {
					active++
				}
			}
		}
		totalWorkers += workerCount
		activeWorkers += active

		queueDetails[queueKey] = QueueStatus{
			Depth:         depth.(int),
			Workers:       workerCount,
			ActiveWorkers: active,
			QPS:           s.GetQueueQPS(queueKey),
		}
		return true
	})

	pendingTasks := 0
	s.taskIndex.Range(func(_, _ interface{}) bool {
		pendingTasks++
		return true
	})

	return &SchedulerStatus{
		Running:        s.ctx.Err() == nil,
		Paused:         s.isPaused.Load(),
		Config:         s.config,
		QueueDetails:   queueDetails,
		TotalWorkers:   totalWorkers,
		ActiveWorkers:  activeWorkers,
		QueueDepth:     queueDepth,
		ProcessedTasks: processed,
		FailedTasks:    failed,
		PendingTasks:   pendingTasks,
	}
}

func (s *Scheduler) GetTaskTree() map[string]interface{} {
	tasks := make([]*Task, 0)
	s.taskIndex.Range(func(_, value interface{}) bool {
		tasks = append(tasks, value.(*Task))
		return true
	})
	root := buildTaskTree(tasks)
	return taskToJSON(root)
}

func (s *Scheduler) GetTaskStatistics() (int, int, int, error) {
	// 初始化统计变量
	var totalQueueDepth int
	var totalProcessedTasks int
	var totalFailedTasks int

	// 获取队列深度
	s.metrics.QueueDepths.Range(func(_, v interface{}) bool {
		totalQueueDepth += v.(int)
		return true
	})

	// 获取已处理任务的数量
	s.metrics.ProcessedTasks.Range(func(_, v interface{}) bool {
		totalProcessedTasks += v.(int)
		return true
	})

	// 获取失败任务的数量
	s.metrics.FailedTasks.Range(func(_, v interface{}) bool {
		totalFailedTasks += v.(int)
		return true
	})

	// 返回统计信息
	return totalQueueDepth, totalProcessedTasks, totalFailedTasks, nil
}

// Reset 重置调度器状态，准备下一次运行
func (s *Scheduler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 取消现有上下文并创建新的
	if s.ctx.Err() == nil {
		s.cancel()
	}
	s.wg.Wait() // 确保所有旧 goroutine 结束
	s.ctx, s.cancel = context.WithCancel(s.mainCtx)

	// 清空现有数据
	s.queues.Clear()
	s.queueLocks.Clear()
	s.workers.Clear()
	s.taskIndex.Clear()
	// 重置metrics
	s.metrics.Reset()

	go s.autoScaler()
	go s.taskStateChecker()
}

// Shutdown 停止调度器
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 取消上下文，触发所有 goroutine 退出
	if s.ctx.Err() == nil {
		s.cancel()
	}

	// 停止所有 worker
	s.workers.Range(func(key, value interface{}) bool {
		workers := value.([]*worker)
		for _, w := range workers {
			w.stop()
		}
		return true
	})
	// 等待所有 goroutine 结束
	s.wg.Wait()
	s.dispatcherWg.Wait()

	// 清空所有 sync.Map
	s.queues.Clear()
	s.queueLocks.Clear()
	s.workers.Clear()
	s.taskIndex.Clear()
	// 重置metrics
	s.metrics.Reset()
}

func (s *Scheduler) WaitUntilEmpty() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		totalQueueDepth := 0
		s.metrics.QueueDepths.Range(func(_, v interface{}) bool {
			totalQueueDepth += v.(int)
			return true
		})
		// 检查 taskIndex 是否为空
		totalTaskIndex := 0
		s.taskIndex.Range(func(_, v interface{}) bool {
			totalTaskIndex++
			return true
		})
		// 如果队列深度为 0 且 taskIndex 也为空，表示所有任务都已经完成
		if totalQueueDepth == 0 && totalTaskIndex == 0 {
			return
		}
		select {
		case <-ticker.C:
		case <-s.ctx.Done():
			return
		}
	}
}
