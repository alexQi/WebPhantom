package queue

import (
	"container/heap"
	"sync"
	"time"
)

// PriorityItem 是一个泛型接口，定义了元素需要满足的方法
type PriorityItem[T any] interface {
	GetPriority() int
	GetEnqueuedAt() time.Time
	GetID() string
	SetIndex(int)
}

// PriorityQueue 是一个泛型的优先队列
type PriorityQueue[T PriorityItem[T]] struct {
	items []T
	index map[string]int // 新增：ID 到索引的映射
	mu    sync.RWMutex
}

// NewPriorityQueue 创建并返回一个新的优先队列实例
func NewPriorityQueue[T PriorityItem[T]]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: make([]T, 0),
		index: make(map[string]int),
	}
}

// List 返回优先队列中的所有任务副本
func (pq *PriorityQueue[T]) List() []T {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	tasks := make([]T, len(pq.items))
	copy(tasks, pq.items)
	return tasks
}

// Len 返回优先队列的长度
func (pq *PriorityQueue[T]) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// Less 比较两个元素的优先级和入队时间，仅供 heap 包内部调用。
// 注意：调用者需确保在并发环境下已持有锁。
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	if pq.items[i].GetPriority() == pq.items[j].GetPriority() {
		return pq.items[i].GetEnqueuedAt().Before(pq.items[j].GetEnqueuedAt())
	}
	return pq.items[i].GetPriority() > pq.items[j].GetPriority()
}

// Swap 交换两个元素的位置，并更新它们在堆中的索引
func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].SetIndex(i)
	pq.items[j].SetIndex(j)
	pq.index[pq.items[i].GetID()] = i
	pq.index[pq.items[j].GetID()] = j
}

// Push 向优先队列中添加一个元素
func (pq *PriorityQueue[T]) Push(x interface{}) {
	item, ok := x.(T)
	if !ok {
		return
	}
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.index[item.GetID()] = len(pq.items)
	pq.items = append(pq.items, item)
}

// Rem 根据任务 ID 从优先队列中移除指定任务
func (pq *PriorityQueue[T]) Rem(taskID string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return false
	}

	index, exists := pq.index[taskID]
	if !exists {
		return false
	}

	heap.Remove(pq, index)
	delete(pq.index, taskID)
	// 更新受影响元素的索引
	for i := index; i < len(pq.items); i++ {
		pq.index[pq.items[i].GetID()] = i
	}
	return true
}

// Contains 检查指定的 ID 列表中哪些在队列中，返回每个 ID 的存在状态
func (pq *PriorityQueue[T]) Contains(ids []string) map[string]bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		_, exists := pq.index[id]
		result[id] = exists
	}
	return result
}

// Pop 从优先队列中移除并返回最后一个元素
func (pq *PriorityQueue[T]) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.items) == 0 {
		return nil
	}
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	delete(pq.index, item.GetID())
	return item
}

// Clear 清空优先队列中的所有元素
func (pq *PriorityQueue[T]) Clear() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items = make([]T, 0)
	pq.index = make(map[string]int)
}
