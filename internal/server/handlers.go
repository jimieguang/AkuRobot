package server

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
)

// 处理系统重启请求
func handleSystemReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "System is rebooting...",
	})

	// 异步执行重启命令
	go func() {
		time.Sleep(1 * time.Second) // 等待响应发送完成
		exec.Command("reboot").Run()
	}()
}
