package scheduler

import "sync"

func depthInc(queueKey string, m *sync.Map) int {
	depth, _ := m.LoadOrStore(queueKey, 0)
	newDepth := depth.(int) + 1
	m.Store(queueKey, newDepth)
	return newDepth
}

func buildTaskTree(tasks []*Task) *TaskNode {
	taskMap := make(map[string]*TaskNode)
	var root *TaskNode

	// 构建任务节点
	for _, t := range tasks {
		taskMap[t.ID] = &TaskNode{Task: t}
	}

	// 建立父子关系
	for _, t := range tasks {
		if t.ParentTaskID == "" {
			root = taskMap[t.ID]
		} else if parent, ok := taskMap[t.ParentTaskID]; ok {
			parent.Children = append(parent.Children, taskMap[t.ID])
		}
	}
	return root
}

func taskToJSON(node *TaskNode) map[string]interface{} {
	if node == nil {
		return make(map[string]interface{})
	}
	result := map[string]interface{}{
		"id":       node.Task.ID,
		"queue":    node.Task.QueueKey,
		"finished": node.Task.IsFinished,
		"active":   node.Task.IsActive,
		"children": make([]map[string]interface{}, 0),
	}
	for _, child := range node.Children {
		result["children"] = append(result["children"].([]map[string]interface{}), taskToJSON(child))
	}
	return result
}
