package middleware

import (
	"strings"
	"sync"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var lastTimeCache sync.Map

type Claims struct {
	UID   int    `json:"uid"`
	User  string `json:"user"`
	Grade string `json:"grade"`
	jwt.RegisteredClaims
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Global.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		c.Set("uid", claims.UID)
		c.Set("username", claims.User)
		c.Set("grade", claims.Grade)
		// 限制同一用户每分钟最多更新一次 lasttime
		if v, ok := lastTimeCache.Load(claims.UID); !ok || time.Since(v.(time.Time)) >= time.Minute {
			lastTimeCache.Store(claims.UID, time.Now())
			go func(uid int) {
				database.DB.Exec("UPDATE qingka_wangke_user SET lasttime = NOW() WHERE uid = ?", uid)
			}(claims.UID)
		}
		c.Next()
	}
}

func WSAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Global.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		c.Set("uid", claims.UID)
		c.Set("username", claims.User)
		c.Set("grade", claims.Grade)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		grade, exists := c.Get("grade")
		if !exists || (grade.(string) != "2" && grade.(string) != "3") {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
