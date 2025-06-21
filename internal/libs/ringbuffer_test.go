package libs

import (
	"reflect"
	"testing"
	"time"
)

// TestRingBuffer 测试 RingBuffer 的所有功能
func TestRingBuffer(t *testing.T) {
	// 测试用例
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"EmptyBuffer", testEmptyBuffer},
		{"AddAndCount", testAddAndCount},
		{"GetAll", testGetAll},
		{"FirstAndLast", testFirstAndLast},
		{"FullBufferAndOverflow", testFullBufferAndOverflow},
		{"Clear", testClear},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// testEmptyBuffer 测试空缓冲区
func testEmptyBuffer(t *testing.T) {
	rb := NewRingBuffer[time.Time](5)

	if got := rb.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
	if got := rb.GetAll(); len(got) != 0 {
		t.Errorf("GetAll() = %v, want empty slice", got)
	}
	if _, ok := rb.First(); ok {
		t.Errorf("First() should return false for empty buffer")
	}
	if _, ok := rb.Last(0); ok {
		t.Errorf("Last(0) should return false for empty buffer")
	}
}

// testAddAndCount 测试添加和计数
func testAddAndCount(t *testing.T) {
	rb := NewRingBuffer[time.Time](3)
	now := time.Now()

	rb.Add(now)
	if got := rb.Count(); got != 1 {
		t.Errorf("Count() = %d, want 1", got)
	}

	rb.Add(now.Add(1 * time.Minute))
	if got := rb.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
}

// testGetAll 测试获取所有元素
func testGetAll(t *testing.T) {
	rb := NewRingBuffer[time.Time](3)
	now := time.Now()
	times := []time.Time{
		now,
		now.Add(1 * time.Minute),
		now.Add(2 * time.Minute),
	}

	// 未满时
	rb.Add(times[0])
	rb.Add(times[1])
	if got := rb.GetAll(); !reflect.DeepEqual(got, times[:2]) {
		t.Errorf("GetAll() = %v, want %v", got, times[:2])
	}

	// 填满后
	rb.Add(times[2])
	if got := rb.GetAll(); !reflect.DeepEqual(got, times) {
		t.Errorf("GetAll() = %v, want %v", got, times)
	}

	// 覆盖后
	rb.Add(now.Add(3 * time.Minute))
	want := []time.Time{times[1], times[2], now.Add(3 * time.Minute)}
	if got := rb.GetAll(); !reflect.DeepEqual(got, want) {
		t.Errorf("GetAll() = %v, want %v", got, want)
	}
}

// testFirstAndLast 测试获取首尾元素
func testFirstAndLast(t *testing.T) {
	rb := NewRingBuffer[time.Time](3)
	now := time.Now()
	times := []time.Time{
		now,
		now.Add(1 * time.Minute),
		now.Add(2 * time.Minute),
	}

	// 添加两个元素
	rb.Add(times[0])
	rb.Add(times[1])

	if first, ok := rb.First(); !ok || !first.Equal(times[0]) {
		t.Errorf("First() = %v, %v, want %v, true", first, ok, times[0])
	}
	if last, ok := rb.Last(0); !ok || !last.Equal(times[1]) {
		t.Errorf("Last(0) = %v, %v, want %v, true", last, ok, times[1])
	}
	if last, ok := rb.Last(1); !ok || !last.Equal(times[0]) {
		t.Errorf("Last(1) = %v, %v, want %v, true", last, ok, times[0])
	}
	if _, ok := rb.Last(2); ok {
		t.Errorf("Last(2) should return false")
	}

	// 填满并覆盖
	rb.Add(times[2])
	rb.Add(now.Add(3 * time.Minute))
	if first, ok := rb.First(); !ok || !first.Equal(times[1]) {
		t.Errorf("First() = %v, %v, want %v, true", first, ok, times[1])
	}
	if last, ok := rb.Last(0); !ok || !last.Equal(now.Add(3*time.Minute)) {
		t.Errorf("Last(0) = %v, %v, want %v, true", last, ok, now.Add(3*time.Minute))
	}
}

// testFullBufferAndOverflow 测试缓冲区满后覆盖
func testFullBufferAndOverflow(t *testing.T) {
	rb := NewRingBuffer[time.Time](3)
	now := time.Now()
	times := []time.Time{
		now,
		now.Add(1 * time.Minute),
		now.Add(2 * time.Minute),
		now.Add(3 * time.Minute),
	}

	// 填满
	for i := 0; i < 3; i++ {
		rb.Add(times[i])
	}
	if got := rb.Count(); got != 3 {
		t.Errorf("Count() = %d, want 3", got)
	}

	// 覆盖
	rb.Add(times[3])
	if got := rb.Count(); got != 3 {
		t.Errorf("Count() = %d, want 3", got)
	}
	want := []time.Time{times[1], times[2], times[3]}
	if got := rb.GetAll(); !reflect.DeepEqual(got, want) {
		t.Errorf("GetAll() = %v, want %v", got, want)
	}
}

// testClear 测试清空缓冲区
func testClear(t *testing.T) {
	rb := NewRingBuffer[time.Time](3)
	now := time.Now()

	rb.Add(now)
	rb.Add(now.Add(1 * time.Minute))
	rb.Clear()

	if got := rb.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
	if got := rb.GetAll(); len(got) != 0 {
		t.Errorf("GetAll() = %v, want empty slice", got)
	}
	if _, ok := rb.First(); ok {
		t.Errorf("First() should return false after clear")
	}
}
