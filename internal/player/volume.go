package player

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"aku-web/internal/config"
)

// GetVolume 获取当前音量
func GetVolume() (int, error) {
	cmd := exec.Command("tinymix", "get", "0")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get volume: %v", err)
	}

	volumeStr := strings.TrimSpace(string(output))
	volume, err := strconv.Atoi(volumeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse volume: %v", err)
	}

	return volume, nil
}

// SetVolume 设置音量
func SetVolume(volume int) error {
	// 验证音量范围
	if volume < 0 || volume > config.MaxVolume {
		return fmt.Errorf("volume must be between 0 and %d", config.MaxVolume)
	}

	// 设置音量
	cmd := exec.Command("tinymix", "set", "0", strconv.Itoa(volume))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set volume: %v", err)
	}

	return nil
}
