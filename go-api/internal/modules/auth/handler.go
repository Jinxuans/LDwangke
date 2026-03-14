package auth

import (
	"fmt"
	"net/http"

	"go-api/internal/config"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var authService = NewService()

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入用户名和密码")
		return
	}

	result, refreshToken, err := authService.Login(req)
	if err != nil {
		if err.Error() == "NEED_ADMIN_AUTH" {
			response.BusinessError(c, 5, "需要管理员二次验证")
			return
		}
		response.BusinessError(c, 1001, err.Error())
		return
	}

	SetRefreshTokenCookie(c.Writer, refreshToken, config.Global.JWT.RefreshTTL)

	uid := 0
	fmt.Sscanf(result.UserId, "%d", &uid)
	go authService.CheckLoginDevice(
		uid,
		req.Username,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	response.Success(c, result)
}

func SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	switch req.Purpose {
	case "register":
		if err := authService.SendRegisterCode(req.Email); err != nil {
			response.BusinessError(c, 1003, err.Error())
			return
		}
	default:
		response.BadRequest(c, "无效的验证码类型")
		return
	}

	response.SuccessMsg(c, "验证码已发送")
}

func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	if err := authService.SendResetCode(req.Email); err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}

	response.SuccessMsg(c, "重置密码验证码已发送")
}

func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := authService.ResetPassword(req.Email, req.Code, req.Password); err != nil {
		response.BusinessError(c, 1005, err.Error())
		return
	}

	response.SuccessMsg(c, "密码重置成功")
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请确保所有信息不能为空！")
		return
	}

	if err := authService.Register(req); err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.SuccessMsg(c, "注册成功")
}

func RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		response.Forbidden(c, "缺少 refresh_token")
		return
	}

	newAccessToken, newRefreshToken, err := authService.RefreshAccessToken(refreshToken)
	if err != nil {
		response.Forbidden(c, err.Error())
		return
	}

	SetRefreshTokenCookie(c.Writer, newRefreshToken, config.Global.JWT.RefreshTTL)
	c.String(http.StatusOK, newAccessToken)
}

func Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	response.SuccessMsg(c, "登出成功")
}

func UserInfo(c *gin.Context) {
	uid := c.GetInt("uid")
	info, err := authService.GetUserInfo(uid)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, info)
}

func AccessCodes(c *gin.Context) {
	grade := c.GetString("grade")
	codes := authService.GetAccessCodes(grade)
	response.Success(c, codes)
}

func Impersonate(c *gin.Context) {
	var body struct {
		UID int `json:"uid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.UID <= 0 {
		response.BadRequest(c, "请指定目标用户")
		return
	}

	result, refreshToken, err := authService.Impersonate(body.UID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	SetRefreshTokenCookie(c.Writer, refreshToken, config.Global.JWT.RefreshTTL)
	response.Success(c, result)
}
