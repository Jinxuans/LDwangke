package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var mailService = service.NewMailService()

func MailList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var req model.MailListRequest
	_ = c.ShouldBindQuery(&req)
	if req.Type == "" {
		req.Type = "inbox"
	}

	list, total, err := mailService.List(uid, grade, req)
	if err != nil {
		response.ServerError(c, "查询站内信失败")
		return
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"page":  req.Page,
			"limit": req.Limit,
			"total": total,
		},
	})
}

func MailDetail(c *gin.Context) {
	uid := c.GetInt("uid")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的邮件 ID")
		return
	}

	mail, err := mailService.Detail(uid, id)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, mail)
}

func MailSend(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var req model.MailSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写标题和收件人")
		return
	}

	id, err := mailService.Send(uid, grade, req)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{"id": id})
}

func MailUnread(c *gin.Context) {
	uid := c.GetInt("uid")
	count, err := mailService.UnreadCount(uid)
	if err != nil {
		response.ServerError(c, "查询未读数失败")
		return
	}
	response.Success(c, gin.H{"unread": count})
}

func MailDelete(c *gin.Context) {
	uid := c.GetInt("uid")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的邮件 ID")
		return
	}

	if err := mailService.Delete(uid, id); err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}

	response.SuccessMsg(c, "删除成功")
}

func MailUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	// 限制 20MB
	if file.Size > 20*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 20MB")
		return
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	newName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), c.GetInt("uid"), ext)

	// 保存目录
	uploadDir := "uploads/mail"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.ServerError(c, "创建上传目录失败")
		return
	}

	savePath := filepath.Join(uploadDir, newName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.ServerError(c, "保存文件失败")
		return
	}

	response.Success(c, gin.H{
		"file_url":  "/" + savePath,
		"file_name": file.Filename,
	})
}
