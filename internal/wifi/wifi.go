package wifi

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"aku-web/internal/config"
)

var (
	IsWifiConnected bool
	IsApRunning     bool
)

// CheckConnection 检查 WiFi 连接状态
func CheckConnection() bool {
	// 获取网卡 IP 地址
	out, err := exec.Command("ip", "addr", "show", config.AP_INTERFACE).Output()
	if err != nil {
		log.Printf("获取网卡信息失败: %v", err)
		return false
	}

	ipInfo := string(out)

	// 使用正则表达式匹配 IP 地址
	re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(ipInfo, -1)

	for _, match := range matches {
		ip := match[1]
		// 排除以下IP:
		// - 192.168.4.x (AP地址)
		// - 169.254.x.x (APIPA地址)
		// - 127.0.0.1 (本地回环)
		if !strings.HasPrefix(ip, "192.168.4.") &&
			!strings.HasPrefix(ip, "169.254.") &&
			ip != "127.0.0.1" {
			log.Printf("Found valid IP: %s", ip)
			return true
		}
	}
	return false
}

// StartMonitoring 启动 WiFi 状态监控
func StartMonitoring() {
	for {
		IsWifiConnected = CheckConnection()
		// 如果无法连接到 WiFi 且 AP 未运行，创建 AP 热点
		if !IsWifiConnected && !IsApRunning {
			log.Println("Wifi not connected, creating AP hotspot...")
			if err := CreateAP(); err != nil {
				log.Printf("Failed to create AP hotspot: %v", err)
			} else {
				IsApRunning = true
			}
		} else if IsWifiConnected && IsApRunning {
			// 如果 WiFi 已连接且 AP 正在运行，停止 AP
			log.Println("Wifi connected, stopping AP hotspot...")
			StopAP()
			IsApRunning = false
		} else {
			log.Println("Heartbeat: Wifi_status", IsWifiConnected, "AP_status", IsApRunning)
		}
		time.Sleep(config.WifiCheckInterval)
	}
}

// ConfigureWifi 配置 WiFi 连接
func ConfigureWifi(ssid, password string) error {
	// 打印 WiFi 配置信息
	log.Printf("收到 WiFi 配置请求 - SSID: %s, Password: %s", ssid, password)

	// 生成 wpa_supplicant.conf 的内容
	configContent := fmt.Sprintf(`ctrl_interface=/var/log/wpa_supplicant
update_config=1

network={
    ssid="%s"
    psk="%s"
}`, ssid, password)

	// 写入配置文件
	if err := os.WriteFile(config.WPAConfigPath, []byte(configContent), 0600); err != nil {
		log.Printf("写入配置文件失败: %v", err)
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	log.Printf("WiFi 配置已更新 - SSID: %s", ssid)

	// 停止 AP
	StopAP()
	// 假装AP正在运行以防止监控程序干扰
	IsApRunning = true
	time.Sleep(time.Second * 3)

	// 重启 WiFi 网络
	cmd := exec.Command("/etc/init.d/S50wpa_supplicant", "restart")
	if err := cmd.Run(); err != nil {
		log.Printf("重启 WiFi 失败: %v", err)
		IsApRunning = false
		return fmt.Errorf("failed to restart WiFi: %v", err)
	}
	log.Println("WiFi 服务已重启")

	// 等待 WiFi 连接（最多等待30秒）
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		// 检查 WiFi 连接状态
		if CheckConnection() {
			log.Printf("WiFi 连接成功 - SSID: %s (尝试次数: %d)", ssid, i+1)
			IsWifiConnected = true
			IsApRunning = false
			return nil
		}

		// 每次检查之间等待1秒
		if i < maxAttempts-1 { // 如果不是最后一次尝试
			time.Sleep(time.Second)
			if (i+1)%5 == 0 { // 每5秒记录一次日志
				log.Printf("等待 WiFi 连接中... (%d/%d)", i+1, maxAttempts)
				log.Println("IsApRunning", IsApRunning)
			}
		}
	}

	log.Printf("WiFi 连接超时 - SSID: %s", ssid)
	IsWifiConnected = false
	IsApRunning = false
	return fmt.Errorf("WiFi connection timeout")
}
