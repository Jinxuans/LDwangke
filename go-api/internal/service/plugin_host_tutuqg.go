package service

import (
	"net/http"
	"time"
)

type TutuQGConfig struct {
	BaseURL        string  `json:"base_url"`
	Key            string  `json:"key"`
	PriceIncrement float64 `json:"price_increment"`
}

type TutuQGOrder struct {
	OID     int     `json:"oid"`
	UID     int     `json:"uid"`
	User    string  `json:"user"`
	Pass    string  `json:"pass"`
	KCName  string  `json:"kcname"`
	Days    string  `json:"days"`
	PTName  string  `json:"ptname"`
	Fees    string  `json:"fees"`
	AddTime string  `json:"addtime"`
	IP      *string `json:"IP"`
	Status  *string `json:"status"`
	Remarks *string `json:"remarks"`
	GUID    *string `json:"guid"`
	Score   string  `json:"score"`
	Scores  *string `json:"scores"`
	ZDXF    *string `json:"zdxf"`
}

type TutuQGService struct {
	client *http.Client
}

var tutuqgService = &TutuQGService{
	client: &http.Client{Timeout: 15 * time.Second},
}

func TutuQG() *TutuQGService {
	return tutuqgService
}
