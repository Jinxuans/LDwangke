package service

import (
	"sync/atomic"
	"time"
)

// ===== 运维看板服务 =====

type OpsService struct{}

var opsService = &OpsService{}

func GetOpsDashboard() OpsDashboard {
	return opsService.GetDashboard()
}

func ProbeSuppliers() []SupplierProbe {
	return opsService.ProbeSuppliers()
}

func GetOpsTableSizes() []TableSize {
	return opsService.GetTableSizes()
}

var startTime = time.Now()

var (
	opsErrCount   int64
	opsDockFail   int64
	opsHTTPErrors int64
)

func opsErrorStatsSnapshot() ErrorStats {
	return ErrorStats{
		ErrorCounter:   atomic.LoadInt64(&opsErrCount),
		DockFailCount:  atomic.LoadInt64(&opsDockFail),
		HTTPErrorCount: atomic.LoadInt64(&opsHTTPErrors),
	}
}
