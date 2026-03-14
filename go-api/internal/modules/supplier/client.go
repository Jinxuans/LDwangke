package supplier

import "go-api/internal/model"

func (s *Service) GetSupplierByHID(hid int) (*model.SupplierFull, error) {
	return s.suppliers.GetFullByHID(hid)
}

func (s *Service) GetClassFull(cid int) (*model.ClassFull, error) {
	return s.classes.GetFullByCID(cid)
}
