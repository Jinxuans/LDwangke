package order

type SyncService struct {
	repo Repository
}

func NewSyncService(repo Repository) *SyncService {
	return &SyncService{repo: repo}
}

func (s *SyncService) ManualDock(oids []int) (int, int, error) {
	return s.repo.ManualDockOrders(oids)
}

func (s *SyncService) SyncProgress(oids []int) (int, error) {
	return s.repo.SyncOrderProgress(oids)
}

func (s *SyncService) BatchSync(oids []int) (int, error) {
	return s.repo.BatchSyncOrders(oids)
}

func (s *SyncService) BatchResend(oids []int) (int, int, error) {
	return s.repo.BatchResendOrders(oids)
}
