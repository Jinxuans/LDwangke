package service

import (
	"fmt"
	"math"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
)

type ClassService struct{}

func NewClassService() *ClassService {
	return &ClassService{}
}

// List 获取课程列表（按 PHP getclass case: 应用 addprice 加价 + yunsuan 运算符 + mijia 密价）
func (s *ClassService) List(uid int, req model.ClassListRequest) ([]model.Class, error) {
	// 获取用户的加价系数 addprice（按 PHP: $userrow['addprice']）
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

	// 收集课程和CID列表
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

	// 批量查询密价 (按 PHP: SELECT * FROM qingka_wangke_mijia WHERE uid=? AND cid IN (...))
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

	// 应用价格计算 (按 PHP getclass case)
	var classes []model.Class
	for _, item := range tempList {
		c := item.c
		basePrice, _ := strconv.ParseFloat(c.Price, 64)

		// 按 yunsuan 计算基础价格 (按 PHP: if yunsuan == "+" 则加法, 否则乘法)
		var price float64
		if item.yunsuan == "+" {
			price = math.Round((basePrice+addprice)*100) / 100
		} else {
			price = math.Round((basePrice*addprice)*100) / 100
		}
		price1 := price // 保存原始加价后价格

		// 应用密价 (按 PHP: mode 0=减价, 1=底价*加价, 2=固定价, 4=倍率定价)
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

		// 密价不能高于原价 (按 PHP: if ($price > $price1) $price = $price1)
		if price > price1 {
			price = price1
		}

		c.Price = fmt.Sprintf("%.2f", price)
		classes = append(classes, c)
	}

	if classes == nil {
		classes = []model.Class{}
	}
	return classes, nil
}

func (s *ClassService) Categories() ([]model.Category, error) {
	rows, err := database.DB.Query(
		"SELECT id, name, sort, status, COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0) FROM qingka_wangke_fenlei WHERE status >= 1 ORDER BY recommend DESC, sort ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Sort, &c.Status, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []model.Category{}
	}
	return categories, nil
}
