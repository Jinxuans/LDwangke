package xm

import (
	"context"
	"fmt"
	"time"

	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// RunXMCron 启动 XM 后台同步任务。
// 这里直接复用 modules/xm 的同步逻辑，避免 service 与 modules 双份实现长期漂移。
func RunXMCron(ctx context.Context) {
	obslogger.L().Info("XM 后台批量同步任务启动")
	go xmCronBatchSync(ctx)
}

func xmCronBatchSync(ctx context.Context) {
	if !sleepWithContext(ctx, 2*time.Minute+30*time.Second) {
		return
	}

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					obslogger.L().Error("XM cron sync panic", "panic", r)
				}
			}()

			rows, err := database.DB.Query(`
				SELECT y_oid, project_id
				FROM xm_order
				WHERE status NOT IN ('已取消', '已退款', '退款成功', '已完成', '已删除')
				  AND y_oid IS NOT NULL AND y_oid > 0
				  AND is_deleted = 0`)
			if err != nil {
				obslogger.L().Warn("XM cron sync 查询订单失败", "error", err)
				return
			}
			defer rows.Close()

			projectOrders := map[int][]int{}
			for rows.Next() {
				var yOid, projectID int
				if err := rows.Scan(&yOid, &projectID); err != nil {
					continue
				}
				if yOid > 0 && projectID > 0 {
					projectOrders[projectID] = append(projectOrders[projectID], yOid)
				}
			}

			if len(projectOrders) == 0 {
				return
			}

			svc := XM()
			for projectID, yOids := range projectOrders {
				projectRow, err := svc.getProjectRow(projectID)
				if err != nil {
					obslogger.L().Warn("XM cron sync 项目不存在", "project_id", projectID, "error", err)
					continue
				}

				projectName := fmt.Sprintf("%v", projectRow["name"])
				syncCount := 0
				for _, yOid := range yOids {
					if _, err := svc.syncOrderFromUpstream(yOid, projectRow); err != nil {
						obslogger.L().Warn("XM cron sync 订单同步失败", "project_name", projectName, "y_oid", yOid, "error", err)
					} else {
						syncCount++
					}
					if !sleepWithContext(ctx, 500*time.Millisecond) {
						return
					}
				}

				if syncCount > 0 {
					obslogger.L().Info("XM cron sync 同步完成", "project_name", projectName, "success", syncCount, "total", len(yOids))
				}
			}
		}()

		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}
