package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aku-web/internal/config"
)

var srv *http.Server

// Start 启动 HTTP 服务器
func Start() error {
	// 注册路由
	RegisterRoutes()

	// 创建 HTTP 服务器（限制了超时时间！！）
	srv = &http.Server{
		Addr:         fmt.Sprintf(":%s", config.DefaultPort),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 6 * time.Hour,
		IdleTimeout:  60 * time.Second,
	}

	// 创建通道监听系统信号
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// 在新的 goroutine 中启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s", config.DefaultPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	<-done
	log.Print("服务器正在关闭...")

	// 创建一个 5 秒的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅地关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭出错: %v", err)
		return err
	}

	log.Print("服务器已关闭")
	return nil
}
