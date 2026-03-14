package xm

import (
	"net/http"
	"time"
)

// XMProject 小米运动项目
type XMProject struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Query       int     `json:"query"`
	Password    int     `json:"password"`
}

// XMOrder 小米运动订单
type XMOrder struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	School    string      `json:"school"`
	Account   string      `json:"account"`
	Password  string      `json:"password"`
	ProjectID int         `json:"project_id"`
	Status    string      `json:"status_name"`
	Type      interface{} `json:"type"`
	Pace      *float64    `json:"pace"`
	Distance  *float64    `json:"distance"`
	TotalKM   int         `json:"total_km"`
	IsDeleted bool        `json:"is_deleted"`
	RunKM     *float64    `json:"run_km"`
	RunDate   interface{} `json:"run_date"`
	StartDay  string      `json:"start_day"`
	StartTime string      `json:"start_time"`
	EndTime   string      `json:"end_time"`
	Deduction float64     `json:"deduction"`
	UpdatedAt string      `json:"updated_at"`
}

type XMService struct {
	client *http.Client
}

var xmService = &XMService{
	client: &http.Client{Timeout: 15 * time.Second},
}

func XM() *XMService {
	return xmService
}
