package app

import (
	"context"
	"errors"

	"go-api/internal/pluginruntime"
)

// Close 统一关闭可关闭的基础设施资源。
func (a *App) Close(_ context.Context) error {
	var err error

	if a == nil {
		return nil
	}
	if a.License != nil {
		a.License.Stop()
	}
	if a.Hub != nil {
		a.Hub.Stop()
	}
	if a.DockQueue != nil {
		a.DockQueue.Stop()
	}
	pluginruntime.StopHZWSocket()
	if a.DB != nil {
		err = errors.Join(err, a.DB.Close())
	}
	if a.Redis != nil {
		err = errors.Join(err, a.Redis.Close())
	}

	return err
}
