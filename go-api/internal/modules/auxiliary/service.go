package auxiliary

type Service struct{}

var auxiliaryService = &Service{}

func Auxiliary() *Service {
	return auxiliaryService
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
