package service

import (
	"fmt"
	"log"
	"sync"
)

// ServiceRegistry 全局服务注册表
var (
	registry = make(map[string]Service)
	regMux   sync.RWMutex
)

// GetService 获取服务实例，如果不存在则创建
func GetService(name string) (Service, error) {
	regMux.Lock()
	defer regMux.Unlock()

	if svc, exists := registry[name]; exists {
		return svc, nil
	}

	var svc Service
	switch name {
	case "xiaozhi":
		svc = NewXiaozhiService()
	default:
		return nil, fmt.Errorf("unknown service: %s", name)
	}

	registry[name] = svc
	return svc, nil
}

// Service 定义第三方服务的接口
type Service interface {
	Start() error
	Stop() error
	GetStatus() Status
	GetOutput() <-chan string
}

// Status 表示服务的状态
type Status struct {
	Running bool
	Name    string
	Error   string
}

// BaseService 提供基础的服务实现
type BaseService struct {
	name     string
	output   chan string
	mu       sync.Mutex
	running  bool
	stopChan chan struct{}
}

// NewBaseService 创建一个基础服务实例
func NewBaseService(name string) *BaseService {
	return &BaseService{
		name:     name,
		output:   make(chan string, 100),
		stopChan: make(chan struct{}),
	}
}

// Start 启动服务
func (s *BaseService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("service %s is already running", s.name)
	}

	// 重新初始化channels
	s.output = make(chan string, 1000)
	s.stopChan = make(chan struct{})
	s.running = true
	return nil
}

// Stop 停止服务
func (s *BaseService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("service %s is not running", s.name)
	}

	close(s.stopChan)
	s.running = false
	return nil
}

// GetStatus 获取服务状态
func (s *BaseService) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Status{
		Running: s.running,
		Name:    s.name,
	}
}

// GetOutput 获取服务输出通道
func (s *BaseService) GetOutput() <-chan string {
	return s.output
}

// SendOutput 安全地发送输出信息
func (s *BaseService) SendOutput(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		select {
		case s.output <- msg:
		default:
			// 如果channel已满，移除最旧的消息
			select {
			case <-s.output:
				s.output <- msg
			default:
				log.Printf("Failed to send message")
			}
		}
	} else {
		log.Printf("%s Service not running, message not sent: %s", s.name, msg)
	}
}
