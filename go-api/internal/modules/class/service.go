package class

import (
	"fmt"
	"math"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
)

type classService struct{}

var classes = &classService{}

func Classes() *classService {
	return classes
}

// ListClasses 获取课程列表（按 PHP getclass case: 应用 addprice 加价 + yunsuan 运算符 + mijia 密价）
func (s *classService) ListClasses(uid int, req model.ClassListRequest) ([]model.Class, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT COALESCE(addprice, 1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil || addprice == 0 {
		addprice = 1
	}

	query := "SELECT cid, COALESCE(name,''), COALESCE(noun,''), COALESCE(price,'0'), COALESCE(docking,'0'), COALESCE(fenlei,''), COALESCE(status,1), COALESCE(sort,10), COALESCE(content,''), COALESCE(yunsuan,'*') FROM qingka_wangke_class WHERE status = 1"
	var args []interface{}

	if req.Fenlei > 0 {
		query += " AND fenlei = ?"
		args = append(args, req.Fenlei)
	}
	if req.Search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+req.Search+"%")
	}

	query += " ORDER BY sort ASC, cid ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type classWithYunsuan struct {
		c       model.Class
		yunsuan string
	}

	var tempList []classWithYunsuan
	var cids []int

	for rows.Next() {
		var c model.Class
		var yunsuan string
		if err := rows.Scan(&c.CID, &c.Name, &c.Noun, &c.Price, &c.Docking, &c.Fenlei, &c.Status, &c.Sort, &c.Content, &yunsuan); err != nil {
			continue
		}
		tempList = append(tempList, classWithYunsuan{c: c, yunsuan: yunsuan})
		cids = append(cids, c.CID)
	}

	mijiaMap := map[int]struct {
		mode  int
		price float64
	}{}
	if len(cids) > 0 {
		placeholders := ""
		mijiaArgs := []interface{}{uid}
		for i, cid := range cids {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			mijiaArgs = append(mijiaArgs, cid)
		}
		mijiaRows, err := database.DB.Query(
			fmt.Sprintf("SELECT cid, COALESCE(mode,0), COALESCE(price,0) FROM qingka_wangke_mijia WHERE uid = ? AND cid IN (%s)", placeholders),
			mijiaArgs...,
		)
		if err == nil {
			defer mijiaRows.Close()
			for mijiaRows.Next() {
				var cid, mode int
				var price float64
				mijiaRows.Scan(&cid, &mode, &price)
				mijiaMap[cid] = struct {
					mode  int
					price float64
				}{mode, price}
			}
		}
	}

	var list []model.Class
	for _, item := range tempList {
		c := item.c
		basePrice, _ := strconv.ParseFloat(c.Price, 64)

		var price float64
		if item.yunsuan == "+" {
			price = math.Round((basePrice+addprice)*100) / 100
		} else {
			price = math.Round((basePrice*addprice)*100) / 100
		}
		originalPrice := price

		if mj, ok := mijiaMap[c.CID]; ok {
			switch mj.mode {
			case 0:
				price = math.Round((price-mj.price)*100) / 100
			case 1:
				price = math.Round(((basePrice-mj.price)*addprice)*100) / 100
			case 2:
				price = mj.price
			case 4:
				price = math.Round((basePrice*mj.price)*100) / 100
			}
			if price <= 0 {
				price = 0
			}
			c.Name = "【质押】" + c.Name
		}

		if price > originalPrice {
			price = originalPrice
		}

		c.Price = fmt.Sprintf("%.2f", price)
		list = append(list, c)
	}

	if list == nil {
		list = []model.Class{}
	}
	return list, nil
}
