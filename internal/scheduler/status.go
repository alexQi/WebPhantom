package scheduler

type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "Pending"    // 任务待处理
	TaskStatusProgressing TaskStatus = "Processing" // 任务处理中
	TaskStatusProgressed  TaskStatus = "Processed"  // 任务已处理
	TaskStatusFailed      TaskStatus = "Failed"     // 任务失败
	TaskStatusWaitingSub  TaskStatus = "WaitingSub" // 主任务已完成，等待子任务
)
