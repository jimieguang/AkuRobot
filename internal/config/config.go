package config

import "time"

// Server 配置
const (
	DefaultPort = "80"
	DefaultDir  = "static"
)

// AP 热点配置
const (
	AP_SSID      = "HlameMastar"
	AP_PASSWORD  = "12345678"
	AP_INTERFACE = "wlan0"
	AP_IP        = "192.168.4.1"
)

// 音频相关配置
const (
	MaxVolume = 63
)

// 网络相关配置
const (
	WifiCheckInterval  = 5 * time.Second  // WiFi 状态检查间隔
	WifiConnectTimeout = 30 * time.Second // WiFi 连接超时时间
)

// 文件路径配置
const (
	WPAConfigPath     = "/etc/wpa_supplicant.conf"
	HostAPDConfigPath = "/etc/hostapd.conf"
)
