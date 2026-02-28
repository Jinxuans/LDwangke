package handler

import (
	"encoding/json"
	"net/http"

	"go-api/internal/database"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var tuboshuSvc = service.NewTuboshuService()

// TuboshuEnsureTable 启动时建表
func TuboshuEnsureTable() {
	tuboshuSvc.EnsureTable()
}

// TuboshuUserConfigGet 用户端获取土拨鼠配置（返回价格配置+页面可见性+用户倍率）
func TuboshuUserConfigGet(c *gin.Context) {
	uid := c.GetInt("uid")
	cfg, err := tuboshuSvc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	// 获取用户 addprice
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice == 0 {
		addprice = 1
	}
	response.Success(c, map[string]interface{}{
		"price_ratio":      cfg.PriceRatio,
		"price_config":     cfg.PriceConfig,
		"page_visibility":  cfg.PageVisibility,
		"user_price_ratio": addprice * cfg.PriceRatio,
	})
}

// TuboshuConfigGet 获取土拨鼠配置
func TuboshuConfigGet(c *gin.Context) {
	cfg, err := tuboshuSvc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

// TuboshuConfigSave 保存土拨鼠配置
func TuboshuConfigSave(c *gin.Context) {
	var cfg service.TuboshuConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := tuboshuSvc.SaveConfig(&cfg); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// TuboshuRoute 处理 tuboshu_route 请求（JSON API 代理）
func TuboshuRoute(c *gin.Context) {
	uid := c.GetInt("uid")
	grade, _ := c.Get("grade")
	isAdmin := grade == "2" || grade == "3"
	clientIP := c.ClientIP()

	var req service.TuboshuRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, isBlob, err := tuboshuSvc.HandleRoute(uid, isAdmin, req, clientIP)
	if err != nil {
		// 返回土拨鼠标准格式的错误
		c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Blob 响应（文件下载）
	if isBlob {
		if blobData, ok := result.([]byte); ok {
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=\"download.file\"")
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Data(http.StatusOK, "application/octet-stream", blobData)
			return
		}
	}

	// JSON 响应 - 直接返回上游格式
	c.JSON(http.StatusOK, result)
}

// TuboshuRouteFormData 处理 tuboshu_route_formdata 请求（文件上传代理）
func TuboshuRouteFormData(c *gin.Context) {
	path := c.PostForm("path")
	method := c.PostForm("method")
	if path == "" {
		response.BadRequest(c, "path不能为空")
		return
	}
	if method == "" {
		method = "POST"
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败: "+err.Error())
		return
	}
	defer file.Close()

	result, err := tuboshuSvc.HandleFormDataRoute(path, method, file, fileHeader)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TuboshuOrderList 获取论文订单列表（独立路由，供后台对接中心使用）
func TuboshuOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade, _ := c.Get("grade")
	isAdmin := grade == "2" || grade == "3"

	params := map[string]interface{}{
		"page": c.DefaultQuery("page", "1"),
		"size": c.DefaultQuery("size", "10"),
	}

	result, _, err := tuboshuSvc.HandleRoute(uid, isAdmin, service.TuboshuRouteRequest{
		Method: "GET",
		Path:   "/task/list",
		Params: params,
	}, c.ClientIP())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 将结果序列化再反序列化以便使用 response 包
	if resultMap, ok := result.(map[string]interface{}); ok {
		if data, ok := resultMap["data"]; ok {
			response.Success(c, data)
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

// TuboshuSavePriceConfig 保存价格配置（管理员）
func TuboshuSavePriceConfig(c *gin.Context) {
	var priceConfig map[string]interface{}
	raw, err := c.GetRawData()
	if err != nil {
		response.BadRequest(c, "读取请求体失败")
		return
	}

	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		response.BadRequest(c, "参数格式错误")
		return
	}

	if pc, ok := body["priceConfig"].(map[string]interface{}); ok {
		priceConfig = pc
	} else {
		priceConfig = body
	}

	if err := tuboshuSvc.SavePriceConfig(priceConfig); err != nil {
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}
