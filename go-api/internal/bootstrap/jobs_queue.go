package bootstrap

import (
	"context"
	"log"

	"go-api/internal/app"
	"go-api/internal/database"
)

// StartDockQueue 启动对接队列并恢复未完成的待对接订单。
func StartDockQueue(ctx context.Context, a *app.App) {
	if a == nil || a.DockQueue == nil {
		return
	}

	a.DockQueue.Start()

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var oids []int64
		rows, err := database.DB.Query("SELECT oid FROM qingka_wangke_order WHERE dockstatus = 0 ORDER BY oid ASC LIMIT 500")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var oid int64
				if err := rows.Scan(&oid); err != nil {
					continue
				}
				oids = append(oids, oid)
			}
		}
		if len(oids) > 0 {
			log.Printf("[DockQueue] 恢复 %d 个待对接订单", len(oids))
			a.DockQueue.PushBatch(oids)
		}
	}()
}
