package routes

import (
	legacymodule "go-api/internal/legacy/module"
	legacyopenapi "go-api/internal/legacy/openapi"
	legacyphp "go-api/internal/legacy/php"

	"github.com/gin-gonic/gin"
)

func registerLegacyRoutes(r *gin.Engine, api *gin.RouterGroup) {
	legacyopenapi.RegisterCompatRoutes(r)
	legacyopenapi.RegisterRoutes(r)
	legacyphp.RegisterRoutes(r, api)
	legacymodule.RegisterRoutes(api)
}
