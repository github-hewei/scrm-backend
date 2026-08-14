package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/241x/zero-kit/job"
	"github.com/241x/zero-kit/logger"
	"github.com/241x/zero-kit/queue"
)

const (
	queueDefaultKey = "default"
	queueTestKey    = "test"
)

// Server 队列工作服务，管理队列消费、工作线程池与作业执行器。
type Server struct {
	manager   *queue.QueueManager
	executors []*job.Executor
	logger    logger.Logger
}

// NewServer 创建队列工作服务。
func NewServer(manager *queue.QueueManager, registry *Registry, log logger.Logger) *Server {
	config := queue.DefaultConfig().WithName(queueDefaultKey)
	pool, err := manager.RegisterWorkerPool(queueDefaultKey, registry, config)
	if err != nil {
		log.Err(err, "Failed to register default worker pool")
	} else {
		pool.WithLogger(log)
	}
	return &Server{manager: manager, logger: log}
}

// AddExecutor 注册作业执行器，与队列工作池一并启停。
func (s *Server) AddExecutor(ex *job.Executor) *Server {
	s.executors = append(s.executors, ex)
	return s
}

// Run 启动队列工作服务与作业执行器，阻塞直到收到退出信号。
func (s *Server) Run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.manager.StartAllWorkerPools(ctx); err != nil {
		s.logger.Err(err, "Failed to start worker pools")
		return
	}
	for _, ex := range s.executors {
		if err := ex.Start(ctx); err != nil {
			s.logger.Err(err, "Failed to start job executor")
			return
		}
	}

	s.logger.Info("Worker started, waiting for tasks...")

	<-sig
	s.logger.Info("Shutting down worker...")

	if err := s.manager.StopAllWorkerPools(); err != nil {
		s.logger.Err(err, "Failed to stop worker pools")
	}
	for _, ex := range s.executors {
		if err := ex.Stop(); err != nil {
			s.logger.Err(err, "Failed to stop job executor")
		}
	}

	s.logger.Info("Worker stopped")
}
