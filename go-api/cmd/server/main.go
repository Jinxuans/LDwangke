package main

import (
	"log"
	"time"

	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/handler"
	"go-api/internal/middleware"
	"go-api/internal/queue"
	"go-api/internal/service"
	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load("config/config.yaml")

	// 连接数据库
	database.Connect(cfg.Database)

	// 连接 Redis
	cache.Connect(cfg.Redis)

	// 初始化对接并发队列（5并发，1000缓冲）
	// checker: 查 DB dockstatus=1 判断对接是否成功，用于准确统计 completed/failed
	dockChecker := func(oid int64) bool {
		var ds int
		err := database.DB.QueryRow("SELECT dockstatus FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&ds)
		return err == nil && ds == 1
	}
	queue.GlobalDockQueue = queue.NewDockQueue(5, 1000, service.DockSingleOrder, dockChecker)
	queue.GlobalDockQueue.Start()

	// 恢复未完成的待对接订单
	go func() {
		var oids []int64
		rows, err := database.DB.Query("SELECT oid FROM qingka_wangke_order WHERE dockstatus = 0 ORDER BY oid ASC LIMIT 500")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var oid int64
				rows.Scan(&oid)
				oids = append(oids, oid)
			}
		}
		if len(oids) > 0 {
			log.Printf("[DockQueue] 恢复 %d 个待对接订单", len(oids))
			queue.GlobalDockQueue.PushBatch(oids)
		}
	}()

	// 定时同步"进行中"订单的进度（可被狂暴模式热替换间隔）
	service.InitSyncTicker(2 * time.Minute)

	// 聊天消息定时清理（每天凌晨3点）
	go func() {
		chatSvc := service.NewChatService()
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))

			archived, err := chatSvc.ArchiveOldMessages()
			if err != nil {
				log.Printf("[ChatCleanup] 归档失败: %v", err)
			} else if archived > 0 {
				log.Printf("[ChatCleanup] 归档了 %d 条过期消息", archived)
			}

			trimmed, err := chatSvc.TrimSessionMessages()
			if err != nil {
				log.Printf("[ChatCleanup] 截断失败: %v", err)
			} else if trimmed > 0 {
				log.Printf("[ChatCleanup] 截断了 %d 条超限消息", trimmed)
			}
		}
	}()

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 初始化路由
	r := gin.Default()
	r.Use(middleware.CORS())
	r.Static("/uploads", "./uploads")

	// ===== 公开路由 =====
	loginLimiter := middleware.NewRateLimiter(10, time.Minute)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", loginLimiter.Handler(), handler.Login)
		auth.POST("/register", handler.Register)
		auth.POST("/refresh-token", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)
		auth.POST("/send-code", handler.SendCode)
		auth.POST("/forgot-password", handler.ForgotPassword)
		auth.POST("/reset-password", handler.ResetPassword)
	}

	// ===== 公开路由（无需认证） =====
	r.GET("/api/v1/site/config", handler.SiteConfigGet)

	// ===== 外部API（密钥认证，对应 PHP apisub.php 密钥调用） =====
	openapi := r.Group("/api/v1/open", middleware.APIKeyAuth())
	{
		openapi.GET("/classlist", handler.OpenAPIGetClass)
		openapi.POST("/classlist", handler.OpenAPIGetClass)
		openapi.GET("/query", handler.OpenAPIQuery)
		openapi.POST("/query", handler.OpenAPIQuery)
		openapi.GET("/order", handler.OpenAPIAddOrder)
		openapi.POST("/order", handler.OpenAPIAddOrder)
		openapi.GET("/orderlist", handler.OpenAPIOrderList)
		openapi.POST("/orderlist", handler.OpenAPIOrderList)
		openapi.GET("/balance", handler.OpenAPIBalance)
	}

	// ===== 需要认证的路由 =====
	api := r.Group("/api/v1", middleware.JWTAuth())
	{
		// 认证相关（需 token）
		api.GET("/user/info", handler.UserInfo)
		api.GET("/auth/codes", handler.AccessCodes)

		// 订单
		order := api.Group("/order")
		{
			order.POST("/list", handler.OrderList)
			order.GET("/list", handler.OrderList)
			order.GET("/stats", handler.OrderStats)
			order.GET("/:oid", handler.OrderDetail)
			order.POST("/add", handler.OrderAdd)
			order.POST("/status", handler.OrderChangeStatus)
			order.POST("/cancel", handler.OrderCancel)
			order.POST("/cancel/:oid", handler.OrderCancel)
			order.POST("/refund", handler.OrderRefund)
			order.GET("/pause", handler.OrderPause)
			order.POST("/changepass", handler.OrderChangePassword)
			order.GET("/resubmit", handler.OrderResubmit)
			order.GET("/logs", handler.OrderLogs)
		}

		// 聊天
		chat := api.Group("/chat")
		{
			chat.GET("/sessions", handler.ChatSessions)
			chat.GET("/messages/:list_id", handler.ChatMessages)
			chat.GET("/history/:list_id", handler.ChatHistory)
			chat.GET("/new/:list_id", handler.ChatNew)
			chat.POST("/send", handler.ChatSend)
			chat.POST("/send-image", handler.ChatSendImage)
			chat.POST("/read/:list_id", handler.ChatMarkRead)
			chat.GET("/unread", handler.ChatUnread)
			chat.POST("/create", handler.ChatCreate)
		}

		// 课程
		class := api.Group("/class")
		{
			class.GET("/list", handler.ClassList)
			class.GET("/search", handler.ClassSearch)
			class.POST("/search", handler.ClassQueryCourse)
			class.GET("/categories", handler.ClassCategories)
			class.GET("/category-switches", handler.ClassCategorySwitches)
		}

		// 站内信
		mail := api.Group("/mail")
		{
			mail.GET("/list", handler.MailList)
			mail.GET("/unread", handler.MailUnread)
			mail.GET("/:id", handler.MailDetail)
			mail.POST("/send", handler.MailSend)
			mail.POST("/upload", handler.MailUpload)
			mail.DELETE("/:id", handler.MailDelete)
		}

		// 公告（用户端）
		api.GET("/announcements", handler.AnnouncementListPublic)

		// 动态模块（运动/实习/论文）
		api.GET("/modules", handler.ModuleList)
		api.GET("/module/:app_id/frame-url", handler.ModuleFrameURL)
		api.Any("/module/:app_id", handler.ModuleProxy)

		// 用户中心
		uc := api.Group("/user")
		{
			uc.GET("/profile", handler.UserProfile)
			uc.POST("/change-password", handler.UserChangePassword)
			uc.POST("/change-pass2", handler.UserChangePass2)
			uc.POST("/change-email/code", handler.SendChangeEmailCode)
			uc.POST("/change-email", handler.ChangeEmail)
			uc.GET("/pay/channels", handler.UserPayChannels)
			uc.POST("/pay", handler.UserCreatePay)
			uc.GET("/pay/orders", handler.UserPayOrders)
			uc.GET("/moneylog", handler.UserMoneyLog)
			uc.GET("/tickets", handler.UserTicketList)
			uc.POST("/ticket/create", handler.UserTicketCreate)
			uc.POST("/ticket/reply", handler.UserTicketReply)
			uc.POST("/ticket/close/:id", handler.UserTicketClose)
			uc.GET("/favorites", handler.UserGetFavorites)
			uc.POST("/favorite/add", handler.UserAddFavorite)
			uc.POST("/favorite/remove", handler.UserRemoveFavorite)
			uc.POST("/pay/check", handler.UserCheckPayStatus)
			uc.POST("/invite-code", handler.UserSetInviteCode)
			uc.GET("/grades", handler.UserGradeList)
			uc.POST("/set-grade", handler.UserSetMyGrade)
			uc.POST("/invite-rate", handler.UserSetInviteRate)
			uc.POST("/secret-key", handler.UserChangeSecretKey)
			uc.POST("/push-token", handler.UserSetPushToken)
			uc.GET("/logs", handler.UserLogList)
			uc.POST("/checkin", handler.UserCheckin)
			uc.GET("/checkin/status", handler.UserCheckinStatus)
		}

		// 代理管理（所有登录用户可访问）
		agent := api.Group("/agent")
		{
			agent.POST("/list", handler.AgentList)
			agent.POST("/create", handler.AgentCreate)
			agent.POST("/recharge", handler.AgentRecharge)
			agent.POST("/deduct", handler.AgentDeduct)
			agent.POST("/change-grade", handler.AgentChangeGrade)
			agent.POST("/change-status", handler.AgentChangeStatus)
			agent.POST("/reset-password", handler.AgentResetPassword)
			agent.POST("/open-key", handler.AgentOpenSecretKey)
			agent.POST("/set-invite-code", handler.AgentSetInviteCode)
			agent.POST("/migrate-superior", handler.AgentMigrateSuperior)
		}

		// 管理后台（需要管理员权限）
		admin := api.Group("/admin", middleware.AdminOnly())
		{
			admin.POST("/impersonate", handler.Impersonate)
			admin.GET("/dashboard", handler.AdminDashboard)
			admin.GET("/users", handler.AdminUserList)
			admin.POST("/user/reset-pass", handler.AdminUserResetPass)
			admin.POST("/user/balance", handler.AdminUserSetBalance)
			admin.POST("/user/grade", handler.AdminUserSetGrade)
			admin.GET("/categories", handler.AdminCategoryList)
			admin.GET("/categories/paged", handler.AdminCategoryListPaged)
			admin.POST("/category/save", handler.AdminCategorySave)
			admin.DELETE("/category/:id", handler.AdminCategoryDelete)
			admin.POST("/category/quick-modify", handler.AdminCategoryQuickModify)
			admin.GET("/classes", handler.AdminClassList)
			admin.POST("/class/save", handler.AdminClassSave)
			admin.POST("/class/toggle", handler.AdminClassToggle)
			admin.POST("/class/batch-delete", handler.AdminClassBatchDelete)
			admin.POST("/class/batch-category", handler.AdminClassBatchCategory)
			admin.POST("/class/batch-price", handler.AdminClassBatchPrice)
			admin.GET("/suppliers", handler.AdminSupplierList)
			admin.POST("/supplier/save", handler.AdminSupplierSave)
			admin.GET("/supplier/products", handler.AdminSupplierProducts)
			admin.GET("/supplier/balance", handler.AdminSupplierBalance)
			admin.POST("/class/add", handler.AdminAddClass)
			admin.GET("/supplier/import", handler.AdminSupplierImport)
			admin.GET("/supplier/sync-status", handler.AdminSupplierSyncStatus)
			admin.POST("/category/batch-toggle", handler.AdminCategoryBatchToggle)
			admin.POST("/order/dock", handler.OrderManualDock)
			admin.POST("/order/redock-pending", handler.AdminRedockPending)
			admin.POST("/order/sync", handler.OrderSyncProgress)
			admin.POST("/order/batch-sync", handler.OrderBatchSync)
			admin.POST("/order/batch-resend", handler.OrderBatchResend)
			admin.POST("/order/remarks", handler.OrderModifyRemarks)
			admin.GET("/platform-names", handler.AdminPlatformNames)
			admin.DELETE("/supplier/:hid", handler.AdminSupplierDelete)
			admin.GET("/moneylog", handler.AdminMoneyLog)
			admin.GET("/announcements", handler.AdminAnnouncementList)
			admin.POST("/announcement/save", handler.AdminAnnouncementSave)
			admin.DELETE("/announcement/:id", handler.AdminAnnouncementDelete)
			admin.GET("/stats", handler.AdminStats)
			admin.GET("/config", handler.AdminConfigGet)
			admin.POST("/config", handler.AdminConfigSave)
			admin.GET("/paydata", handler.AdminPayDataGet)
			admin.POST("/paydata", handler.AdminPayDataSave)
			admin.GET("/grades", handler.AdminGradeList)
			admin.POST("/grade/save", handler.AdminGradeSave)
			admin.DELETE("/grade/:id", handler.AdminGradeDelete)
			admin.GET("/class/dropdown", handler.AdminClassDropdown)
			admin.GET("/mijia", handler.AdminMiJiaList)
			admin.POST("/mijia/save", handler.AdminMiJiaSave)
			admin.POST("/mijia/delete", handler.AdminMiJiaDelete)
			admin.POST("/mijia/batch", handler.AdminMiJiaBatch)
			admin.GET("/order/pause", handler.OrderPause)
			admin.POST("/order/changepass", handler.OrderChangePassword)
			admin.GET("/order/resubmit", handler.OrderResubmit)
			admin.GET("/order/logs", handler.OrderLogs)
			admin.GET("/queue/stats", handler.AdminQueueStats)
			admin.POST("/queue/concurrency", handler.AdminQueueSetConcurrency)
			admin.GET("/rank/suppliers", handler.AdminSupplierRanking)
			admin.GET("/rank/agent-products", handler.AdminAgentProductRanking)
			admin.GET("/chat/sessions", handler.AdminChatSessions)
			admin.GET("/chat/messages/:list_id", handler.AdminChatMessages)
			admin.GET("/chat/stats", handler.AdminChatStats)
			admin.POST("/chat/cleanup", handler.AdminChatCleanup)
			admin.POST("/email/send", handler.AdminEmailSend)
			admin.GET("/email/logs", handler.AdminEmailLogs)
			admin.GET("/email/preview", handler.AdminEmailPreview)
			admin.GET("/smtp/config", handler.AdminSMTPGet)
			admin.POST("/smtp/config", handler.AdminSMTPSave)
			admin.POST("/smtp/test", handler.AdminSMTPTest)
			// 工单管理
			admin.GET("/tickets", handler.AdminTicketList)
			admin.GET("/ticket/stats", handler.AdminTicketStats)
			admin.POST("/ticket/reply", handler.AdminTicketReply)
			admin.POST("/ticket/close/:id", handler.AdminTicketClose)
			admin.POST("/ticket/auto-close", handler.AdminTicketAutoClose)
			admin.POST("/ticket/report", handler.AdminTicketReport)
			admin.POST("/ticket/sync-report", handler.AdminTicketSyncReport)
			admin.GET("/ticket/order-counts", handler.OrderTicketCounts)
			// 动态模块管理
			admin.GET("/modules", handler.ModuleListAll)
			admin.POST("/module/save", handler.AdminModuleSave)
			admin.DELETE("/module/:id", handler.AdminModuleDelete)
			// 平台配置管理
			admin.GET("/platform-configs", handler.AdminPlatformConfigList)
			admin.POST("/platform-config/save", handler.AdminPlatformConfigSave)
			admin.DELETE("/platform-config/:pt", handler.AdminPlatformConfigDelete)
			admin.POST("/platform-config/parse-php", handler.AdminParsePHPCode)
			admin.POST("/platform-config/detect", handler.AdminDetectPlatform)
			// 商品同步监控
			admin.GET("/sync/config", handler.SyncGetConfig)
			admin.POST("/sync/config", handler.SyncSaveConfig)
			admin.GET("/sync/preview", handler.SyncPreview)
			admin.POST("/sync/execute", handler.SyncExecute)
			admin.GET("/sync/logs", handler.SyncLogs)
			admin.GET("/sync/suppliers", handler.SyncMonitoredSuppliers)

			// 邮箱轮询池
			admin.GET("/email-pool", handler.AdminEmailPoolList)
			admin.POST("/email-pool/save", handler.AdminEmailPoolSave)
			admin.DELETE("/email-pool/:id", handler.AdminEmailPoolDelete)
			admin.POST("/email-pool/toggle", handler.AdminEmailPoolToggle)
			admin.POST("/email-pool/test", handler.AdminEmailPoolTest)
			admin.GET("/email-pool/stats", handler.AdminEmailPoolStats)
			admin.POST("/email-pool/reset-counters", handler.AdminEmailPoolResetCounters)
			admin.GET("/email-send-logs", handler.AdminEmailSendLogs)

			// 邮件模板
			admin.GET("/email-templates", handler.AdminEmailTemplateList)
			admin.POST("/email-templates/save", handler.AdminEmailTemplateSave)
			admin.GET("/email-templates/preview", handler.AdminEmailTemplatePreview)
			admin.POST("/email-templates/test", handler.AdminEmailTemplateTest)

			// 运维看板
			admin.GET("/ops/dashboard", handler.AdminOpsDashboard)
			admin.GET("/ops/probe-suppliers", handler.AdminOpsProbeSuppliers)
			admin.GET("/ops/table-sizes", handler.AdminOpsTableSizes)

			// 狂暴模式
			admin.GET("/ops/turbo", handler.AdminGetTurbo)
			admin.POST("/ops/turbo", handler.AdminSetTurbo)

			// 租户管理
			admin.GET("/tenants", handler.AdminTenantList)
			admin.POST("/tenant/create", handler.AdminTenantCreate)
			admin.POST("/tenant/:tid/status", handler.AdminTenantSetStatus)

			// 签到管理
			admin.GET("/checkin/stats", handler.AdminCheckinStats)
		}
	}

	// ===== 商城：C端公开接口 =====
	// 商城支付回调（不带 :tid，全局路由）
	r.POST("/api/v1/mall/pay/notify", handler.MallPayNotify)
	r.GET("/api/v1/mall/pay/notify", handler.MallPayNotify)

	mall := r.Group("/api/v1/mall/:tid")
	{
		mall.GET("/info", handler.MallShopInfo)
		mall.POST("/login", handler.MallCUserLogin)
		mall.GET("/products", handler.MallProductList)
		mall.GET("/product/:cid", handler.MallProductDetail)
		mall.POST("/query", handler.MallQueryCourse)
		mall.GET("/pay/channels", handler.MallPayChannels)
		mall.POST("/pay", handler.MallCreatePay)
		mall.POST("/order", handler.MallOrderAdd)
		mall.GET("/search", handler.MallOrderSearch)
		mall.GET("/orders", handler.MallOrderList)
		mall.GET("/order/:oid", handler.MallOrderDetail)
		mall.GET("/pay/check", handler.MallCheckPay)
		mall.POST("/pay/confirm", handler.MallConfirmPay)
	}

	// ===== B端后台（登录用户）=====
	tenant := api.Group("/tenant")
	{
		tenant.GET("/mall-open-price", handler.TenantMallOpenPrice)
		tenant.POST("/mall-open", handler.TenantMallOpen)
		tenant.GET("/shop", handler.TenantShopGet)
		tenant.POST("/shop", handler.TenantShopSave)
		tenant.POST("/pay-config", handler.TenantPayConfigSave)
		tenant.GET("/products", handler.TenantProductList)
		tenant.POST("/product/save", handler.TenantProductSave)
		tenant.DELETE("/product/:cid", handler.TenantProductDelete)
		tenant.GET("/order/stats", handler.TenantOrderStats)
		tenant.GET("/mall-orders", handler.TenantMallOrders)
		tenant.GET("/cusers", handler.TenantCUserList)
		tenant.POST("/cuser/save", handler.TenantCUserSave)
		tenant.DELETE("/cuser/:id", handler.TenantCUserDelete)
	}

	// ===== WebSocket 推送 =====
	r.GET("/ws/push", middleware.WSAuth(), ws.HandlePush(hub))

	// 启动服务
	addr := ":" + cfg.Server.Port
	log.Printf("Go API 启动于 %s (模式: %s)", addr, cfg.Server.Mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
