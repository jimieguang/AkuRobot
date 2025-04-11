package netease

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// API 端点
const (
	playlistDetailAPI = "https://music.163.com/api/v6/playlist/detail"
	songDetailAPI     = "https://music.163.com/api/v3/song/detail"
	batchSize         = 50 // 每批处理的歌曲数量
)

// 创建跳过证书验证的 HTTP 客户端
var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// PlaylistResponse 表示歌单详情的响应结构
type PlaylistResponse struct {
	Code     int `json:"code"`
	Playlist struct {
		Id         int64  `json:"id"`
		Name       string `json:"name"`
		TrackCount int    `json:"trackCount"`
		TrackIds   []struct {
			Id uint `json:"id"`
		} `json:"trackIds"`
	} `json:"playlist"`
}

// SongResponse 表示歌曲详情的响应结构
type SongResponse struct {
	Songs []struct {
		Id   uint   `json:"id"`
		Name string `json:"name"`
		Fee  int    `json:"fee"`
		Ar   []struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"ar"`
	} `json:"songs"`
}

// Song 表示歌曲的基本信息
type Song struct {
	Id      uint     `json:"id"`
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
	Url     string   `json:"url"`
	Fee     int      `json:"fee"` // 1 表示 VIP 歌曲
}

// Playlist 表示歌单及其歌曲
type Playlist struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Songs       []Song `json:"songs"`
}

// GetPlaylist 根据歌单ID获取歌单信息和歌曲
func GetPlaylist(playlistId string, page, pageSize int) (*Playlist, error) {
	// 1. 获取歌单基本信息
	playlistInfo, err := getPlaylistInfo(playlistId)
	if err != nil {
		return nil, fmt.Errorf("获取歌单信息失败: %w", err)
	}

	if playlistInfo.Code == 401 {
		return nil, errors.New("无权限访问此歌单")
	}

	// 2. 计算分页
	totalSongs := len(playlistInfo.Playlist.TrackIds)
	start := (page - 1) * pageSize
	if start >= totalSongs {
		return nil, fmt.Errorf("页码超出范围")
	}
	end := start + pageSize
	if end > totalSongs {
		end = totalSongs
	}

	// 3. 获取当前页的歌曲详情
	var allSongs []Song
	songIds := make([]uint, end-start)
	for i, track := range playlistInfo.Playlist.TrackIds[start:end] {
		songIds[i] = track.Id
	}

	// 4. 分批处理歌曲
	for i := 0; i < len(songIds); i += batchSize {
		batchEnd := i + batchSize
		if batchEnd > len(songIds) {
			batchEnd = len(songIds)
		}

		songs, err := getSongsDetail(songIds[i:batchEnd])
		if err != nil {
			log.Printf("警告: 获取歌曲 %d-%d 详情失败: %v", i, batchEnd, err)
			continue
		}
		allSongs = append(allSongs, songs...)
	}

	if len(allSongs) == 0 {
		return nil, fmt.Errorf("未能获取任何歌曲详情")
	}

	// 5. 控制并发获取音乐URL
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // 限制并发数为5
	urlChan := make(chan struct {
		index int
		url   string
	}, len(allSongs))

	for i, song := range allSongs {
		wg.Add(1)
		go func(i int, id uint) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			url := getMusicUrl(fmt.Sprintf("%d", id))
			if url != "" {
				urlChan <- struct {
					index int
					url   string
				}{i, url}
			}
		}(i, song.Id)
	}

	// 所有goroutine完成后关闭channel
	go func() {
		wg.Wait()
		close(urlChan)
	}()

	// 收集URL
	for result := range urlChan {
		allSongs[result.index].Url = result.url
	}

	// 6. 创建响应
	playlist := &Playlist{
		Id:          playlistInfo.Playlist.Id,
		Name:        playlistInfo.Playlist.Name,
		Description: "",
		Songs:       allSongs,
	}

	return playlist, nil
}

// getPlaylistInfo 获取歌单基本信息
func getPlaylistInfo(playlistId string) (*PlaylistResponse, error) {
	// 创建请求
	data := strings.NewReader("id=" + playlistId)
	req, err := http.NewRequest("POST", playlistDetailAPI, data)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	playlistResp := &PlaylistResponse{}
	if err = json.Unmarshal(body, playlistResp); err != nil {
		return nil, err
	}

	return playlistResp, nil
}

// getSongsDetail 获取多首歌曲的详细信息
func getSongsDetail(songIds []uint) ([]Song, error) {
	// 创建歌曲ID对象
	songIdObjs := make([]map[string]uint, len(songIds))
	for i, id := range songIds {
		songIdObjs[i] = map[string]uint{"id": id}
	}

	// 序列化歌曲ID
	jsonData, err := json.Marshal(songIdObjs)
	if err != nil {
		return nil, err
	}

	// 创建请求
	data := strings.NewReader("c=" + string(jsonData))
	req, err := http.NewRequest("POST", songDetailAPI, data)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	songResp := &SongResponse{}
	if err = json.Unmarshal(body, songResp); err != nil {
		return nil, err
	}

	// 转换为Song对象
	songs := make([]Song, len(songResp.Songs))
	for i, song := range songResp.Songs {
		artists := make([]string, len(song.Ar))
		for j, ar := range song.Ar {
			artists[j] = ar.Name
		}

		songs[i] = Song{
			Id:      song.Id,
			Name:    song.Name,
			Artists: artists,
			Fee:     song.Fee,
		}
	}

	return songs, nil
}

// getMusicUrl 获取音乐的直接URL
func getMusicUrl(id string) string {
	resp, err := insecureClient.Get("https://music.163.com/song/media/outer/url?id=" + id)
	if err != nil {
		log.Printf("检查歌曲是否可用出错: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.Request.URL.Path == "/404" {
		return ""
	}

	return resp.Request.URL.String()
}

// GetSongUrl 获取单个歌曲的URL
func GetSongUrl(id uint) (string, error) {
	url := getMusicUrl(fmt.Sprintf("%d", id))
	if url == "" {
		return "", fmt.Errorf("无法获取歌曲播放地址")
	}
	return url, nil
}
