package player

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	playerMux sync.Mutex
	cmd       *exec.Cmd // 保存当前播放的命令
)

// PlayLocalFile 播放本地音乐文件
func PlayLocalFile(filePath string) error {
	playerMux.Lock()
	defer playerMux.Unlock()

	// 停止当前播放
	StopPlayback()

	// 根据文件扩展名选择播放器
	var err error
	switch {
	case strings.HasSuffix(strings.ToLower(filePath), ".mp3"):
		cmd = exec.Command("mpg123", filePath)
	case strings.HasSuffix(strings.ToLower(filePath), ".wav"):
		cmd = exec.Command("aplay", filePath)
	default:
		return fmt.Errorf("unsupported audio format")
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start playback: %v", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Playback error: %v", err)
		}
	}()

	return nil
}

// PlayStream 播放流媒体
func PlayStream(url string, position float64) error {
	playerMux.Lock()
	defer playerMux.Unlock()

	// 停止当前播放
	StopPlayback()

	// 构建播放命令，支持从指定位置开始播放
	skipParam := ""
	if position > 0 {
		// mpg123 的 -k 参数以帧为单位，1秒约等于38.28帧
		frames := int(position * 38.28)
		skipParam = fmt.Sprintf("-k %d", frames)
	}

	// 使用 curl 和 mpg123 播放流媒体
	cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -k -L '%s' | mpg123 %s -", url, skipParam))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start stream playback: %v", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Stream playback error: %v", err)
		}
	}()

	return nil
}

// StopPlayback 停止当前播放
func StopPlayback() {
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		exec.Command("killall", "mpg123").Run()
		if err != nil {
			log.Printf("Failed to kill process: %v", err)
		} else {
			log.Println("Playback stopped")
		}
		cmd = nil
	}
}

// GetAudioInfo 获取音频文件信息
func GetAudioInfo(url string) (*AudioInfo, error) {
	// 使用 curl 和 mpg123 获取音频信息
	cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -k -L -s '%s' | mpg123 --skip-printing -t -", url))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %v", err)
	}

	// 解析输出信息
	info := &AudioInfo{}
	outputStr := string(output)

	// 解析时长
	if match := regexp.MustCompile(`Length:\s+(\d+:\d+)`).FindStringSubmatch(outputStr); len(match) > 1 {
		timeStr := match[1]
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			minutes, _ := strconv.Atoi(parts[0])
			seconds, _ := strconv.Atoi(parts[1])
			info.Duration = float64(minutes*60 + seconds)
		}
	}

	// 解析标题
	if match := regexp.MustCompile(`Title:\s+(.+)`).FindStringSubmatch(outputStr); len(match) > 1 {
		info.Title = match[1]
	}

	// 解析艺术家
	if match := regexp.MustCompile(`Artist:\s+(.+)`).FindStringSubmatch(outputStr); len(match) > 1 {
		info.Artist = match[1]
	}

	// 解析比特率
	if match := regexp.MustCompile(`(\d+)\s+kbit/s`).FindStringSubmatch(outputStr); len(match) > 1 {
		info.Bitrate, _ = strconv.Atoi(match[1])
	}

	return info, nil
}

// AudioInfo 音频文件信息结构体
type AudioInfo struct {
	Duration float64 `json:"duration"` // 总时长（秒）
	Title    string  `json:"title"`    // 标题
	Artist   string  `json:"artist"`   // 艺术家
	Bitrate  int     `json:"bitrate"`  // 比特率
}
