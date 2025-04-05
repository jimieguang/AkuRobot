package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"aku-web/internal/config"
	"aku-web/internal/netease"
	"aku-web/internal/player"
	"aku-web/internal/wifi"
)

// HtmlFile 表示 HTML 文件信息
type HtmlFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

// HandleMusicList 处理获取音乐列表的请求
func HandleMusicList(w http.ResponseWriter, r *http.Request) {
	musicDir := filepath.Join(config.DefaultDir, "music")
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

// HandlePlayMusic 处理播放本地音乐的请求
func HandlePlayMusic(w http.ResponseWriter, r *http.Request) {
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

	// 构建音乐文件路径
	musicPath := filepath.Join(config.DefaultDir, "music", request.Filename)

	// 检查文件是否存在
	if _, err := os.Stat(musicPath); os.IsNotExist(err) {
		http.Error(w, "Music file not found", http.StatusNotFound)
		return
	}

	if err := player.PlayLocalFile(musicPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to play music: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStreamPlay 处理播放流媒体的请求
func HandleStreamPlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		URL      string  `json:"url"`
		Position float64 `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := player.PlayStream(request.URL, request.Position); err != nil {
		http.Error(w, fmt.Sprintf("Failed to play stream: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStreamStop 处理停止播放的请求
func HandleStreamStop(w http.ResponseWriter, r *http.Request) {
	player.StopPlayback()
	w.WriteHeader(http.StatusOK)
}

// HandleVolumeGet 处理获取音量的请求
func HandleVolumeGet(w http.ResponseWriter, r *http.Request) {
	volume, err := player.GetVolume()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get volume: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"volume": volume})
}

// HandleVolumeSet 处理设置音量的请求
func HandleVolumeSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Volume interface{} `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var volume int
	switch v := request.Volume.(type) {
	case float64:
		volume = int(v)
	case string:
		var err error
		volume, err = strconv.Atoi(v)
		if err != nil {
			http.Error(w, "Invalid volume value", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Invalid volume type", http.StatusBadRequest)
		return
	}

	if err := player.SetVolume(volume); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set volume: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleApConfig 处理 WiFi 配置的请求
func HandleApConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		SSID     string `json:"ssid"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.SSID == "" {
		http.Error(w, "SSID cannot be empty", http.StatusBadRequest)
		return
	}

	log.Printf("收到 WiFi 配置请求 - SSID: %s", request.SSID)
	if err := wifi.ConfigureWifi(request.SSID, request.Password); err != nil {
		http.Error(w, fmt.Sprintf("Failed to configure WiFi: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "WiFi configuration updated",
	})
}

// HandlePlaylistPlay 处理播放歌单的请求
func HandlePlaylistPlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		SongId uint `json:"song_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 获取歌曲URL
	url, err := netease.GetSongUrl(request.SongId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get song URL: %v", err), http.StatusInternalServerError)
		return
	}
	if url == "" {
		http.Error(w, "Failed to get playable URL for the song", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"url":    url,
	})
}

// HandlePlaylistDetail 处理获取歌单详情的请求
func HandlePlaylistDetail(w http.ResponseWriter, r *http.Request) {
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

// HandleGetAudioInfo 处理获取音频信息的请求
func HandleGetAudioInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	info, err := player.GetAudioInfo(request.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get audio info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// HandleGetHtmlFiles 处理获取 HTML 文件列表的请求
func HandleGetHtmlFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取 static 目录下的所有文件
	files, err := os.ReadDir(config.DefaultDir)
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
