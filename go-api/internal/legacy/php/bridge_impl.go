package php

import (
	"crypto/md5"
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func requireBridgeSecret(c *gin.Context) (string, bool) {
	if config.Global == nil || strings.TrimSpace(config.Global.Server.BridgeSecret) == "" {
		response.ServerError(c, "bridge_secret 未配置")
		return "", false
	}
	return config.Global.Server.BridgeSecret, true
}

// ===== PHP 反向代理 =====

// PhpProxy 将 /php-api/* 请求反向代理到 PHP 后端
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		backend := config.Global.Server.PhpBackend
		if backend == "" {
			response.BusinessError(c, 1001, "PHP 后端未配置")
			return
		}

		target, err := url.Parse(backend)
		if err != nil {
			response.BusinessError(c, 1002, "PHP 后端地址无效")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			response.BusinessError(c, 1003, "PHP 后端连接失败: "+err.Error())
		}

		// 去掉 /php-api 前缀，PHP 内置服务器的 document root 是 public/
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/php-api")
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// ===== PHP 桥接内部 API（bridge_secret 认证） =====

// verifyBridgeSign 验证桥接签名：sign = md5(uid + ts + bridge_secret)
func verifyBridgeSign(c *gin.Context) bool {
	sign := c.PostForm("sign")
	if sign == "" {
		sign = c.Query("sign")
	}
	ts := c.PostForm("ts")
	if ts == "" {
		ts = c.Query("ts")
	}
	uid := c.PostForm("uid")
	if uid == "" {
		uid = c.Query("uid")
	}

	if sign == "" || ts == "" || uid == "" {
		response.BusinessError(c, 401, "缺少签名参数")
		return false
	}

	// 时间戳校验（±300秒）
	tsInt, _ := strconv.ParseInt(ts, 10, 64)
	if math.Abs(float64(time.Now().Unix()-tsInt)) > 300 {
		response.BusinessError(c, 401, "签名已过期")
		return false
	}

	secret, ok := requireBridgeSecret(c)
	if !ok {
		return false
	}
	expected := fmt.Sprintf("%x", md5.Sum([]byte(uid+ts+secret)))
	if sign != expected {
		response.BusinessError(c, 401, "签名验证失败")
		return false
	}

	return true
}

// BridgeMoneyChange PHP 通知 Go：用户余额变动
// POST /internal/php-bridge/money
// 参数: uid, amount(正=充值 负=扣费), reason, ts, sign
func BridgeMoneyChange(c *gin.Context) {
	if !verifyBridgeSign(c) {
		return
	}

	uid, _ := strconv.Atoi(c.PostForm("uid"))
	amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64)
	reason := c.PostForm("reason")

	if uid <= 0 {
		response.BadRequest(c, "uid 无效")
		return
	}
	if amount == 0 {
		response.BadRequest(c, "amount 不能为 0")
		return
	}
	if reason == "" {
		reason = "PHP 桥接操作"
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 开启事务
	tx, err := database.DB.Begin()
	if err != nil {
		response.BusinessError(c, 1003, "系统繁忙")
		return
	}
	defer tx.Rollback()

	// 锁定用户行
	var currentMoney float64
	err = tx.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ? FOR UPDATE", uid).Scan(&currentMoney)
	if err != nil {
		response.BusinessError(c, 1001, "用户不存在")
		return
	}

	// 扣费时检查余额
	if amount < 0 && currentMoney < -amount {
		response.BusinessError(c, 1002, fmt.Sprintf("余额不足，当前 %.2f，需扣 %.2f", currentMoney, -amount))
		return
	}

	// 更新余额
	_, err = tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", amount, uid)
	if err != nil {
		response.BusinessError(c, 1003, "更新余额失败")
		return
	}

	// 记录流水
	logType := "扣费"
	if amount > 0 {
		logType = "充值"
	}
	tx.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, ?, ?, ?, ?, ?)",
		uid, logType, amount, currentMoney+amount, reason, now,
	)

	if err := tx.Commit(); err != nil {
		response.BusinessError(c, 1003, "提交失败")
		return
	}

	obslogger.L().Info("PHP-Bridge 余额变动", "uid", uid, "amount", amount, "reason", reason)

	response.Success(c, gin.H{
		"uid":     uid,
		"amount":  amount,
		"balance": currentMoney + amount,
	})
}

