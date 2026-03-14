package service

import (
	"net/http"
	"sync"
	"time"
)

const (
	zwBaseURL       = "http://www.zwgflw.top/"
	zwTokenCacheKey = "zhiwen_paper_token"
	zwTokenTTL      = 23 * time.Hour
	zwMaxRetry      = 3
)

type PaperConfig struct {
	Username string            `json:"lunwen_api_username"`
	Password string            `json:"lunwen_api_password"`
	Prices   map[string]string `json:"prices"`
}

type PaperOrder struct {
	ID       int     `json:"id"`
	UID      int     `json:"uid"`
	OrderID  string  `json:"order_id"`
	ShopCode string  `json:"shopcode"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
}

type PaperService struct {
	client *http.Client
	mu     sync.RWMutex
}

var paperService *PaperService
var paperServiceOnce sync.Once

func Paper() *PaperService {
	paperServiceOnce.Do(func() {
		paperService = &PaperService{
			client: &http.Client{Timeout: 120 * time.Second},
		}
	})
	return paperService
}
