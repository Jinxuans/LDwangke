package admin

import (
	"fmt"

	"go-api/internal/database"
	ordermodule "go-api/internal/modules/order"
	"go-api/internal/queue"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(admin *gin.RouterGroup) {
	ordermodule.SetOrderStatusNotifier(notifyAdminOrderStatusChange)
	ordermodule.RegisterAdminRoutes(admin)
	admin.POST("/order/redock-pending", AdminRedockPending)
}

func AdminRedockPending(c *gin.Context) {
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "对接队列未初始化")
		return
	}
	rows, err := database.DB.Query("SELECT oid FROM qingka_wangke_order WHERE dockstatus = 0")
	if err != nil {
		response.ServerError(c, fmt.Sprintf("查询失败: %v", err))
		return
	}
	defer rows.Close()

	var oids []int64
	for rows.Next() {
		var oid int64
		rows.Scan(&oid)
		oids = append(oids, oid)
	}
	if len(oids) == 0 {
		response.Success(c, "无待对接订单")
		return
	}
	queue.GlobalDockQueue.PushBatch(oids)
	response.Success(c, fmt.Sprintf("已推入 %d 个订单", len(oids)))
}
