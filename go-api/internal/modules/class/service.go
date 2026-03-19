package class

import (
	"fmt"
	"strconv"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
)

type classService struct{}

var classes = &classService{}

func Classes() *classService {
	return classes
}

func loadUserAddPrice(uid int) float64 {
	var addprice float64
	err := database.DB.QueryRow("SELECT COALESCE(addprice, 1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil || addprice == 0 {
		return 1
	}
	return addprice
}

func buildPublicClassWhere(uid int, req model.ClassListRequest) (string, []interface{}) {
	where := []string{"c.status = 1"}
	args := []interface{}{}

	if req.Fenlei > 0 {
		where = append(where, "c.fenlei = ?")
		args = append(args, req.Fenlei)
	}
	if req.Search != "" {
		where = append(where, "c.name LIKE ?")
		args = append(args, "%"+req.Search+"%")
	}
	if req.Favorite == 1 {
		where = append(where, "EXISTS (SELECT 1 FROM qingka_wangke_user_favorite uf WHERE uf.uid = ? AND uf.cid = c.cid)")
		args = append(args, uid)
	}

	return strings.Join(where, " AND "), args
}

func loadClassesWithPricing(uid int, query string, args ...interface{}) ([]model.Class, error) {
	addprice := loadUserAddPrice(uid)

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

	mijiaMap := map[int]MiJiaRule{}
	if len(cids) > 0 {
		if loaded, err := LoadMiJiaMap(uid, cids); err == nil {
			mijiaMap = loaded
		}
	}

	var list []model.Class
	for _, item := range tempList {
		c := item.c
		basePrice, _ := strconv.ParseFloat(c.Price, 64)

		price := ComputeClassBasePrice(basePrice, addprice, item.yunsuan, 4)

		if mj, ok := mijiaMap[c.CID]; ok {
			adjustedPrice, _, applied := ApplyMiJia(basePrice, addprice, item.yunsuan, mj.Mode, mj.Price, 4)
			price = adjustedPrice
			if applied {
				c.Name = "【密价】" + c.Name
			}
		}

		c.Price = fmt.Sprintf("%.2f", price)
		list = append(list, c)
	}

	if list == nil {
		list = []model.Class{}
	}
	return list, nil
}

// ListClasses 获取课程列表（按 PHP getclass case: 应用 addprice 加价 + yunsuan 运算符 + mijia 密价）
func (s *classService) ListClasses(uid int, req model.ClassListRequest) ([]model.Class, error) {
	where, args := buildPublicClassWhere(uid, req)
	query := "SELECT c.cid, COALESCE(c.name,''), COALESCE(c.noun,''), COALESCE(c.price,'0'), COALESCE(c.docking,'0'), COALESCE(c.fenlei,''), COALESCE(c.status,1), COALESCE(c.sort,10), COALESCE(c.content,''), COALESCE(c.yunsuan,'*') FROM qingka_wangke_class c WHERE " + where + " ORDER BY c.sort ASC, c.cid ASC"
	return loadClassesWithPricing(uid, query, args...)
}

func (s *classService) ListClassesPaged(uid int, req model.ClassListRequest) ([]model.Class, int64, int, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	where, args := buildPublicClassWhere(uid, req)

	var total int64
	countQuery := "SELECT COUNT(*) FROM qingka_wangke_class c WHERE " + where
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, req.Page, req.Limit, err
	}

	offset := (req.Page - 1) * req.Limit
	argsWithPage := append(append([]interface{}{}, args...), req.Limit, offset)
	query := "SELECT c.cid, COALESCE(c.name,''), COALESCE(c.noun,''), COALESCE(c.price,'0'), COALESCE(c.docking,'0'), COALESCE(c.fenlei,''), COALESCE(c.status,1), COALESCE(c.sort,10), COALESCE(c.content,''), COALESCE(c.yunsuan,'*') FROM qingka_wangke_class c WHERE " + where + " ORDER BY c.sort ASC, c.cid ASC LIMIT ? OFFSET ?"
	list, err := loadClassesWithPricing(uid, query, argsWithPage...)
	if err != nil {
		return nil, 0, req.Page, req.Limit, err
	}

	return list, total, req.Page, req.Limit, nil
}
