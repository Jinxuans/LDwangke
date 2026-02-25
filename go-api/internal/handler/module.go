package handler

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-api/internal/config"
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var moduleService = service.NewModuleService()

func ModuleList(c *gin.Context) {
	moduleType := c.Query("type")
	var modules []model.DynamicModule
	var err error
	if moduleType != "" {
		modules, err = moduleService.ListActiveByType(moduleType)
	} else {
		modules, err = moduleService.ListActive()
	}
	if err != nil {
		response.ServerError(c, "查询模块失败")
		return
	}
	response.Success(c, modules)
}

func ModuleListAll(c *gin.Context) {
	modules, err := moduleService.ListAll()
	if err != nil {
		response.ServerError(c, "查询模块失败")
		return
	}
	response.Success(c, modules)
}

func ModuleProxy(c *gin.Context) {
	appID := c.Param("app_id")
	act := c.Query("act")
	if appID == "" || act == "" {
		response.BadRequest(c, "缺少参数")
		return
	}

	mod, err := moduleService.GetByAppID(appID)
	if err != nil {
		response.BusinessError(c, 1001, "模块不存在或未启用")
		return
	}

	// PHP 后端基础地址（从配置读取，默认 http://127.0.0.1）
	phpBackend := "http://127.0.0.1"
	if cfg := config.Global; cfg != nil && cfg.Server.PhpBackend != "" {
		phpBackend = cfg.Server.PhpBackend
	}

	// 构造目标 URL
	apiBase := mod.ApiBase
	if apiBase == "" {
		apiBase = "/jingyu/api.php"
	}
	// 透传所有 query 参数到 PHP 后端
	targetQuery := url.Values{}
	targetQuery.Set("appId", appID)
	for key, vals := range c.Request.URL.Query() {
		if key == "app_id" {
			continue // app_id 已映射为 appId
		}
		for _, v := range vals {
			targetQuery.Set(key, v)
		}
	}
	targetURL := phpBackend + apiBase + "?" + targetQuery.Encode()

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		response.ServerError(c, "读取请求体失败")
		return
	}

	// 转发请求到 PHP 后端
	client := &http.Client{Timeout: 180 * time.Second}
	var proxyReq *http.Request

	if c.Request.Method == "GET" {
		proxyReq, err = http.NewRequest("GET", targetURL, nil)
	} else {
		proxyReq, err = http.NewRequest("POST", targetURL, bytes.NewReader(body))
	}
	if err != nil {
		response.ServerError(c, "创建代理请求失败")
		return
	}

	// 透传认证信息
	proxyReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))
	if auth := c.GetHeader("Authorization"); auth != "" {
		proxyReq.Header.Set("Authorization", auth)
	}
	uid := c.GetInt("uid")
	proxyReq.Header.Set("X-User-UID", strconv.Itoa(uid))

	// 转发 cookie
	for _, cookie := range c.Request.Cookies() {
		proxyReq.AddCookie(cookie)
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		response.ServerError(c, "代理请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 直接返回 PHP 后端的响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ModuleFrameURL 生成 iframe 认证桥接签名 URL
func ModuleFrameURL(c *gin.Context) {
	appID := c.Param("app_id")
	if appID == "" {
		response.BadRequest(c, "缺少 app_id")
		return
	}

	mod, err := moduleService.GetByAppID(appID)
	if err != nil {
		response.BusinessError(c, 1001, "模块不存在或未启用")
		return
	}

	if mod.ViewURL == "" {
		response.BusinessError(c, 1002, "该模块未配置前端页面")
		return
	}

	uid := c.GetInt("uid")
	if uid <= 0 {
		response.BadRequest(c, "未获取到用户信息")
		return
	}

	// 读取配置
	phpPublicURL := ""
	bridgeSecret := "qingka_bridge_secret_2024"
	if cfg := config.Global; cfg != nil {
		phpPublicURL = cfg.Server.PhpPublicURL
		if cfg.Server.BridgeSecret != "" {
			bridgeSecret = cfg.Server.BridgeSecret
		}
	}

	// 生成签名
	ts := time.Now().Unix()
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d%d%s", uid, ts, bridgeSecret))))

	// 构造 iframe URL（phpPublicURL 为空时用相对路径，同域通过 Vite/nginx 代理）
	frameURL := fmt.Sprintf("%s/auth_bridge.php?uid=%d&ts=%d&sign=%s&target=%s",
		phpPublicURL, uid, ts, sign, url.QueryEscape(mod.ViewURL))

	response.Success(c, map[string]string{"frame_url": frameURL})
}

func AdminModuleSave(c *gin.Context) {
	var req model.ModuleSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := moduleService.Save(req); err != nil {
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func AdminModuleDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "ID无效")
		return
	}
	if err := moduleService.Delete(id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.Success(c, nil)
}
