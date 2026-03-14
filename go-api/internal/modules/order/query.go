package order

import "go-api/internal/model"

type QueryService struct {
	repo Repository
}

func NewQueryService(repo Repository) *QueryService {
	return &QueryService{repo: repo}
}

func (s *QueryService) List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error) {
	return s.repo.List(uid, grade, req)
}

func (s *QueryService) Detail(uid int, grade string, oid int) (*model.Order, error) {
	return s.repo.Detail(uid, grade, oid)
}

func (s *QueryService) Stats(uid int, grade string) (*model.OrderStats, error) {
	return s.repo.Stats(uid, grade)
}
