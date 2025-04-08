package main

import (
	"aku-web/internal/server"
	"io"
	"log"
	"os"
	"os/exec"
	// "aku-web/internal/wifi"
)

func main() {
	// 设置日志输出到控制台和文件
	logFile, err := os.OpenFile("aku-web.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("无法打开日志文件: %v，将只输出到控制台", err)
	} else {
		defer logFile.Close()
		// 同时输出到控制台和文件
		log.SetOutput(os.Stdout)
		if logFile != nil {
			mw := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(mw)
		}
	}

	// 启动 WiFi 状态监控
	// go wifi.StartMonitoring()

	// 使用阿里云NTP服务器同步时间
	exec.Command("ntpdate", "-u", "ntp1.aliyun.com")

	// 启动 HTTP 服务器
	log.Print("Aku Web 服务启动")
	if err := server.Start(); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}
