package middleware

import (
	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth 外部API密钥认证中间件
// PHP: 通过 uid + key 参数验证身份（对应 apisub.php 的密钥校验逻辑）
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetHeader("X-API-UID")
		key := c.GetHeader("X-API-Key")
		if uid == "" || key == "" {
			uid = c.Query("uid")
			key = c.Query("key")
		}
		if uid == "" || key == "" {
			// 也支持 POST form
			uid = c.PostForm("uid")
			key = c.PostForm("key")
		}
		if uid == "" || key == "" {
			response.BusinessError(c, 0, "缺少 uid 或 key 参数")
			c.Abort()
			return
		}

		var dbUID int
		var dbKey, dbGrade, dbUser string
		var dbMoney, dbAddPrice float64
		err := database.DB.QueryRow(
			"SELECT uid, COALESCE(`key`,''), COALESCE(grade,'0'), COALESCE(user,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?",
			uid,
		).Scan(&dbUID, &dbKey, &dbGrade, &dbUser, &dbMoney, &dbAddPrice)
		if err != nil {
			response.BusinessError(c, 0, "用户不存在")
			c.Abort()
			return
		}

		if dbKey == "" || dbKey == "0" {
			response.BusinessError(c, 0, "未开通API密钥")
			c.Abort()
			return
		}

		if dbKey != key {
			response.BusinessError(c, 0, "密钥错误")
			c.Abort()
			return
		}

		c.Set("uid", dbUID)
		c.Set("username", dbUser)
		c.Set("grade", dbGrade)
		c.Set("money", dbMoney)
		c.Set("addprice", dbAddPrice)
		c.Next()
	}
}
