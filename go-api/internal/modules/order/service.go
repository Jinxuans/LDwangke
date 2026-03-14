package order

// Services 聚合订单模块的查询、命令与同步服务。
type Services struct {
	Query   *QueryService
	Command *CommandService
	Sync    *SyncService
}

func NewServices() *Services {
	repo := NewRepository()
	return &Services{
		Query:   NewQueryService(repo),
		Command: NewCommandService(repo),
		Sync:    NewSyncService(repo),
	}
}
