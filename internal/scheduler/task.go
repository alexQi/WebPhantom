package scheduler

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	DefaultMaxRetries = 3
	DefaultTimeout    = 30 * time.Second
	DefaultPriority   = 8
	DefaultQPS        = 100
)

// Task 定义任务结构
type Task struct {
	ID           string
	ParentTaskID string  // 父级任务ID，如果是主任务则为""，子任务为主任务的ID
	SourceTaskID string  // 原始任务ID
	IsActive     bool    // 是否激活或还有子任务
	IsFinished   bool    // 当前任务是否完成
	Children     []*Task // 子任务列表
	QueueKey     string  // 队列名称
	Priority     int     // 0-9 (9为最高优先级)
	Payload      interface{}
	MaxRetries   int
	CurrentRetry int
	Status       TaskStatus
	Dependencies []string
	CreatedAt    time.Time
	Timeout      time.Duration
}

type TaskItem struct {
	Task       *Task
	EnqueuedAt time.Time
	Index      int
}

func (ti *TaskItem) GetID() string {
	return ti.Task.ID
}

// GetPriority 实现 PriorityItem 接口的 GetPriority 方法
func (ti *TaskItem) GetPriority() int {
	return ti.Task.Priority
}

// GetEnqueuedAt 实现 PriorityItem 接口的 GetEnqueuedAt 方法
func (ti *TaskItem) GetEnqueuedAt() time.Time {
	return ti.EnqueuedAt
}

// SetIndex 实现 PriorityItem 接口的 SetIndex 方法
func (ti *TaskItem) SetIndex(index int) {
	ti.Index = index
}

type TaskNode struct {
	Task     *Task
	Children []*TaskNode
}

// TaskOptions 用于定义任务的可选参数
type TaskOptions struct {
	ParentTaskID string        // 父级任务ID，如果是主任务则为""，子任务为主任务的ID
	SourceTaskID string        // 原始任务ID
	Priority     int           // 优先级
	Payload      interface{}   // 任务负载
	MaxRetries   int           // 最大重试次数
	Dependencies []string      // 任务依赖
	Timeout      time.Duration // 超时时间
}

// NewTask 创建一个新的任务，使用 TaskOptions 作为可选参数
func NewTask(queueKey string, payload interface{}, options TaskOptions) (Task, error) {
	// 设置默认值
	if options.MaxRetries == 0 {
		options.MaxRetries = DefaultMaxRetries // 默认最大重试次数为 3
	}
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout // 默认超时时间为 30 秒
	}
	if options.Priority == 0 {
		options.Priority = DefaultPriority // 默认优先级为 8
	}
	// 生成任务ID
	taskID := fmt.Sprintf("%s-%s", queueKey, uuid.New().String())
	var sourceTaskID string
	if options.ParentTaskID == "" && options.SourceTaskID == "" {
		sourceTaskID = taskID
	} else {
		sourceTaskID = options.SourceTaskID
	}

	return Task{
		ID:           taskID,
		QueueKey:     queueKey,
		Payload:      payload,
		ParentTaskID: options.ParentTaskID,
		SourceTaskID: sourceTaskID,
		Children:     make([]*Task, 0),
		IsFinished:   false,
		IsActive:     true,
		Priority:     options.Priority,
		MaxRetries:   options.MaxRetries,
		Status:       TaskStatusPending,
		CurrentRetry: 0,
		Dependencies: options.Dependencies,
		CreatedAt:    time.Now(),
		Timeout:      options.Timeout,
	}, nil
}
