package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var paperSvc = service.NewPaperService()

// PaperEnsureTable 启动时建表
func PaperEnsureTable() {
	paperSvc.EnsureTable()
}

// ==================== 用户端 ====================

// PaperPriceInfo 获取价格信息
func PaperPriceInfo(c *gin.Context) {
	uid := c.GetInt("uid")
	info := paperSvc.GetPriceInfo(uid)
	response.Success(c, info)
}

// PaperGenerateTitles 生成论文标题
func PaperGenerateTitles(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := paperSvc.GenerateTitles(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperGenerateOutline 生成论文大纲
func PaperGenerateOutline(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := paperSvc.GenerateOutline(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperOutlineStatus 获取大纲状态
func PaperOutlineStatus(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		response.BadRequest(c, "缺少orderId")
		return
	}
	result, err := paperSvc.OutlineStatus(orderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperOrderSubmit 论文下单
func PaperOrderSubmit(c *gin.Context) {
	uid := c.GetInt("uid")
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := paperSvc.PaperOrderSubmit(uid, params)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperOrderList 论文订单列表
func PaperOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade, _ := c.Get("grade")
	isAdmin := grade == "2" || grade == "3"

	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	searchParams := map[string]string{}
	if v := c.Query("title"); v != "" {
		searchParams["title"] = v
	}
	if v := c.Query("shopname"); v != "" {
		searchParams["shopname"] = v
	}
	if v := c.Query("studentName"); v != "" {
		searchParams["studentName"] = v
	}
	if v := c.Query("state"); v != "" {
		searchParams["state"] = v
	}

	result, err := paperSvc.GetOrderList(uid, isAdmin, page, pageSize, searchParams)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperDownload 论文下载
func PaperDownload(c *gin.Context) {
	orderID := c.Query("orderId")
	fileName := c.Query("fileName")
	if orderID == "" {
		response.BadRequest(c, "缺少orderId")
		return
	}
	result, err := paperSvc.PaperDownload(orderID, fileName)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperTextRewrite 文本降重（SSE流式）
func PaperTextRewrite(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&params); err != nil || params.Content == "" {
		response.BadRequest(c, "请输入内容")
		return
	}
	err := paperSvc.TextRewriteSubmit(uid, params.Content, c.Writer)
	if err != nil {
		// 如果还没开始写SSE头，返回JSON错误
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

// PaperTextRewriteAIGC 降低AIGC率（SSE流式）
func PaperTextRewriteAIGC(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&params); err != nil || params.Content == "" {
		response.BadRequest(c, "请输入内容")
		return
	}
	err := paperSvc.TextRewriteAIGCSubmit(uid, params.Content, c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

// PaperParaEdit 段落修改（SSE流式）
func PaperParaEdit(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		Content string `json:"content"`
		Yijian  string `json:"yijian"`
	}
	if err := c.ShouldBindJSON(&params); err != nil || params.Content == "" {
		response.BadRequest(c, "请输入内容")
		return
	}
	err := paperSvc.PaperParaEditSubmit(uid, params.Content, params.Yijian, c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

// PaperFileDedupSubmit 文件降重
func PaperFileDedupSubmit(c *gin.Context) {
	uid := c.GetInt("uid")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败")
		return
	}
	defer file.Close()

	wordCount, _ := strconv.Atoi(c.PostForm("wordCount"))
	aigc, _ := strconv.Atoi(c.PostForm("aigc"))
	jiangchong, _ := strconv.Atoi(c.PostForm("jiangchong"))

	result, err := paperSvc.FileDedupSubmit(uid, file, header, wordCount, aigc, jiangchong)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperCountWords 统计字数（文件上传）
func PaperCountWords(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败")
		return
	}
	defer file.Close()

	result, err := paperSvc.CountWords(file, header)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperUploadCover 上传模板文件
func PaperUploadCover(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败")
		return
	}
	defer file.Close()

	result, err := paperSvc.UploadCover(file, header)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperGetTemplates 获取模板列表
func PaperGetTemplates(c *gin.Context) {
	params := map[string]string{}
	for _, key := range []string{"pageNum", "pageSize", "name", "isPublic"} {
		if v := c.Query(key); v != "" {
			params[key] = v
		}
	}
	result, err := paperSvc.GetTemplateList(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperSaveTemplate 保存模板
func PaperSaveTemplate(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := paperSvc.SaveTemplate(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperGenerateTaskWithFee 生成任务书（扣费）
func PaperGenerateTaskWithFee(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		ID json.RawMessage `json:"id"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	orderID := strings.TrimSpace(strings.Trim(string(params.ID), `"`))
	if orderID == "" {
		response.BadRequest(c, "缺少参数id")
		return
	}
	result, err := paperSvc.GenerateTaskWithFee(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaperGenerateProposalWithFee 生成开题报告（扣费）
func PaperGenerateProposalWithFee(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		ID json.RawMessage `json:"id"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	orderID := strings.TrimSpace(strings.Trim(string(params.ID), `"`))
	if orderID == "" {
		response.BadRequest(c, "缺少参数id")
		return
	}
	result, err := paperSvc.GenerateProposalWithFee(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ==================== 管理端 ====================

// PaperConfigGet 获取论文配置（管理员）
func PaperConfigGet(c *gin.Context) {
	conf, err := paperSvc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, conf)
}

// PaperConfigSave 保存论文配置（管理员）
func PaperConfigSave(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := paperSvc.SaveConfig(data); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}
