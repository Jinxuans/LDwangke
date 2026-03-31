package sdxy

import (
	"net/http"
	"time"
)

type SDXYConfig struct {
	BaseURL  string  `json:"base_url"`
	Endpoint string  `json:"endpoint"`
	UID      string  `json:"uid"`
	Key      string  `json:"key"`
	Timeout  int     `json:"timeout"`
	Price    float64 `json:"price"`
}

type SDXYOrder struct {
	ID          int    `json:"id"`
	UID         int    `json:"uid"`
	AggOrderID  string `json:"agg_order_id"`
	SDXYOrderID string `json:"sdxy_order_id"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	School      string `json:"school"`
	Num         int    `json:"num"`
	Distance    string `json:"distance"`
	RunType     string `json:"run_type"`
	RunRule     string `json:"run_rule"`
	Pause       int    `json:"pause"`
	Status      string `json:"status"`
	Fees        string `json:"fees"`
	CreatedAt   string `json:"created_at"`
}

type SDXYService struct {
	client *http.Client
}

var sdxyService = &SDXYService{client: &http.Client{Timeout: 30 * time.Second}}

func SDXY() *SDXYService {
	return sdxyService
}
