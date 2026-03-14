package tuboshu

import (
	"net/http"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// TuboshuUserConfigGet 用户端获取土拨鼠配置（返回价格配置+页面可见性+用户倍率）
func TuboshuUserConfigGet(c *gin.Context) {
	uid := c.GetInt("uid")
	cfg, err := Tuboshu().GetConfig()
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

// TuboshuRoute 处理 tuboshu_route 请求（JSON API 代理）
func TuboshuRoute(c *gin.Context) {
	uid := c.GetInt("uid")
	grade, _ := c.Get("grade")
	isAdmin := grade == "2" || grade == "3"
	clientIP := c.ClientIP()

	var req TuboshuRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, isBlob, err := Tuboshu().HandleRoute(uid, isAdmin, req, clientIP)
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

	result, err := Tuboshu().HandleFormDataRoute(path, method, file, fileHeader)
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

	result, _, err := Tuboshu().HandleRoute(uid, isAdmin, TuboshuRouteRequest{
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
