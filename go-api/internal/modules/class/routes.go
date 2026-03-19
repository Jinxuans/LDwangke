package class

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册课程域路由。
func RegisterRoutes(api *gin.RouterGroup) {
	class := api.Group("/class")
	{
		class.GET("/list", List)
		class.GET("/list-paged", ListPaged)
		class.GET("/search", Search)
		class.POST("/search", QueryCourse)
		class.GET("/categories", Categories)
		class.GET("/category-switches", CategorySwitches)
	}
}
