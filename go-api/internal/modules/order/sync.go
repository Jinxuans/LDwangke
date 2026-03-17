package order

type SyncService struct {
	repo Repository
}

type AutoSyncRule struct {
	Key             string `json:"key"`
	Label           string `json:"label"`
	MinAgeHours     int    `json:"min_age_hours"`
	MaxAgeHours     int    `json:"max_age_hours"`
	IntervalMinutes int    `json:"interval_minutes"`
	Enabled         bool   `json:"enabled"`
}

type AutoSyncOptions struct {
	RecentHours      int
	SupplierHIDs     []int
	ExcludedStatuses []string
	Rules            []AutoSyncRule
}

// NewSyncService 只负责装配同步相关用例，不承载业务逻辑。
// 真正的“如何查上游、如何更新订单、自动轮询怎么跑”都下沉在 Repository 中，
// 这样手动同步、批量同步、自动同步三种入口可以共用同一套核心规则。
func NewSyncService(repo Repository) *SyncService {
	return &SyncService{repo: repo}
}

// ManualDock 负责把未对接订单推送到上游，不负责查询进度。
func (s *SyncService) ManualDock(oids []int) (int, int, error) {
	return s.repo.ManualDockOrders(oids)
}

// SyncProgress 是“手动查询上游进度”的服务入口。
// 这里不加业务判断，直接把订单 oid 列表下发给 Repository 处理，
// 保持服务层只做用例编排，不重复实现查询细节。
func (s *SyncService) SyncProgress(oids []int) (int, error) {
	return s.repo.SyncOrderProgress(oids)
}

// AutoSyncAllProgress 是后台定时任务使用的全局自动轮询入口。
// 它和手动同步共享同一套供应商查询能力，但调度范围是“所有已对接主订单”。
func (s *SyncService) AutoSyncAllProgress(opts AutoSyncOptions) (int, int, error) {
	return s.repo.AutoSyncAllProgress(opts)
}

// BatchSync 保留批量同步的独立方法名，便于前端接口和旧调用方继续使用。
func (s *SyncService) BatchSync(oids []int) (int, error) {
	return s.repo.BatchSyncOrders(oids)
}

func (s *SyncService) BatchResend(oids []int) (int, int, error) {
	return s.repo.BatchResendOrders(oids)
}
