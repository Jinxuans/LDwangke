package routes

import (
	appuimodule "go-api/internal/modules/appui"
	papermodule "go-api/internal/modules/paper"
	sdxymodule "go-api/internal/modules/sdxy"
	sxdkmodule "go-api/internal/modules/sxdk"
	tuboshumodule "go-api/internal/modules/tuboshu"
	tutuqgmodule "go-api/internal/modules/tutuqg"
	tuzhimodule "go-api/internal/modules/tuzhi"
	wmodule "go-api/internal/modules/w"
	xmmodule "go-api/internal/modules/xm"
	ydsjmodule "go-api/internal/modules/ydsj"
	yfdkmodule "go-api/internal/modules/yfdk"
	yongyemodule "go-api/internal/modules/yongye"

	"github.com/gin-gonic/gin"
)

func registerPluginRoutes(api *gin.RouterGroup) {
	// Plugin system: product-specific plugin domains outside the main web-course flow.
	tutuqgmodule.RegisterRoutes(api)
	yfdkmodule.RegisterRoutes(api)
	sxdkmodule.RegisterRoutes(api)
	xmmodule.RegisterRoutes(api)
	wmodule.RegisterRoutes(api)
	tuzhimodule.RegisterRoutes(api)
	appuimodule.RegisterRoutes(api)
	sdxymodule.RegisterRoutes(api)
	ydsjmodule.RegisterRoutes(api)
	yongyemodule.RegisterRoutes(api)
	tuboshumodule.RegisterRoutes(api)
	papermodule.RegisterRoutes(api)
}
