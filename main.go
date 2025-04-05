package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"aku-web/netease"
)

var (
	port = flag.String("port", "80", "port to listen on")
	dir  = flag.String("dir", "static", "directory to serve")

	playerMux       sync.Mutex
	cmd             *exec.Cmd // 保存当前播放的命令
	isWifiConnected bool      // 添加 WiFi 连接状态标志
	isApRunning     bool      // 添加 AP 运行状态标志
)

// AP 热点配置
const (
	AP_SSID      = "HlameMastar"
	AP_PASSWORD  = "12345678"
	AP_INTERFACE = "wlan0"
	AP_IP        = "192.168.4.1"
)

// AP 配网请求结构体
type apConfigRequest struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

// 检查 WiFi 连接状态
func checkWifiConnection() bool {
	// 获取网卡 IP 地址
	out, err := exec.Command("ip", "addr", "show", AP_INTERFACE).Output()
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

// 更新 WiFi 状态的 goroutine
func updateWifiStatus() {
	for {
		isWifiConnected = checkWifiConnection()
		// 如果无法连接到 WiFi 且 AP 未运行，创建 AP 热点
		if !isWifiConnected && !isApRunning {
			log.Println("Wifi not connected, creating AP hotspot...")
			if err := createAP(); err != nil {
				log.Printf("Failed to create AP hotspot: %v", err)
			} else {
				isApRunning = true // 设置 AP 运行标志
			}
		} else if isWifiConnected && isApRunning {
			// 如果 WiFi 已连接且 AP 正在运行，停止 AP
			log.Println("Wifi connected, stopping AP hotspot...")
			stopAP()
			isApRunning = false
		} else {
			log.Println("Heartbeat: Wifi_status", isWifiConnected, "AP_status", isApRunning)
		}
		time.Sleep(5 * time.Second)
	}
}

// 创建 AP 热点
func createAP() error {
	// 1. 停止网络服务
	exec.Command("/etc/init.d/S50wpa_supplicant", "stop").Run()
	exec.Command("killall", "hostapd").Run()
	exec.Command("killall", "avahi-daemon").Run()

	time.Sleep(2 * time.Second)

	// 2. 配置网络接口
	exec.Command("ip", "link", "set", AP_INTERFACE, "down").Run()
	time.Sleep(time.Second)

	exec.Command("ip", "addr", "flush", "dev", AP_INTERFACE).Run()
	exec.Command("sysctl", "-w", fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1", AP_INTERFACE)).Run()
	exec.Command("ip", "addr", "add", AP_IP+"/24", "dev", AP_INTERFACE).Run()
	exec.Command("ip", "link", "set", AP_INTERFACE, "up").Run()

	// 停止 DHCP 客户端
	exec.Command("killall", "udhcpc").Run()

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
fragm_threshold=2346`, AP_INTERFACE, AP_SSID, AP_PASSWORD)

	if err := os.WriteFile("/etc/hostapd.conf", []byte(hostapdConf), 0644); err != nil {
		return fmt.Errorf("failed to write hostapd config: %v", err)
	}

	// 4. 启动 hostapd
	cmd := exec.Command("hostapd", "-B", "/etc/hostapd.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start hostapd: %v", err)
	}

	log.Printf("AP hotspot created - SSID: %s", AP_SSID)
	return nil
}

// 停止 AP 热点
func stopAP() {
	// 停止 hostapd
	exec.Command("killall", "hostapd").Run()
	// 清理网络配置
	exec.Command("ip", "addr", "flush", "dev", AP_INTERFACE).Run()
	log.Println("AP hotspot stopped")
}

type HtmlFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

func handleGetHtmlFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取 static 目录下的所有文件
	files, err := os.ReadDir(*dir)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	// 过滤出 HTML 文件
	var htmlFiles []HtmlFile
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".html") {
			// 为特定文件提供默认描述
			description := "HTML页面"
			switch file.Name() {
			case "music_url.html":
				description = "支持网易云音乐歌单和流媒体播放"
			case "ap_config.html":
				description = "配置设备的 WiFi 连接"
			case "music_user.html":
				description = "本地音乐播放"
			case "index.html":
				continue // 跳过 index.html
			}

			htmlFiles = append(htmlFiles, HtmlFile{
				Name:        strings.TrimSuffix(file.Name(), ".html"),
				Path:        "/" + file.Name(),
				Description: description,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(htmlFiles)
}

func main() {
	flag.Parse()

	// 确保静态文件目录存在
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", *dir)
	}

	//  周期性检查 WiFi 状态
	go updateWifiStatus()

	// 创建自定义的 FileServer
	fs := http.FileServer(http.Dir(*dir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		// 如果未连接 WiFi 且不是访问配网页面或其 API，重定向到配网页面
		if !isWifiConnected &&
			r.URL.Path != "/ap_config.html" &&
			r.URL.Path != "/api/ap/config" &&
			!strings.HasPrefix(r.URL.Path, "/css/") &&
			!strings.HasPrefix(r.URL.Path, "/js/") {
			http.Redirect(w, r, "/ap_config.html", http.StatusTemporaryRedirect)
			log.Println("Redirecting to AP config page")
			return
		}

		if r.URL.Path == "/favicon.ico" {
			http.ServeFile(w, r, filepath.Join(*dir, "icon/favicon3.ico"))
			return
		}

		if r.URL.Path == "/api/music/list" {
			handleMusicList(w, r)
			return
		}

		if r.URL.Path == "/api/music/play" {
			handlePlayMusic(w, r)
			return
		}

		if r.URL.Path == "/api/music/stop" {
			handleStopMusic(w, r)
			return
		}

		if r.URL.Path == "/api/stream/play" {
			handleStreamPlay(w, r)
			return
		}

		if r.URL.Path == "/api/stream/stop" {
			handleStreamStop(w, r)
			return
		}

		// 音量控制路由
		if r.URL.Path == "/api/volume/get" {
			handleVolumeGet(w, r)
			return
		}

		if r.URL.Path == "/api/volume/set" {
			handleVolumeSet(w, r)
			return
		}

		// 添加 AP 配网路由
		if r.URL.Path == "/api/ap/config" {
			handleApConfig(w, r)
			return
		}

		if r.URL.Path == "/api/playlist/play" {
			handlePlaylistPlay(w, r)
			return
		}

		if r.URL.Path == "/api/playlist/detail" {
			handlePlaylistDetail(w, r)
			return
		}

		if r.URL.Path == "/api/html/list" {
			handleGetHtmlFiles(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on :%s", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleMusicList(w http.ResponseWriter, r *http.Request) {
	musicDir := filepath.Join(*dir, "music")
	files, err := os.ReadDir(musicDir)
	if err != nil {
		http.Error(w, "Failed to read music directory", http.StatusInternalServerError)
		return
	}

	var musicList []string
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".mp3" || ext == ".wav" {
				musicList = append(musicList, file.Name())
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(musicList)
}

func handlePlayMusic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playerMux.Lock()
	defer playerMux.Unlock()

	// 停止当前播放
	stopCurrentPlayback()

	// 构建音乐文件路径
	musicPath := filepath.Join(*dir, "music", request.Filename)

	// 检查文件是否存在
	if _, err := os.Stat(musicPath); os.IsNotExist(err) {
		http.Error(w, "Music file not found", http.StatusNotFound)
		return
	}

	// 根据文件扩展名选择播放器
	ext := strings.ToLower(filepath.Ext(request.Filename))
	switch ext {
	case ".mp3":
		cmd = exec.Command("mpg123", musicPath)
	case ".wav":
		cmd = exec.Command("aplay", musicPath)
	default:
		http.Error(w, "Unsupported audio format", http.StatusBadRequest)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start music playback: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Music playback error: %v", err)
		}
	}()
}

func handleStopMusic(w http.ResponseWriter, r *http.Request) {
	// 停止当前播放
	stopCurrentPlayback()
	w.WriteHeader(http.StatusOK)
}

func handleStreamPlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Stream playback started")

	var request struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playerMux.Lock()
	defer playerMux.Unlock()

	// 停止当前播放
	stopCurrentPlayback()

	// 使用 curl 和 mpg123 播放流媒体
	cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -k -L '%s' | mpg123 -", request.URL))
	fmt.Println("URL: " + request.URL)
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start stream playback: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Stream playback error: %v", err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}

func handleStreamStop(w http.ResponseWriter, r *http.Request) {
	// 停止当前播放
	stopCurrentPlayback()
	w.WriteHeader(http.StatusOK)
}

// 终止整个进程组
func stopCurrentPlayback() {
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		exec.Command("killall", "mpg123").Run()
		if err != nil {
			log.Printf("终止进程组失败: %v", err)
		} else {
			log.Println("播放已停止")
		}
		cmd = nil
	} else {
		log.Println("无进程需要终止")
	}
}

// 在文件末尾添加 AP 配网处理函数
func handleApConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request apConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if request.SSID == "" {
		http.Error(w, "SSID cannot be empty", http.StatusBadRequest)
		return
	}
	// 打印 WiFi 配置信息
	log.Printf("收到 WiFi 配置请求 - SSID: %s, Password: %s", request.SSID, request.Password)

	// 生成 wpa_supplicant.conf 的内容
	configContent := fmt.Sprintf(`ctrl_interface=/var/log/wpa_supplicant
			update_config=1

			network={
				ssid="%s"
				psk="%s"
			}`, request.SSID, request.Password)

	// 写入配置文件
	err := os.WriteFile("/etc/wpa_supplicant.conf", []byte(configContent), 0600)
	if err != nil {
		log.Printf("写入配置文件失败: %v", err)
		http.Error(w, "Failed to write configuration", http.StatusInternalServerError)
		return
	}

	log.Printf("WiFi 配置已更新 - SSID: %s", request.SSID)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "WiFi configuration updated and restarting",
	})

	// 停止 AP
	stopAP()
	// 重启 WiFi 网络
	cmd := exec.Command("/etc/init.d/S50wpa_supplicant", "restart")
	if err := cmd.Run(); err != nil {
		log.Printf("重启 WiFi 失败: %v", err)
		http.Error(w, "Failed to restart WiFi", http.StatusInternalServerError)
		return
	}
	log.Println("WiFi 服务已重启")

	// 等待 WiFi 连接（最多等待30秒）
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		// 检查 WiFi 连接状态
		if checkWifiConnection() {
			log.Printf("WiFi 连接成功 - SSID: %s (尝试次数: %d)", request.SSID, i+1)
			isWifiConnected = true
			isApRunning = false
			return
		}

		// 每次检查之间等待1秒
		if i < maxAttempts-1 { // 如果不是最后一次尝试
			time.Sleep(time.Second)
			if (i+1)%5 == 0 { // 每5秒记录一次日志
				log.Printf("等待 WiFi 连接中... (%d/%d)", i+1, maxAttempts)
			}
		}
	}

	log.Printf("WiFi 连接超时 - SSID: %s", request.SSID)
	isWifiConnected = false
	isApRunning = false
}

// 音量控制相关函数

// 获取当前音量
func handleVolumeGet(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("tinymix", "get", "0")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("获取音量失败: %v", err)
		http.Error(w, "Failed to get volume", http.StatusInternalServerError)
		return
	}

	// 返回音量信息
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"volume": strings.TrimSpace(string(output)),
	})
}

// 设置音量
func handleVolumeSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Volume string `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 设置音量
	cmd := exec.Command("tinymix", "set", "0", request.Volume)
	if err := cmd.Run(); err != nil {
		log.Printf("设置音量失败: %v", err)
		http.Error(w, "Failed to set volume", http.StatusInternalServerError)
		return
	}

	// 返回设置的音量值
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"volume": request.Volume,
	})
}

// 播放歌单
func handlePlaylistPlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PlaylistId string `json:"playlist_id"`
		Random     bool   `json:"random"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 获取歌单信息
	playlist, err := netease.GetPlaylist(request.PlaylistId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get playlist: %v", err), http.StatusInternalServerError)
		return
	}

	if len(playlist.Songs) == 0 {
		http.Error(w, "Playlist is empty", http.StatusInternalServerError)
		return
	}

	// 获取第一首歌的播放地址
	url := playlist.Songs[0].Url
	if url == "" {
		http.Error(w, "Failed to get music URL", http.StatusInternalServerError)
		return
	}

	playerMux.Lock()
	defer playerMux.Unlock()

	// 停止当前播放
	stopCurrentPlayback()

	// 使用 curl 和 mpg123 播放流媒体
	cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -k -L '%s' | mpg123 -", url))
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start playlist playback: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Playlist playback error: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "success",
		"message":      "Started playing playlist",
		"total_tracks": len(playlist.Songs),
	})
}

// 获取歌单详情
func handlePlaylistDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playlistId := r.URL.Query().Get("id")
	if playlistId == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	playlist, err := netease.GetPlaylist(playlistId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get playlist: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"songs":  playlist.Songs,
	})
}
