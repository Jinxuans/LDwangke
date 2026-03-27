package user

import (
	auxmodule "go-api/internal/modules/auxiliary"
	checkinmodule "go-api/internal/modules/checkin"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes 注册无需登录的充值支付回调路由。
func RegisterPublicRoutes(r *gin.Engine) {
	r.POST("/api/v1/pay/notify", UserPayNotify)
	r.GET("/api/v1/pay/notify", UserPayNotify)
}

// RegisterRoutes 注册用户中心路由。
func RegisterRoutes(api *gin.RouterGroup) {
	uc := api.Group("/user")
	{
		uc.GET("/profile", UserProfile)
		uc.POST("/change-password", UserChangePassword)
		uc.POST("/change-pass2", UserChangePass2)
		uc.POST("/change-email/code", SendChangeEmailCode)
		uc.POST("/change-email", ChangeEmail)
		uc.GET("/pay/channels", UserPayChannels)
		uc.POST("/pay", UserCreatePay)
		uc.GET("/pay/orders", UserPayOrders)
		uc.GET("/moneylog", UserMoneyLog)
		uc.GET("/withdraw/requests", UserWithdrawRequests)
		uc.POST("/withdraw/request", UserWithdrawCreate)
		uc.GET("/tickets", UserTicketList)
		uc.POST("/ticket/create", UserTicketCreate)
		uc.POST("/ticket/reply", UserTicketReply)
		uc.POST("/ticket/close/:id", UserTicketClose)
		uc.GET("/favorites", UserGetFavorites)
		uc.POST("/favorite/add", UserAddFavorite)
		uc.POST("/favorite/remove", UserRemoveFavorite)
		uc.POST("/pay/check", UserCheckPayStatus)
		uc.POST("/invite-code", UserSetInviteCode)
		uc.GET("/grades", UserGradeList)
		uc.POST("/set-grade", UserSetMyGrade)
		uc.POST("/invite-rate", UserSetInviteRate)
		uc.POST("/secret-key", UserChangeSecretKey)
		uc.POST("/push-token", UserSetPushToken)
		uc.GET("/logs", UserLogList)
		uc.POST("/checkin", checkinmodule.UserCheckin)
		uc.GET("/checkin/status", checkinmodule.UserCheckinStatus)
		uc.POST("/cardkey/use", auxmodule.UserCardKeyUse)
	}
}
