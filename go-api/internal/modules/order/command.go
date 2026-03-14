package order

import "go-api/internal/model"

type CommandService struct {
	repo Repository
}

func NewCommandService(repo Repository) *CommandService {
	return &CommandService{repo: repo}
}

func (s *CommandService) Add(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	return s.repo.AddOrders(uid, req)
}

func (s *CommandService) AddForMall(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	return s.repo.AddOrdersForMall(bUID, tid, cUID, retailPrice, req)
}

func (s *CommandService) ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error {
	return s.repo.ChangeStatus(uid, grade, req)
}

func (s *CommandService) Cancel(uid int, grade string, oid int) error {
	return s.repo.CancelOrder(uid, grade, oid)
}

func (s *CommandService) Refund(uid int, grade string, oids []int) error {
	return s.repo.RefundOrders(uid, grade, oids)
}

func (s *CommandService) ModifyRemarks(oids []int, remarks string) error {
	return s.repo.ModifyRemarks(oids, remarks)
}
