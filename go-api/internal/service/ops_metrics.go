package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-api/internal/database"
)

type ErrorStats struct {
	TodayFailed    int   `json:"today_failed"`
	TodayException int   `json:"today_exception"`
	PendingDock    int   `json:"pending_dock"`
	StuckOrders    int   `json:"stuck_orders"`
	ErrorCounter   int64 `json:"error_counter"`
	DockFailCount  int64 `json:"dock_fail_count"`
	HTTPErrorCount int64 `json:"http_error_count"`
}

func (s *OpsService) GetErrorStats() ErrorStats {
	es := opsErrorStatsSnapshot()
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '失败' AND DATE(addtime) = CURDATE()").Scan(&es.TodayFailed)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常' AND DATE(addtime) = CURDATE()").Scan(&es.TodayException)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE dockstatus IN (0, 2)").Scan(&es.PendingDock)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '进行中' AND addtime < NOW() - INTERVAL 24 HOUR").Scan(&es.StuckOrders)
	return es
}

type TableSize struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	DataMB  string `json:"data_mb"`
	IndexMB string `json:"index_mb"`
	TotalMB string `json:"total_mb"`
}

func (s *OpsService) GetTableSizes() []TableSize {
	rows, err := database.DB.Query(`
		SELECT TABLE_NAME, TABLE_ROWS, 
			ROUND(DATA_LENGTH/1024/1024, 2),
			ROUND(INDEX_LENGTH/1024/1024, 2),
			ROUND((DATA_LENGTH + INDEX_LENGTH)/1024/1024, 2)
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY (DATA_LENGTH + INDEX_LENGTH) DESC
		LIMIT 15
	`)
	if err != nil {
		return []TableSize{}
	}
	defer rows.Close()

	var result []TableSize
	for rows.Next() {
		var t TableSize
		rows.Scan(&t.Name, &t.Rows, &t.DataMB, &t.IndexMB, &t.TotalMB)
		result = append(result, t)
	}
	if result == nil {
		result = []TableSize{}
	}
	return result
}

type StorageInfo struct {
	UploadsSize  string `json:"uploads_size"`
	UploadsFiles int    `json:"uploads_files"`
}

func (s *OpsService) GetStorageInfo() StorageInfo {
	si := StorageInfo{}
	var totalSize int64
	var fileCount int

	filepath.Walk("./uploads", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	si.UploadsFiles = fileCount
	if totalSize < 1024*1024 {
		si.UploadsSize = fmt.Sprintf("%.1f KB", float64(totalSize)/1024)
	} else if totalSize < 1024*1024*1024 {
		si.UploadsSize = fmt.Sprintf("%.1f MB", float64(totalSize)/1024/1024)
	} else {
		si.UploadsSize = fmt.Sprintf("%.2f GB", float64(totalSize)/1024/1024/1024)
	}
	return si
}

type RecentErrorOrder struct {
	OID     int    `json:"oid"`
	User    string `json:"user"`
	PtName  string `json:"ptname"`
	Status  string `json:"status"`
	AddTime string `json:"addtime"`
}

func (s *OpsService) GetRecentErrorOrders(limit int) []RecentErrorOrder {
	if limit <= 0 {
		limit = 20
	}
	rows, err := database.DB.Query(
		"SELECT oid, COALESCE(user,''), COALESCE(ptname,''), COALESCE(status,''), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_order WHERE status IN ('异常','失败') ORDER BY oid DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return []RecentErrorOrder{}
	}
	defer rows.Close()

	var list []RecentErrorOrder
	for rows.Next() {
		var o RecentErrorOrder
		rows.Scan(&o.OID, &o.User, &o.PtName, &o.Status, &o.AddTime)
		list = append(list, o)
	}
	if list == nil {
		list = []RecentErrorOrder{}
	}
	return list
}

type HourlyOrder struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

func (s *OpsService) GetTodayHourlyOrders() []HourlyOrder {
	rows, err := database.DB.Query(
		"SELECT HOUR(addtime) as h, COUNT(*) as cnt FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() GROUP BY HOUR(addtime) ORDER BY h ASC",
	)
	if err != nil {
		return []HourlyOrder{}
	}
	defer rows.Close()

	hourMap := make(map[int]int)
	for rows.Next() {
		var h, cnt int
		rows.Scan(&h, &cnt)
		hourMap[h] = cnt
	}

	now := time.Now().Hour()
	result := make([]HourlyOrder, 0, now+1)
	for i := 0; i <= now; i++ {
		result = append(result, HourlyOrder{Hour: i, Count: hourMap[i]})
	}
	return result
}

type OpsDashboard struct {
	System        SystemInfo             `json:"system"`
	DB            DBHealth               `json:"db"`
	Redis         RedisHealth            `json:"redis"`
	WS            WSStatus               `json:"ws"`
	DockScheduler map[string]interface{} `json:"dock_scheduler"`
	Errors        ErrorStats             `json:"errors"`
	Storage       StorageInfo            `json:"storage"`
	Tables        []TableSize            `json:"tables"`
	ErrorOrders   []RecentErrorOrder     `json:"error_orders"`
	HourlyOrders  []HourlyOrder          `json:"hourly_orders"`
}

func (s *OpsService) GetDashboard() OpsDashboard {
	return OpsDashboard{
		System:        s.GetSystemInfo(),
		DB:            s.GetDBHealth(),
		Redis:         s.GetRedisHealth(),
		WS:            s.GetWSStatus(),
		DockScheduler: s.GetDockSchedulerStats(),
		Errors:        s.GetErrorStats(),
		Storage:       s.GetStorageInfo(),
		Tables:        s.GetTableSizes(),
		ErrorOrders:   s.GetRecentErrorOrders(20),
		HourlyOrders:  s.GetTodayHourlyOrders(),
	}
}
