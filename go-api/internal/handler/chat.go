package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var chatService = service.NewChatService()

func ChatSessions(c *gin.Context) {
	uid := c.GetInt("uid")
	sessions, err := chatService.Sessions(uid)
	if err != nil {
		response.ServerError(c, "查询会话失败")
		return
	}
	response.Success(c, sessions)
}

func ChatMessages(c *gin.Context) {
	uid := c.GetInt("uid")
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}

	var req model.ChatMessagesRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 50
	}

	msgs, err := chatService.Messages(uid, listID, req.Limit)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, msgs)
}

func ChatHistory(c *gin.Context) {
	uid := c.GetInt("uid")
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}

	var req model.ChatHistoryRequest
	_ = c.ShouldBindQuery(&req)

	msgs, err := chatService.History(uid, listID, req.BeforeID, req.Limit)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, msgs)
}

func ChatNew(c *gin.Context) {
	uid := c.GetInt("uid")
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}

	var req model.ChatNewRequest
	_ = c.ShouldBindQuery(&req)

	msgs, err := chatService.NewMessages(uid, listID, req.AfterID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, msgs)
}

func ChatSend(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写消息内容")
		return
	}

	msg, err := chatService.Send(uid, req)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, msg)
}

func ChatSendImage(c *gin.Context) {
	uid := c.GetInt("uid")

	var req model.ChatSendImageRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "请填写完整参数")
		return
	}

	file, err := c.FormFile("img")
	if err != nil {
		response.BadRequest(c, "图片上传失败")
		return
	}

	const maxSize = 5 * 1024 * 1024
	if file.Size <= 0 || file.Size > maxSize {
		response.BadRequest(c, "文件大小超过限制（最大5MB）")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".bmp": true,
	}
	if !allowedExt[ext] {
		response.BadRequest(c, "不支持的文件格式")
		return
	}

	uploadDir := filepath.Join("uploads", "chat")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		response.ServerError(c, "创建上传目录失败")
		return
	}

	filename := fmt.Sprintf("%s_%d%s", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000, ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.ServerError(c, "图片保存失败")
		return
	}

	imgURL := "/uploads/chat/" + filename
	msg, err := chatService.SendImage(uid, req, imgURL)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, msg)
}

func ChatMarkRead(c *gin.Context) {
	uid := c.GetInt("uid")
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}

	if err := chatService.MarkRead(uid, listID); err != nil {
		response.ServerError(c, "标记已读失败")
		return
	}

	response.SuccessMsg(c, "已标记已读")
}

func ChatUnread(c *gin.Context) {
	uid := c.GetInt("uid")
	total, err := chatService.UnreadTotal(uid)
	if err != nil {
		response.ServerError(c, "查询未读数失败")
		return
	}

	response.Success(c, gin.H{"unread": total})
}

func ChatCreate(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChatCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请指定聊天对象")
		return
	}

	listID, err := chatService.CreateChat(uid, req.TargetUID)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}

	response.Success(c, gin.H{"list_id": listID})
}

func AdminChatSessions(c *gin.Context) {
	sessions, err := chatService.AdminSessions()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, sessions)
}

func AdminChatMessages(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}
	var req model.ChatMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Limit = 50
	}
	rows, err := chatService.AdminMessages(listID, req.Limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, rows)
}

func AdminChatStats(c *gin.Context) {
	stats, err := chatService.ChatStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func AdminChatCleanup(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days < 1 {
		req.Days = 14
	}
	archived, err := chatService.ManualCleanup(req.Days)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	trimmed, _ := chatService.TrimSessionMessages()
	response.Success(c, gin.H{
		"archived": archived,
		"trimmed":  trimmed,
	})
}
