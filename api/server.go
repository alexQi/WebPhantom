package api

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"noctua/api/middlewares"
	"noctua/api/routes"
	"noctua/kernel"
	"noctua/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	App    *iris.Application
	Kernel *kernel.Kernel
}

func NewServer(k *kernel.Kernel) *Server {
	app := iris.New()
	// 设置app logger
	app.Logger().Install(logger.Log)
	// 设置logger level
	app.Logger().SetLevel("info")
	// 加载跨域中间件
	app.Use(middlewares.CorsNew())
	// 加载recover
	app.Use(recover.New())
	// 初始化Server
	server := &Server{App: app, Kernel: k}
	// 加载路由
	routes.ApiRoutes(server.App, server.Kernel)

	return server
}

func (s *Server) Run(ctx context.Context) error {
	httpAddr := "0.0.0.0:8080"
	// 启动 Iris 服务
	go func() {
		if err := s.App.Run(iris.Addr(httpAddr), iris.WithoutServerError(iris.ErrServerClosed)); err != nil {
			logger.Log.Errorf("Failed to run Iris server: %v", err)
		}
	}()

	// 监听信号以优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan // 等待信号

	// 停止 Kernel
	s.Kernel.Stop()

	// 关闭 Iris
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.App.Shutdown(ctx); err != nil {
		logger.Log.Errorf("Failed to shutdown Iris: %v", err)
		return err
	}

	logger.Log.Info("Server shutdown gracefully")
	return nil
}
