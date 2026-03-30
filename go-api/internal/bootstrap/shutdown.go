package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api/internal/app"
	obslogger "go-api/internal/observability/logger"
)

// NotifyContext 返回带系统信号监听的上下文。
func NotifyContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Serve 启动 HTTP 服务，并在收到退出信号后执行优雅停机。
func Serve(ctx context.Context, srv *http.Server, a *app.App) error {
	errCh := make(chan error, 1)

	go func() {
		obslogger.L().Info("Go API 启动", "addr", srv.Addr, "mode", a.Config.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = a.Close(closeCtx)
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = a.Close(shutdownCtx)
			return err
		}
		return a.Close(shutdownCtx)
	}
}
