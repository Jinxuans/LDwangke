package module

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
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context) {
	appID := c.Param("app_id")
	act := c.Query("act")
	if appID == "" || act == "" {
		response.BadRequest(c, "缺少参数")
		return
	}

	mod, err := getByAppID(appID)
	if err != nil {
		response.BusinessError(c, 1001, "模块不存在或未启用")
		return
	}

	phpBackend := "http://127.0.0.1"
	if cfg := config.Global; cfg != nil && cfg.Server.PhpBackend != "" {
		phpBackend = cfg.Server.PhpBackend
	}

	apiBase := mod.ApiBase
	if apiBase == "" {
		apiBase = "/jingyu/api.php"
	}

	targetQuery := url.Values{}
	targetQuery.Set("appId", appID)
	for key, vals := range c.Request.URL.Query() {
		if key == "app_id" {
			continue
		}
		for _, v := range vals {
			targetQuery.Set(key, v)
		}
	}
	targetURL := phpBackend + apiBase + "?" + targetQuery.Encode()

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		response.ServerErrorf(c, err, "读取请求体失败")
		return
	}

	client := &http.Client{Timeout: 180 * time.Second}
	var proxyReq *http.Request

	if c.Request.Method == "GET" {
		proxyReq, err = http.NewRequest("GET", targetURL, nil)
	} else {
		proxyReq, err = http.NewRequest("POST", targetURL, bytes.NewReader(body))
	}
	if err != nil {
		response.ServerErrorf(c, err, "创建代理请求失败")
		return
	}

	proxyReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))
	if auth := c.GetHeader("Authorization"); auth != "" {
		proxyReq.Header.Set("Authorization", auth)
	}
	uid := c.GetInt("uid")
	proxyReq.Header.Set("X-User-UID", strconv.Itoa(uid))

	for _, cookie := range c.Request.Cookies() {
		proxyReq.AddCookie(cookie)
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		response.ServerErrorf(c, err, "代理请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

func FrameURL(c *gin.Context) {
	appID := c.Param("app_id")
	if appID == "" {
		response.BadRequest(c, "缺少 app_id")
		return
	}

	if config.Global == nil || config.Global.Server.PhpPublicURL == "" {
		response.ServerError(c, "PHP 公网地址未配置")
		return
	}
	if config.Global.Server.BridgeSecret == "" {
		response.ServerError(c, "bridge_secret 未配置")
		return
	}

	mod, err := getByAppID(appID)
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

	ts := time.Now().Unix()
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d%d%s", uid, ts, config.Global.Server.BridgeSecret))))

	frameURL := fmt.Sprintf("%s/auth_bridge.php?uid=%d&ts=%d&sign=%s&target=%s",
		config.Global.Server.PhpPublicURL, uid, ts, sign, url.QueryEscape(mod.ViewURL))

	response.Success(c, map[string]string{"frame_url": frameURL})
}
