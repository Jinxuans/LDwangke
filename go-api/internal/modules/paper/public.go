package paper

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func PaperPriceInfo(c *gin.Context) {
	uid := c.GetInt("uid")
	info := papers.GetPriceInfo(uid)
	response.Success(c, info)
}

func PaperGenerateTitles(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := papers.GenerateTitles(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperGenerateOutline(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := papers.GenerateOutline(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperOutlineStatus(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		response.BadRequest(c, "缺少orderId")
		return
	}
	result, err := papers.OutlineStatus(orderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperOrderSubmit(c *gin.Context) {
	uid := c.GetInt("uid")
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := papers.PaperOrderSubmit(uid, params)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

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

	result, err := papers.GetOrderList(uid, isAdmin, page, pageSize, searchParams)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperDownload(c *gin.Context) {
	orderID := c.Query("orderId")
	fileName := c.Query("fileName")
	if orderID == "" {
		response.BadRequest(c, "缺少orderId")
		return
	}
	result, err := papers.PaperDownload(orderID, fileName)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperTextRewrite(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&params); err != nil || params.Content == "" {
		response.BadRequest(c, "请输入内容")
		return
	}
	err := papers.TextRewriteSubmit(uid, params.Content, c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

func PaperTextRewriteAIGC(c *gin.Context) {
	uid := c.GetInt("uid")
	var params struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&params); err != nil || params.Content == "" {
		response.BadRequest(c, "请输入内容")
		return
	}
	err := papers.TextRewriteAIGCSubmit(uid, params.Content, c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

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
	err := papers.PaperParaEditSubmit(uid, params.Content, params.Yijian, c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
	}
}

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

	result, err := papers.FileDedupSubmit(uid, file, header, wordCount, aigc, jiangchong)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperCountWords(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败")
		return
	}
	defer file.Close()

	result, err := papers.CountWords(file, header)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperUploadCover(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败")
		return
	}
	defer file.Close()

	result, err := papers.UploadCover(file, header)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperGetTemplates(c *gin.Context) {
	params := map[string]string{}
	for _, key := range []string{"pageNum", "pageSize", "name", "isPublic"} {
		if v := c.Query(key); v != "" {
			params[key] = v
		}
	}
	result, err := papers.GetTemplateList(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func PaperSaveTemplate(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := papers.SaveTemplate(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

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
	result, err := papers.GenerateTaskWithFee(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

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
	result, err := papers.GenerateProposalWithFee(uid, orderID)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