// BridgeGetUser PHP 获取用户信息
// GET /internal/php-bridge/user?uid=X&ts=T&sign=S
func BridgeGetUser(c *gin.Context) {
	if !verifyBridgeSign(c) {
		return
	}

	uid, _ := strconv.Atoi(c.Query("uid"))
	if uid <= 0 {
		response.BadRequest(c, "uid 无效")
		return
	}

	var user, money, grade string
	err := database.DB.QueryRow(
		"SELECT COALESCE(user,''), COALESCE(money,'0'), COALESCE(grade,'0') FROM qingka_wangke_user WHERE uid = ?", uid,
	).Scan(&user, &money, &grade)
	if err != nil {
		response.BusinessError(c, 1001, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"uid":   uid,
		"user":  user,
		"money": money,
		"grade": grade,
	})
}

// BridgeCreateOrder PHP 通知 Go 创建订单（由 PHP 业务创建后通知，或直接由 Go 创建）
// POST /internal/php-bridge/order
// 参数: uid, cid, school, user, pass, kcname, kcid, fees, ts, sign
func BridgeCreateOrder(c *gin.Context) {
	if !verifyBridgeSign(c) {
		return
	}

	uid, _ := strconv.Atoi(c.PostForm("uid"))
	cid, _ := strconv.Atoi(c.PostForm("cid"))
	school := c.PostForm("school")
	userName := c.PostForm("user")
	pass := c.PostForm("pass")
	kcname := c.PostForm("kcname")
	kcid := c.PostForm("kcid")
	fees := c.PostForm("fees")
	remarkStr := c.PostForm("remark")

	if uid <= 0 || cid <= 0 {
		response.BadRequest(c, "uid 和 cid 不能为空")
		return
	}

	// 获取课程信息
	var clsName, noun, docking string
	err := database.DB.QueryRow(
		"SELECT COALESCE(name,''), COALESCE(noun,''), COALESCE(docking,'0') FROM qingka_wangke_class WHERE cid = ?", cid,
	).Scan(&clsName, &noun, &docking)
	if err != nil {
		response.BusinessError(c, 1001, "课程不存在")
		return
	}

	dockingID, _ := strconv.Atoi(docking)
	dockStatus := 99 // 默认无对接
	if dockingID > 0 {
		dockStatus = 0 // 待对接
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	if fees == "" {
		fees = "0"
	}

	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_order (uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, fees, noun, addtime, dockstatus, remarks) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uid, cid, dockingID, clsName, school, "", userName, pass, kcid, kcname, fees, noun, now, dockStatus, remarkStr,
	)
	if err != nil {
		response.BusinessError(c, 1003, "创建订单失败: "+err.Error())
		return
	}

	oid, _ := result.LastInsertId()

	obslogger.L().Info("PHP-Bridge 创建订单", "oid", oid, "uid", uid, "cid", cid, "user", userName, "kcname", kcname)

	response.Success(c, gin.H{
		"oid": oid,
	})
}

// BridgeAuthURL 生成 PHP 认证桥接 URL（供前端获取，用于 iframe 加载 PHP 页面）
// GET /internal/php-bridge/auth-url?uid=X&target=/path/to/page.php
func BridgeAuthURL(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid <= 0 {
		response.BadRequest(c, "未登录")
		return
	}

	target := c.Query("target")
	if target == "" {
		response.BadRequest(c, "target 不能为空")
		return
	}

	// 安全检查
	if strings.Contains(target, "//") || strings.Contains(target, "..") {
		response.BadRequest(c, "非法路径")
		return
	}

	ts := fmt.Sprintf("%d", time.Now().Unix())
	secret, ok := requireBridgeSecret(c)
	if !ok {
		return
	}
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", uid)+ts+secret)))

	// 构建桥接 URL
	phpURL := config.Global.Server.PhpPublicURL
	if phpURL == "" {
		phpURL = config.Global.Server.PhpBackend
	}
	if strings.TrimSpace(phpURL) == "" {
		response.ServerError(c, "PHP 公网地址未配置")
		return
	}

	bridgeURL := fmt.Sprintf("%s/auth_bridge.php?uid=%d&ts=%s&sign=%s&target=%s",
		strings.TrimRight(phpURL, "/"), uid, ts, sign, url.QueryEscape(target))

	response.Success(c, gin.H{
		"url": bridgeURL,
	})
}
