package routes

import (
	legacymodule "go-api/internal/legacy/module"
	legacyopenapi "go-api/internal/legacy/openapi"

	"github.com/gin-gonic/gin"
)

func registerLegacyRoutes(r *gin.Engine, api *gin.RouterGroup) {
	legacyopenapi.RegisterCompatRoutes(r)
	legacyopenapi.RegisterRoutes(r)
	legacymodule.RegisterRoutes(api)
}
