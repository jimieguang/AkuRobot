package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type SystemInfo struct {
	CPU struct {
		Usage     float64 `json:"usage"`
		NumCPU    int     `json:"num_cpu"`     // CPU核心数
		GoMaxProc int     `json:"go_max_proc"` // Go程序可用的最大CPU数
	} `json:"cpu"`
	Memory struct {
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
	} `json:"memory"`
	System struct {
		OS           string    `json:"os"`            // 操作系统
		Architecture string    `json:"architecture"`  // 系统架构
		NumGoroutine int       `json:"num_goroutine"` // 当前goroutine数量
		GoVersion    string    `json:"go_version"`    // Go版本
		StartTime    time.Time `json:"start_time"`    // 程序启动时间
		WorkDir      string    `json:"work_dir"`      // 工作目录
		Hostname     string    `json:"hostname"`      // 主机名
	} `json:"system"`
}

var (
	startTime    = time.Now()
	lastCPUUsage float64
	lastCPUCheck time.Time
)

// getCPUUsage 获取CPU使用率
func getCPUUsage() float64 {
	// 读取 /proc/stat 文件获取CPU信息
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return lastCPUUsage // 如果读取失败，返回上次的值
	}

	var user, nice, system, idle, iowait, irq, softirq, steal uint64
	_, err = fmt.Sscanf(string(data), "cpu %d %d %d %d %d %d %d %d",
		&user, &nice, &system, &idle, &iowait, &irq, &softirq, &steal)
	if err != nil {
		return lastCPUUsage
	}

	idle_total := idle + iowait
	non_idle := user + nice + system + irq + softirq + steal
	total := idle_total + non_idle

	now := time.Now()
	if !lastCPUCheck.IsZero() {
		// 计算时间差
		timeDiff := now.Sub(lastCPUCheck).Seconds()
		if timeDiff > 0 {
			cpuUsage := (float64(total-lastTotal) - float64(idle_total-lastIdleTotal)) / float64(total-lastTotal) * 100
			lastCPUUsage = cpuUsage
		}
	}

	// 保存当前值用于下次计算
	lastTotal = total
	lastIdleTotal = idle_total
	lastCPUCheck = now

	return lastCPUUsage
}

var (
	lastTotal     uint64
	lastIdleTotal uint64
)

// HandleSystemInfo 处理系统信息请求
func HandleSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := SystemInfo{}

	// 获取CPU信息
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// 获取内存信息
	info.Memory.Total = stats.Sys
	info.Memory.Used = stats.Alloc

	// CPU相关信息
	info.CPU.Usage = getCPUUsage()
	info.CPU.NumCPU = runtime.NumCPU()
	info.CPU.GoMaxProc = runtime.GOMAXPROCS(0)

	// 系统信息
	info.System.OS = runtime.GOOS
	info.System.Architecture = runtime.GOARCH
	info.System.NumGoroutine = runtime.NumGoroutine()
	info.System.GoVersion = runtime.Version()
	info.System.StartTime = startTime

	// 获取工作目录
	if workDir, err := os.Getwd(); err == nil {
		info.System.WorkDir = filepath.Clean(workDir)
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.System.Hostname = hostname
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
