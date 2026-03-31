package paper

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册论文平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	paper := api.Group("/paper")
	{
		paper.GET("/prices", PaperPriceInfo)
		paper.POST("/generate-titles", PaperGenerateTitles)
		paper.POST("/generate-outline", PaperGenerateOutline)
		paper.GET("/outline-status", PaperOutlineStatus)
		paper.POST("/order", PaperOrderSubmit)
		paper.GET("/orders", PaperOrderList)
		paper.GET("/download", PaperDownload)
		paper.POST("/text-rewrite", PaperTextRewrite)
		paper.POST("/text-rewrite-aigc", PaperTextRewriteAIGC)
		paper.POST("/para-edit", PaperParaEdit)
		paper.POST("/file-dedup", PaperFileDedupSubmit)
		paper.POST("/count-words", PaperCountWords)
		paper.POST("/upload-cover", PaperUploadCover)
		paper.GET("/templates", PaperGetTemplates)
		paper.POST("/template", PaperSaveTemplate)
		paper.POST("/generate-task", PaperGenerateTaskWithFee)
		paper.POST("/generate-proposal", PaperGenerateProposalWithFee)
	}
}
