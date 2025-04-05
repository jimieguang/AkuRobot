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
)

// API endpoints
const (
	playlistDetailAPI = "https://music.163.com/api/v6/playlist/detail"
	songDetailAPI     = "https://music.163.com/api/v3/song/detail"
)

// 创建跳过证书验证的 HTTP 客户端
var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// PlaylistResponse represents the response structure for playlist details
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

// SongResponse represents the response structure for song details
type SongResponse struct {
	Songs []struct {
		Id   uint   `json:"id"`
		Name string `json:"name"`
		Ar   []struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"ar"`
	} `json:"songs"`
}

// Song represents a song with its basic information
type Song struct {
	Id      uint     `json:"id"`
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
	Url     string   `json:"url"`
}

// Playlist represents a playlist with its songs
type Playlist struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
	Songs      []Song `json:"songs"`
}

// GetPlaylist retrieves playlist information and its songs by playlist ID
func GetPlaylist(playlistId string) (*Playlist, error) {
	// 1. Get playlist basic information
	playlistInfo, err := getPlaylistInfo(playlistId)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist info: %w", err)
	}

	if playlistInfo.Code == 401 {
		return nil, errors.New("no permission to access this playlist")
	}

	// 2. Get song details
	songIds := make([]uint, len(playlistInfo.Playlist.TrackIds))
	for i, track := range playlistInfo.Playlist.TrackIds {
		songIds[i] = track.Id
	}

	songs, err := getSongsDetail(songIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get songs detail: %w", err)
	}

	// 3. Get music URLs for each song
	for i, song := range songs {
		url := getMusicUrl(fmt.Sprintf("%d", song.Id))
		songs[i].Url = url
	}

	// 4. Create response
	playlist := &Playlist{
		Id:         playlistInfo.Playlist.Id,
		Name:       playlistInfo.Playlist.Name,
		TrackCount: playlistInfo.Playlist.TrackCount,
		Songs:      songs,
	}

	return playlist, nil
}

// getPlaylistInfo retrieves basic playlist information
func getPlaylistInfo(playlistId string) (*PlaylistResponse, error) {
	// Create request
	data := strings.NewReader("id=" + playlistId)
	req, err := http.NewRequest("POST", playlistDetailAPI, data)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request using insecure client
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response
	playlistResp := &PlaylistResponse{}
	if err = json.Unmarshal(body, playlistResp); err != nil {
		return nil, err
	}

	return playlistResp, nil
}

// getSongsDetail retrieves detailed information for multiple songs
func getSongsDetail(songIds []uint) ([]Song, error) {
	// Create song ID objects for request
	songIdObjs := make([]map[string]uint, len(songIds))
	for i, id := range songIds {
		songIdObjs[i] = map[string]uint{"id": id}
	}

	// Marshal song IDs
	jsonData, err := json.Marshal(songIdObjs)
	if err != nil {
		return nil, err
	}

	// Create request
	data := strings.NewReader("c=" + string(jsonData))
	req, err := http.NewRequest("POST", songDetailAPI, data)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request using insecure client
	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response
	songResp := &SongResponse{}
	if err = json.Unmarshal(body, songResp); err != nil {
		return nil, err
	}

	// Convert to Song objects
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
		}
	}

	return songs, nil
}

// getMusicUrl retrieves the direct URL for a music track
func getMusicUrl(id string) string {
	resp, err := insecureClient.Get("https://music.163.com/song/media/outer/url?id=" + id)
	if err != nil {
		log.Println("检查歌曲是否可用出错", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.Request.URL.Path != "/404" {
		return resp.Request.URL.String()
	} else {
		log.Println("需要VIP", id)
		return ""
	}
}
