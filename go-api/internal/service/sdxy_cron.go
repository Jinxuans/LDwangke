package service

import (
	"context"
	"fmt"
	"time"

	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func RunSDXYCron(ctx context.Context) {
	svc := SDXY()

	// 启动30秒后开始首次同步
	if !sleepWithContext(ctx, 30*time.Second) {
		return
	}

	for {
		sdxyCronSync(svc)
		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}

func sdxyCronSync(svc *SDXYService) {
	cfg, err := svc.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return
	}

	page := 1
	limit := 200
	updateTotal := 0

	for {
		result, err := svc.upstreamRequest("orders", map[string]string{
			"page":  fmt.Sprintf("%d", page),
			"limit": fmt.Sprintf("%d", limit),
		})
		if err != nil {
			obslogger.L().Warn("SDXY Cron 获取订单失败", "error", err)
			break
		}

		code := mapGetFloat(result, "code")
		if code != 0 {
			msg := mapGetString(result, "msg")
			obslogger.L().Warn("SDXY Cron 上游返回错误", "message", msg)
			break
		}

		dataArr, ok := result["data"].([]interface{})
		if !ok {
			break
		}

		total := len(dataArr)
		updateTotal += total
		page++

		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				aggOrderId := mapGetString(m, "agg_order_id")
				status := mapGetString(m, "status")
				if aggOrderId != "" && status != "" {
					database.DB.Exec(
						"UPDATE qingka_wangke_flash_sdxy SET status = ? WHERE agg_order_id = ? LIMIT 1",
						status, aggOrderId,
					)
				}
			}
		}

		if total < limit {
			break
		}
	}

	if updateTotal > 0 {
		obslogger.L().Info("SDXY Cron 同步订单状态完成", "updated", updateTotal)
	}
}
