package player

import (
	"fmt"
	"log"
	"os/exec"
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
