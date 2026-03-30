package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	obslogger "go-api/internal/observability/logger"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// demoMode 演示站模式管理器
// 使用文件标记（.demo_mode）控制，Go 运行时定期检查，无需重启
var demoMode struct {
	sync.RWMutex
	enabled    bool
	lastCheck  time.Time
	flagFile   string
	checkEvery time.Duration
}

func init() {
	demoMode.flagFile = ".demo_mode"
	demoMode.checkEvery = 3 * time.Second
}

// IsDemoMode 返回当前是否处于演示模式
func IsDemoMode() bool {
	demoMode.RLock()
	if time.Since(demoMode.lastCheck) < demoMode.checkEvery {
		v := demoMode.enabled
		demoMode.RUnlock()
		return v
	}
	demoMode.RUnlock()

	// 需要刷新
	demoMode.Lock()
	defer demoMode.Unlock()
	// double-check
	if time.Since(demoMode.lastCheck) < demoMode.checkEvery {
		return demoMode.enabled
	}
	_, err := os.Stat(demoMode.flagFile)
	demoMode.enabled = err == nil
	demoMode.lastCheck = time.Now()
	return demoMode.enabled
}

// SetDemoMode 设置演示模式开关
func SetDemoMode(enable bool) error {
	demoMode.Lock()
	defer demoMode.Unlock()

	if enable {
		f, err := os.Create(demoMode.flagFile)
		if err != nil {
			return err
		}
		f.WriteString("demo")
		f.Close()
		obslogger.L().Info("DemoMode 演示模式已开启")
	} else {
		os.Remove(demoMode.flagFile)
		obslogger.L().Info("DemoMode 演示模式已关闭")
	}
	demoMode.enabled = enable
	demoMode.lastCheck = time.Now()
	return nil
}

// demoWhitelist 演示模式下允许通过的路径关键词（即使是写操作也放行）
var demoWhitelist = []string{
	"/api/v1/auth/login",
	"/api/v1/auth/refresh-token",
	"/api/v1/auth/logout",
	"/api/v1/query",
	"/api/v1/admin/demo-mode",
}

func isDemoWhitelisted(path string) bool {
	for _, w := range demoWhitelist {
		if strings.HasPrefix(path, w) {
			return true
		}
	}
	return false
}

// DemoGuard 演示站守卫中间件
// 开启演示模式后，禁止所有写操作（POST/PUT/DELETE），仅允许 GET 和白名单路径
func DemoGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsDemoMode() {
			c.Next()
			return
		}

		method := c.Request.Method

		// GET / HEAD / OPTIONS 始终放行
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		// 白名单路径放行
		if isDemoWhitelisted(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 阻止写操作
		response.BusinessError(c, 4033, "当前为演示站，禁止修改数据")
		c.Abort()
	}
}
