package supplier

import (
	"fmt"

	"go-api/internal/database"
)

func (s *Service) SyncSupplierStatus(hid int) (int, string, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return 0, "", err
	}

	classList, err := s.GetSupplierClasses(sup)
	if err != nil {
		return 0, "", err
	}

	apiCIDs := map[string]bool{}
	for _, item := range classList {
		apiCIDs[item.CID] = true
	}

	rows, err := database.DB.Query(
		"SELECT cid, noun, status FROM qingka_wangke_class WHERE docking = ?", hid,
	)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var downIDs []int
	for rows.Next() {
		var cid int
		var noun string
		var status int
		rows.Scan(&cid, &noun, &status)
		if !apiCIDs[noun] && status != 0 {
			downIDs = append(downIDs, cid)
		}
	}
	if len(downIDs) == 0 {
		return 0, "没有需要更新的商品状态", nil
	}
	for _, cid := range downIDs {
		database.DB.Exec("UPDATE qingka_wangke_class SET status = 0 WHERE cid = ?", cid)
	}
	return len(downIDs), fmt.Sprintf("共下架%d个商品", len(downIDs)), nil
}

func (s *Service) SetSupplierClassStatus(hid, status int, dryRun bool) (int, int, string, error) {
	if _, err := s.GetSupplierByHID(hid); err != nil {
		return 0, 0, "", err
	}

	var total int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_class WHERE docking = ?",
		hid,
	).Scan(&total); err != nil {
		return 0, 0, "", err
	}

	var changed int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_class WHERE docking = ? AND COALESCE(status, 0) <> ?",
		hid, status,
	).Scan(&changed); err != nil {
		return 0, 0, "", err
	}

	action := "上架"
	if status == 0 {
		action = "下架"
	}
	if total == 0 {
		return total, changed, "当前货源没有本地商品", nil
	}
	if dryRun {
		return total, changed, fmt.Sprintf("预计%s%d个商品", action, changed), nil
	}
	if changed == 0 {
		return total, changed, fmt.Sprintf("当前货源下%d个商品已全部%s", total, action), nil
	}

	result, err := database.DB.Exec(
		"UPDATE qingka_wangke_class SET status = ?, addtime = NOW() WHERE docking = ? AND COALESCE(status, 0) <> ?",
		status, hid, status,
	)
	if err != nil {
		return total, 0, "", err
	}
	affected, _ := result.RowsAffected()
	return total, int(affected), fmt.Sprintf("已%s%d个商品", action, affected), nil
}
