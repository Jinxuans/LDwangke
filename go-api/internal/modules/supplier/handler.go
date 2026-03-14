package supplier

import (
	"strconv"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var supplierService = SharedService()

func AdminPlatformNames(c *gin.Context) {
	response.Success(c, GetPlatformNames())
}

func AdminSupplierBalance(c *gin.Context) {
	hid, err := strconv.Atoi(c.Query("hid"))
	if hid <= 0 || err != nil {
		response.BadRequest(c, "请指定供应商hid")
		return
	}
	result, err := supplierService.QueryBalance(hid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func AdminSupplierProducts(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}
	sup, err := supplierService.GetSupplierByHID(hid)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	classes, err := supplierService.GetSupplierClasses(sup)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	localNouns := map[string]bool{}
	rows, dbErr := database.DB.Query("SELECT noun FROM qingka_wangke_class WHERE docking = ? AND status >= 0", hid)
	if dbErr == nil {
		defer rows.Close()
		for rows.Next() {
			var noun string
			rows.Scan(&noun)
			localNouns[noun] = true
		}
	}

	type productItem struct {
		CID          string  `json:"cid"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		Fenlei       string  `json:"fenlei"`
		Content      string  `json:"content"`
		CategoryName string  `json:"category_name"`
		States       int     `json:"states"`
		Sort         int     `json:"sort"`
	}

	var list []productItem
	for _, item := range classes {
		states := 0
		if localNouns[item.CID] {
			states = 1
		}
		list = append(list, productItem{
			CID:          item.CID,
			Name:         item.Name,
			Price:        item.Price,
			Fenlei:       item.Fenlei,
			Content:      item.Content,
			CategoryName: item.CategoryName,
			States:       states,
			Sort:         10,
		})
	}
	if list == nil {
		list = []productItem{}
	}
	response.Success(c, list)
}

func AdminSupplierImport(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	pricee, _ := strconv.ParseFloat(c.Query("pricee"), 64)
	category := c.Query("category")
	name := c.Query("name")
	fd, _ := strconv.Atoi(c.Query("fd"))

	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}
	if pricee <= 0 {
		pricee = 1
	}
	if category == "" {
		category = "999999"
	}

	inserted, updated, msg, err := supplierService.ImportSupplierClasses(hid, pricee, category, name, fd)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"inserted": inserted,
		"updated":  updated,
		"msg":      msg,
	})
}

func AdminSupplierSyncStatus(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}

	count, msg, err := supplierService.SyncSupplierStatus(hid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"count": count,
		"msg":   msg,
	})
}
