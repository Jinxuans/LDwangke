package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

type CloneRequest struct {
	HID             int               `json:"hid"`
	PriceRate       float64           `json:"price_rate"`
	CloneCategory   bool              `json:"clone_category"`
	SkipCategories  []string          `json:"skip_categories"`
	NameReplace     map[string]string `json:"name_replace"`
	SecretPriceRate float64           `json:"secret_price_rate"`
}

// AdminCloneExecute 一键克隆（只克隆新增，不影响已有）
func AdminCloneExecute(c *gin.Context) {
	var req CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.HID <= 0 {
		response.BadRequest(c, "货源ID不能为空")
		return
	}
	if req.PriceRate <= 0 {
		req.PriceRate = 5 // 默认5倍
	}

	// 组装临时配置：只开克隆
	customCfg := &service.SyncConfig{
		SupplierIDs: strconv.Itoa(req.HID),
		PriceRates: map[string]float64{
			strconv.Itoa(req.HID): req.PriceRate,
		},
		CloneEnabled:    true,
		CloneCategory:   req.CloneCategory,
		SkipCategories:  req.SkipCategories,
		NameReplace:     req.NameReplace,
		SecretPriceRate: req.SecretPriceRate,

		// 其他同步全部关闭
		SyncPrice:   false,
		SyncStatus:  false,
		SyncContent: false,
		SyncName:    false,
	}

	result, err := service.SyncExecute(req.HID, customCfg)
	if err != nil {
		response.ServerError(c, "克隆失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "克隆完成",
		"result":  result,
	})
}

// AdminCloneUpdatePrices 更新价格（只更新价格，不克隆新商品，不改状态）
func AdminCloneUpdatePrices(c *gin.Context) {
	var req CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.HID <= 0 {
		response.BadRequest(c, "货源ID不能为空")
		return
	}
	if req.PriceRate <= 0 {
		req.PriceRate = 5 // 默认5倍
	}

	// 组装临时配置：只开价格同步
	customCfg := &service.SyncConfig{
		SupplierIDs: strconv.Itoa(req.HID),
		PriceRates: map[string]float64{
			strconv.Itoa(req.HID): req.PriceRate,
		},
		SkipCategories:  req.SkipCategories,
		SecretPriceRate: req.SecretPriceRate,

		SyncPrice: true, // 开启价格同步

		// 其他同步全部关闭
		CloneEnabled:  false,
		CloneCategory: false,
		SyncStatus:    false,
		SyncContent:   false,
		SyncName:      false,
	}

	result, err := service.SyncExecute(req.HID, customCfg)
	if err != nil {
		response.ServerError(c, "更新价格失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "价格更新完成",
		"result":  result,
	})
}

// AdminCloneAutoSync 自动同步（全量同步：克隆+价格+状态+说明）
func AdminCloneAutoSync(c *gin.Context) {
	var req CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.HID <= 0 {
		response.BadRequest(c, "货源ID不能为空")
		return
	}
	if req.PriceRate <= 0 {
		req.PriceRate = 5 // 默认5倍
	}

	// 组装临时配置：全开
	customCfg := &service.SyncConfig{
		SupplierIDs: strconv.Itoa(req.HID),
		PriceRates: map[string]float64{
			strconv.Itoa(req.HID): req.PriceRate,
		},
		CloneEnabled:  true,
		CloneCategory: true, // 自动同步时默认开分类克隆
		SyncPrice:     true,
		SyncStatus:    true,
		SyncContent:   true,
		SyncName:      false, // 名称默认不开，防覆盖
	}

	result, err := service.SyncExecute(req.HID, customCfg)
	if err != nil {
		response.ServerError(c, "同步失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "同步完成",
		"result":  result,
	})
}
