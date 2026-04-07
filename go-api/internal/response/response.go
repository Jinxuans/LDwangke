package response

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	obslogger "go-api/internal/observability/logger"
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		RequestID: requestID(c),
		Timestamp: time.Now().Unix(),
	})
}

func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   msg,
		RequestID: requestID(c),
		Timestamp: time.Now().Unix(),
	})
}

func Error(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, Response{
		Code:      code,
		Message:   msg,
		RequestID: requestID(c),
		Timestamp: time.Now().Unix(),
	})
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, 422, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, 401, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, 403, msg)
}

func ServerError(c *gin.Context, msg string) {
	reqID := requestID(c)
	_ = c.Error(errors.New(msg)).SetType(gin.ErrorTypePrivate)
	obslogger.Request(c).Error(msg, "request_id", reqID)
	Error(c, http.StatusInternalServerError, 500, msg)
}

// ServerErrorf 将真实错误记录到日志（附带 request_id），并向客户端返回通用错误提示。
// 替代直接调用 ServerError，让排查时日志里能看到 err 原文。
//
// 用法：response.ServerErrorf(c, err, "查询统计失败")
func ServerErrorf(c *gin.Context, err error, msg string) {
	reqID := requestID(c)
	if err != nil {
		_ = c.Error(err).SetType(gin.ErrorTypePrivate)
	}
	obslogger.Request(c).Error(msg, "request_id", reqID, "error", err)
	Error(c, http.StatusInternalServerError, 500, msg)
}

func BusinessError(c *gin.Context, code int, msg string) {
	obslogger.Request(c).Warn("business error", "code", code, "message", msg)
	Error(c, http.StatusOK, code, msg)
}

type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

func SuccessPage(c *gin.Context, list interface{}, total int64, page, size int) {
	Success(c, PageData{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func requestID(c *gin.Context) string {
	reqID, _ := c.Get("request_id")
	if reqID == nil {
		return "-"
	}
	return fmt.Sprint(reqID)
}
