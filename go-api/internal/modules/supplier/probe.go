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
