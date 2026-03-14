package order

import (
	"fmt"
	"strings"

	"go-api/internal/database"
)

func categorySwitchesByOID(oid int) (log, ticket, changepass, allowpause, supplierReport, supplierReportHID int, err error) {
	changepass = 1
	err = database.DB.QueryRow(
		`SELECT COALESCE(f.log,0), COALESCE(f.ticket,0), COALESCE(f.changepass,1), COALESCE(f.allowpause,0), COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		 FROM qingka_wangke_order o
		 JOIN qingka_wangke_class c ON c.cid = o.cid
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE o.oid = ?`, oid,
	).Scan(&log, &ticket, &changepass, &allowpause, &supplierReport, &supplierReportHID)
	if err != nil {
		return 0, 0, 1, 0, 0, 0, nil
	}
	return
}

func ticketCountByOIDs(oids []int) map[int]int {
	result := make(map[int]int)
	if len(oids) == 0 {
		return result
	}
	placeholders := make([]string, len(oids))
	args := make([]interface{}, len(oids))
	for i, oid := range oids {
		placeholders[i] = "?"
		args[i] = oid
	}
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT oid, COUNT(*) FROM qingka_wangke_ticket WHERE oid IN (%s) GROUP BY oid", strings.Join(placeholders, ",")),
		args...,
	)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var oid, cnt int
		rows.Scan(&oid, &cnt)
		result[oid] = cnt
	}
	return result
}
