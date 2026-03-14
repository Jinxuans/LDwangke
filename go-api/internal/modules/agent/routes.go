package agent

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册代理域路由。
func RegisterRoutes(api *gin.RouterGroup) {
	agent := api.Group("/agent")
	{
		agent.POST("/list", AgentList)
		agent.POST("/create", AgentCreate)
		agent.POST("/recharge", AgentRecharge)
		agent.POST("/deduct", AgentDeduct)
		agent.POST("/change-grade", AgentChangeGrade)
		agent.POST("/change-status", AgentChangeStatus)
		agent.POST("/reset-password", AgentResetPassword)
		agent.POST("/open-key", AgentOpenSecretKey)
		agent.POST("/set-invite-code", AgentSetInviteCode)
		agent.POST("/migrate-superior", AgentMigrateSuperior)
		agent.GET("/cross-recharge-check", AgentCrossRechargeCheck)
		agent.POST("/cross-recharge", AgentCrossRecharge)
	}
}
