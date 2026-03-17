package order

import "sync"

type AutoSyncReport struct {
	SupplierNames []string `json:"supplier_names"`
	SampleErrors  []string `json:"sample_errors"`
	Processed     int      `json:"processed"`
}

var (
	autoSyncReportMu sync.RWMutex
	autoSyncReport   AutoSyncReport
)

func setLastAutoSyncReport(report AutoSyncReport) {
	autoSyncReportMu.Lock()
	defer autoSyncReportMu.Unlock()
	autoSyncReport = report
}

func GetLastAutoSyncReport() AutoSyncReport {
	autoSyncReportMu.RLock()
	defer autoSyncReportMu.RUnlock()
	return AutoSyncReport{
		SupplierNames: append([]string(nil), autoSyncReport.SupplierNames...),
		SampleErrors:  append([]string(nil), autoSyncReport.SampleErrors...),
		Processed:     autoSyncReport.Processed,
	}
}
