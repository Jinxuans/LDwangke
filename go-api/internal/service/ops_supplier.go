package service

import (
	"net/http"
	"sync"
	"time"

	"go-api/internal/database"
)

type SupplierProbe struct {
	HID      int    `json:"hid"`
	Name     string `json:"name"`
	PT       string `json:"pt"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Latency  int64  `json:"latency_ms"`
	HTTPCode int    `json:"http_code"`
}

func (s *OpsService) ProbeSuppliers() []SupplierProbe {
	rows, err := database.DB.Query("SELECT hid, COALESCE(name,''), COALESCE(pt,''), COALESCE(url,'') FROM qingka_wangke_huoyuan WHERE status = 1")
	if err != nil {
		return []SupplierProbe{}
	}
	defer rows.Close()

	var suppliers []SupplierProbe
	for rows.Next() {
		var sp SupplierProbe
		rows.Scan(&sp.HID, &sp.Name, &sp.PT, &sp.URL)
		suppliers = append(suppliers, sp)
	}
	if len(suppliers) == 0 {
		return []SupplierProbe{}
	}

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 5 * time.Second}

	for i := range suppliers {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer func() {
				<-sem
				wg.Done()
			}()
			sp := &suppliers[idx]
			if sp.URL == "" {
				sp.Status = "no_url"
				return
			}
			probeURL := sp.URL
			if len(probeURL) > 4 && probeURL[:4] != "http" {
				probeURL = "http://" + probeURL
			}

			start := time.Now()
			resp, err := client.Get(probeURL)
			sp.Latency = time.Since(start).Milliseconds()
			if err != nil {
				sp.Status = "unreachable"
				return
			}
			resp.Body.Close()
			sp.HTTPCode = resp.StatusCode
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				sp.Status = "healthy"
			} else {
				sp.Status = "degraded"
			}
		}(i)
	}
	wg.Wait()
	return suppliers
}
