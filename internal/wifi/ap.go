package wifi

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"aku-web/internal/config"
)

// CreateAP 创建 AP 热点
func CreateAP() error {
	// 1. 停止网络服务
	exec.Command("/etc/init.d/S50wpa_supplicant", "stop").Run()
	exec.Command("killall", "hostapd").Run()

	time.Sleep(2 * time.Second)

	// 2. 配置网络接口
	exec.Command("ip", "link", "set", config.AP_INTERFACE, "down").Run()
	time.Sleep(time.Second)

	exec.Command("ip", "addr", "flush", "dev", config.AP_INTERFACE).Run()
	exec.Command("ip", "addr", "add", config.AP_IP+"/24", "dev", config.AP_INTERFACE).Run()
	exec.Command("ip", "link", "set", config.AP_INTERFACE, "up").Run()

	// 3. 创建 hostapd 配置
	hostapdConf := fmt.Sprintf(`interface=%s
driver=nl80211
ssid=%s
hw_mode=g
channel=1
auth_algs=1
wpa=2
wpa_passphrase=%s
wpa_key_mgmt=WPA-PSK
wpa_pairwise=CCMP
rsn_pairwise=CCMP
beacon_int=100
dtim_period=2
max_num_sta=5
rts_threshold=2347
fragm_threshold=2346`, config.AP_INTERFACE, config.AP_SSID, config.AP_PASSWORD)

	if err := os.WriteFile(config.HostAPDConfigPath, []byte(hostapdConf), 0644); err != nil {
		return fmt.Errorf("failed to write hostapd config: %v", err)
	}

	// 4. 启动 hostapd
	cmd := exec.Command("hostapd", "-B", config.HostAPDConfigPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start hostapd: %v", err)
	}

	log.Printf("AP hotspot created - SSID: %s", config.AP_SSID)
	return nil
}

// StopAP 停止 AP 热点
func StopAP() {
	// 停止 hostapd
	exec.Command("killall", "hostapd").Run()
	log.Println("AP hotspot stopped")
}
