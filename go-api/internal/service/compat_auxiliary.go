package service

import "go-api/internal/model"

type AuxiliaryService struct{}

var auxiliaryService = &AuxiliaryService{}

func Auxiliary() *AuxiliaryService {
	return auxiliaryService
}

func (s *AuxiliaryService) CardKeyList(req model.CardKeyListRequest) ([]model.CardKey, int, error) {
	return s.cardKeyList(req)
}

func (s *AuxiliaryService) ActivityList(req model.ActivityListRequest) ([]model.Activity, int, error) {
	return s.activityList(req)
}

func (s *AuxiliaryService) PledgeRecordList(req model.PledgeListRequest) ([]model.PledgeRecord, int, error) {
	return s.pledgeRecordList(req)
}

func normalizeListPage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizeListLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}
