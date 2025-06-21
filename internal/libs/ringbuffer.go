package libs

// RingBuffer 是一个泛型循环缓冲区
type RingBuffer[T any] struct {
	data   []T
	size   int
	cursor int
	count  int
}

// NewRingBuffer 创建一个新的固定长度的循环数组
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		data:   make([]T, size),
		size:   size,
		cursor: 0,
		count:  0,
	}
}

// Add 向循环数组添加新数据
func (rb *RingBuffer[T]) Add(value T) {
	if rb.count < rb.size {
		rb.count++
	}
	rb.data[rb.cursor] = value
	rb.cursor = (rb.cursor + 1) % rb.size
}

// Count 返回当前元素数量
func (rb *RingBuffer[T]) Count() int {
	return rb.count
}

// GetAll 返回数组中的所有元素，顺序是从最旧的元素到最新的元素
func (rb *RingBuffer[T]) GetAll() []T {
	if rb.count == rb.size {
		return append(rb.data[rb.cursor:], rb.data[:rb.cursor]...)
	}
	return rb.data[:rb.count]
}

// First 获取当前数组的开头元素（最旧的元素）
func (rb *RingBuffer[T]) First() (T, bool) {
	var zero T
	if rb.count == 0 {
		return zero, false // 表示数组为空
	}
	index := (rb.cursor - rb.count + rb.size) % rb.size
	return rb.data[index], true
}

// Last 获取当前索引往前指定 n 的值（n=0 为最新元素）
func (rb *RingBuffer[T]) Last(n int) (T, bool) {
	var zero T
	if n < 0 || n >= rb.count {
		return zero, false // 表示索引无效
	}
	index := (rb.cursor - 1 - n + rb.size) % rb.size
	return rb.data[index], true
}

// Clear 清空循环数组中的所有元素
func (rb *RingBuffer[T]) Clear() {
	rb.data = make([]T, rb.size)
	rb.cursor = 0
	rb.count = 0
}
