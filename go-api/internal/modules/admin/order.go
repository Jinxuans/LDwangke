package admin

import (
	"fmt"

	"go-api/internal/dockscheduler"
	ordermodule "go-api/internal/modules/order"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerOrderRoutes(admin *gin.RouterGroup) {
	ordermodule.SetOrderStatusNotifier(notifyAdminOrderStatusChange)
	ordermodule.RegisterAdminRoutes(admin)
	admin.POST("/order/redock-pending", AdminRedockPending)
}

func AdminRedockPending(c *gin.Context) {
	stats, err := dockscheduler.RunOnce("manual")
	if err != nil {
		response.ServerErrorf(c, err, fmt.Sprintf("执行失败: %v", err))
		return
	}
	if stats.LastFetched == 0 {
		response.Success(c, "无待对接订单")
		return
	}
	response.Success(c, fmt.Sprintf("本轮处理 %d 个订单，成功 %d 个，失败 %d 个", stats.LastFetched, stats.LastSuccess, stats.LastFail))
}
