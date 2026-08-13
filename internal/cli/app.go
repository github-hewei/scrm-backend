package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/241x/zero-kit/gormutil"
	"github.com/241x/zero-kit/locker"
	"github.com/241x/zero-kit/logger"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

type ctxKey struct{}

var lockKey ctxKey

// App CLI 应用框架：封装 root 命令、分布式锁、traceID、信号处理和优雅关闭。
type App struct {
	root *cobra.Command
	l    logger.Logger
	rdb  *redis.Client
}

// New 创建 CLI 应用实例。
func New(l logger.Logger, rdb *redis.Client) *App {
	root := &cobra.Command{
		Use:   "cli",
		Short: "CLI 工具",
		Long:  `命令行管理工具`,
	}
	root.PersistentFlags().IntP("instance-id", "i", 0, "实例编号, 防止并发执行")

	a := &App{root: root, l: l, rdb: rdb}
	a.setupHooks()
	return a
}

// AddCommand 注册子命令。
func (a *App) AddCommand(cmd *cobra.Command) {
	a.root.AddCommand(cmd)
}

// Run 启动命令，处理信号与优雅关闭。命令执行失败时返回非 0 退出码
func (a *App) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	exitCode := 0
	done := make(chan int, 1)
	go func() {
		if err := a.root.ExecuteContext(ctx); err != nil {
			var cost time.Duration
			if t, ok := ctxkeys.BeginTime(a.root.Context()); ok {
				cost = time.Since(t)
			}
			logger.Ctx(a.root.Context()).Err(err, "Error Command", "cost", cost)
			exitCode = 1
		}
		done <- exitCode
	}()

	select {
	case <-sigChan:
		a.l.Info("Exiting manually...")
		cancel()
		timeout := time.NewTimer(10 * time.Second)
		defer timeout.Stop()
		select {
		case code := <-done:
			return code
		case <-timeout.C:
			a.l.Warn("Timeout exit")
			return 1
		}
	case code := <-done:
		return code
	}
}

// setupHooks 设置命令行的预运行和后运行钩子。
func (a *App) setupHooks() {
	a.root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		traceID := uuid.New().String()
		l := a.l.With("traceId", traceID)
		l.Info("Start Command")

		ctx := l.WithContext(cmd.Context())
		ctx = ctxkeys.WithTraceID(ctx, traceID)
		ctx = gormutil.WithTraceID(ctx, traceID)
		ctx = ctxkeys.WithBeginTime(ctx, time.Now())

		id, _ := cmd.Flags().GetInt("instance-id")
		key := fmt.Sprintf("%s:%d", strings.ReplaceAll(cmd.CommandPath(), " ", "-"), id)
		lock, err := locker.NewRedisLocker(a.rdb).Lock(ctx, key, locker.WithTTL(time.Minute), locker.WithWatchDog())
		if err != nil {
			return err
		}
		l.Info("Locked Command", "key", key)

		cmd.SetContext(context.WithValue(ctx, lockKey, lock))
		return nil
	}

	a.root.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if lock, ok := ctx.Value(lockKey).(locker.Lock); ok && lock != nil {
			unlockCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			lock.Unlock(unlockCtx)
			logger.Ctx(ctx).Info("Unlocked Command")
		}
		if t, ok := ctxkeys.BeginTime(ctx); ok {
			logger.Ctx(ctx).Info("End Command", "cost", fmt.Sprintf("%.4f", time.Since(t).Seconds()))
		}
	}
}
