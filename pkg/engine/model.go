package engine

import (
	"encoding/json"
)

type DeviceInfo struct {
	CookieEnabled   bool    `json:"cookie_enabled"`
	ScreenWidth     int     `json:"screen_width"`
	ScreenHeight    int     `json:"screen_height"`
	BrowserLanguage string  `json:"browser_language"`
	BrowserPlatform string  `json:"browser_platform"`
	BrowserName     string  `json:"browser_name"`
	BrowserVersion  string  `json:"browser_version"`
	BrowserOnline   bool    `json:"browser_online"`
	EngineName      string  `json:"engine_name"`
	EngineVersion   string  `json:"engine_version"`
	OsName          string  `json:"os_name"`
	OsVersion       string  `json:"os_version"`
	CpuCoreNum      int     `json:"cpu_core_num"`
	DeviceMemory    float64 `json:"device_memory"`
	Downlink        float64 `json:"downlink"`
	EffectiveType   string  `json:"effective_type"`
	RoundTripTime   int     `json:"round_trip_time"`
}

func (di *DeviceInfo) String() string {
	marshal, err := json.Marshal(di)
	if err != nil {
		return ""
	}
	return string(marshal)
}
