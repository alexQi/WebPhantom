package mixed

import "time"

// GenerateDailyTimeSeries 生成按天的时间序列
func GenerateDailyTimeSeries(start, end time.Time) []string {
	var timeSeries []string
	current := start

	// 当当前时间不超过结束时间时，继续添加
	for !current.After(end) {
		timeSeries = append(timeSeries, current.Format("2006-01-02 15:04:05"))
		current = current.Add(24 * time.Hour) // 按天增加
	}
	return timeSeries
}
