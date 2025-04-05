package server

import (
	"net/http"
	"path/filepath"

	"aku-web/internal/api"
)

// RegisterRoutes 注册所有 HTTP 路由
func RegisterRoutes() {
	// Favicon 处理
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "icon", "favicon.ico"))
	})

	// HTML 文件列表路由
	http.HandleFunc("/api/html/list", api.HandleGetHtmlFiles)

	// 音乐播放相关路由
	http.HandleFunc("/api/music/list", api.HandleMusicList)
	http.HandleFunc("/api/music/play", api.HandlePlayMusic)
	http.HandleFunc("/api/music/stream", api.HandleStreamPlay)
	http.HandleFunc("/api/music/stop", api.HandleStreamStop)
	http.HandleFunc("/api/music/info", api.HandleGetAudioInfo)

	// 音量控制路由
	http.HandleFunc("/api/volume", api.HandleVolumeGet)
	http.HandleFunc("/api/volume/set", api.HandleVolumeSet)

	// WiFi 配置路由
	http.HandleFunc("/api/wifi/config", api.HandleApConfig)

	// 歌单相关路由
	http.HandleFunc("/api/playlist/detail", api.HandlePlaylistDetail)
	http.HandleFunc("/api/playlist/play", api.HandlePlaylistPlay)

	// 添加重启设备的路由
	http.HandleFunc("/api/system/reboot", handleSystemReboot)

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
}
