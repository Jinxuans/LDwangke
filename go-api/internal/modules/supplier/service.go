package supplier

import (
	shared "go-api/internal/shared/db"
	"net/http"
	"time"
)

// Service 为旧 SupplierService 提供模块边界。
type Service struct {
	suppliers *shared.SupplierRepo
	classes   *shared.ClassRepo
	client    *http.Client
}

type SupplierClassItem struct {
	CID          string  `json:"cid"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Fenlei       string  `json:"fenlei"`
	Content      string  `json:"content"`
	CategoryName string  `json:"category_name"`
}

var sharedHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func NewService() *Service {
	return &Service{
		suppliers: shared.NewSupplierRepo(),
		classes:   shared.NewClassRepo(),
		client:    sharedHTTPClient,
	}
}

var sharedService = NewService()

func SharedService() *Service {
	return sharedService
}
