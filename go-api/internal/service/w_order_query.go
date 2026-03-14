package service

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go-api/internal/database"
)

func (s *WService) GetApps(uid int) ([]WAppUser, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	rows, err := database.DB.Query("SELECT id, org_app_id, code, name, COALESCE(description,''), price, cac_type FROM w_app WHERE deleted = 0 AND status = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []WAppUser
	for rows.Next() {
		var a WAppUser
		if err := rows.Scan(&a.AppID, &a.OrgAppID, &a.Code, &a.Name, &a.Desc, &a.Price, &a.CacType); err != nil {
			continue
		}
		a.Price = math.Round(a.Price*addprice*100) / 100
		list = append(list, a)
	}
	if list == nil {
		list = []WAppUser{}
	}
	return list, nil
}

func (s *WService) GetOrders(uid int, isAdmin bool, page, pageSize int, filters map[string]string) ([]WOrder, int, error) {
	offset := (page - 1) * pageSize
	where := "o.deleted = 0"
	var args []interface{}

	if !isAdmin {
		where += " AND o.user_id = ?"
		args = append(args, uid)
	}
	if v := filters["account"]; v != "" {
		where += " AND o.account = ?"
		args = append(args, v)
	}
	if v := filters["school"]; v != "" {
		where += " AND o.school = ?"
		args = append(args, v)
	}
	if v := filters["status"]; v != "" {
		where += " AND o.status = ?"
		args = append(args, v)
	}
	if v := filters["app_id"]; v != "" && v != "0" {
		where += " AND o.app_id = ?"
		args = append(args, v)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM w_order o WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`SELECT o.id, o.agg_order_id, o.user_id, COALESCE(o.school,''), o.account, o.password,
		o.app_id, COALESCE(a.name,''), o.status, o.num, o.cost, o.pause, o.sub_order, o.deleted, o.created, o.updated
		FROM w_order o LEFT JOIN w_app a ON o.app_id = a.id
		WHERE %s ORDER BY o.id DESC LIMIT ?, ?`, where)
	args = append(args, offset, pageSize)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []WOrder
	for rows.Next() {
		var o WOrder
		var pauseInt, deletedInt int
		var aggOrderID *string
		var subOrderStr *string
		var createdTime, updatedTime time.Time
		err := rows.Scan(&o.ID, &aggOrderID, &o.UserID, &o.School, &o.Account, &o.Password,
			&o.AppID, &o.AppName, &o.Status, &o.Num, &o.Cost, &pauseInt, &subOrderStr, &deletedInt, &createdTime, &updatedTime)
		if err != nil {
			continue
		}
		o.AggOrderID = aggOrderID
		o.Pause = pauseInt == 1
		o.Deleted = deletedInt == 1
		o.Created = createdTime.Format("2006-01-02 15:04:05")
		o.Updated = updatedTime.Format("2006-01-02 15:04:05")
		if subOrderStr != nil {
			var sub interface{}
			json.Unmarshal([]byte(*subOrderStr), &sub)
			o.SubOrder = sub
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []WOrder{}
	}
	return orders, total, nil
}
