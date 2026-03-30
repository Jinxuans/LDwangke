package auth

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/middleware"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"
	commonmodule "go-api/internal/modules/common"
	obslogger "go-api/internal/observability/logger"
	shared "go-api/internal/shared/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service 为 auth 模块提供带 repo 依赖的服务入口。
type Service struct {
	users *shared.UserRepo
}

func NewService() *Service {
	return &Service{
		users: shared.NewUserRepo(),
	}
}

func (s *Service) Login(req LoginRequest) (*model.VbenLoginResponse, string, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, IFNULL(pass2,''), name, money, grade, active FROM qingka_wangke_user WHERE user = ?",
		req.Username,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Pass2, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil && strings.Contains(err.Error(), "pass2") {
		err = database.DB.QueryRow(
			"SELECT uid, uuid, user, pass, name, money, grade, active FROM qingka_wangke_user WHERE user = ?",
			req.Username,
		).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Name, &user.Money, &user.Grade, &user.Active)
	}
	if err == sql.ErrNoRows {
		return nil, "", errors.New("用户不存在")
	}
	if err != nil {
		return nil, "", fmt.Errorf("查询用户失败: %v", err)
	}

	if user.Active != "1" {
		return nil, "", errors.New("账号已被禁用")
	}

	if strings.HasPrefix(user.Pass, "$2a$") || strings.HasPrefix(user.Pass, "$2b$") {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(req.Password)); err != nil {
			return nil, "", errors.New("密码错误")
		}
	} else {
		if config.Global != nil && !config.Global.AllowLegacyPlaintextPasswords() {
			return nil, "", errors.New("当前系统已禁用明文密码登录，请重置密码后再试")
		}
		if user.Pass != req.Password {
			return nil, "", errors.New("密码错误")
		}
		if hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost); err == nil {
			obslogger.L().Info("Auth 用户使用明文密码登录，已自动迁移为 bcrypt", "uid", user.UID)
			database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", string(hashed), user.UID)
		}
	}

	pass2Enabled := true
	var pass2Kg string
	if err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v`='pass2_kg'").Scan(&pass2Kg); err == nil && pass2Kg == "0" {
		pass2Enabled = false
	}

	if pass2Enabled && user.Grade == "3" {
		if req.Pass2 == "" {
			return nil, "", errors.New("NEED_ADMIN_AUTH")
		}
		if user.Pass2 == "" {
		} else if strings.HasPrefix(user.Pass2, "$2a$") || strings.HasPrefix(user.Pass2, "$2b$") {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Pass2), []byte(req.Pass2)); err != nil {
				return nil, "", errors.New("管理员验证失败")
			}
		} else if user.Pass2 != req.Pass2 {
			return nil, "", errors.New("管理员验证失败")
		}
	}

	accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL, "access")
	if err != nil {
		return nil, "", errors.New("生成 Token 失败")
	}
	refreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL, "refresh")
	if err != nil {
		return nil, "", errors.New("生成 RefreshToken 失败")
	}

	_, _ = database.DB.Exec("UPDATE qingka_wangke_user SET lasttime = NOW(), endtime = NOW() WHERE uid = ?", user.UID)

	return &model.VbenLoginResponse{
		AccessToken: accessToken,
		UserId:      fmt.Sprintf("%d", user.UID),
		Username:    user.User,
		RealName:    user.User,
		Avatar:      "",
		Desc:        fmt.Sprintf("等级: %s", user.Grade),
		HomePath:    "/dashboard",
		Roles:       gradeToRoles(user.Grade),
	}, refreshToken, nil
}

func (s *Service) SendRegisterCode(email string) error {
	if err := commonmodule.CheckVerificationRateLimit("register", email); err != nil {
		return err
	}

	code, err := commonmodule.GenerateVerificationCode("register", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	subject := siteName + " - 注册验证码"
	htmlBody := templateVerifyCode(siteName, code, 10)

	if err := commonmodule.SendEmailWithType(email, subject, htmlBody, "register"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

func (s *Service) Register(req RegisterRequest) error {
	qqRe := regexp.MustCompile(`^[1-9][0-9]{4,10}$`)
	if !qqRe.MatchString(req.Username) {
		return errors.New("账号必须为有效的 QQ 号（5-11位数字，首位不为0）")
	}

	if s.adminConfigEnabled("login_email_verify") {
		if req.Email == "" {
			return errors.New("请填写邮箱地址")
		}
		if req.VerifyCode == "" {
			return errors.New("请输入邮箱验证码")
		}
		if !commonmodule.VerifyVerificationCode("register", req.Email, req.VerifyCode) {
			return errors.New("验证码错误或已过期")
		}
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE user = ?", req.Username).Scan(&count)
	if count > 0 {
		return errors.New("该账号已存在")
	}

	parentUID := 1
	inviteGradeID := 0
	addPrice := 1.0
	if req.Invite != "" {
		err := database.DB.QueryRow("SELECT uid, COALESCE(invite_grade_id,0) FROM qingka_wangke_user WHERE yqm = ?", req.Invite).Scan(&parentUID, &inviteGradeID)
		if err != nil {
			return errors.New("邀请码无效")
		}
		record, err := classmodule.Classes().GetGradeByID(inviteGradeID, true)
		if err != nil {
			return errors.New("邀请码暂未配置邀请等级")
		}
		addPrice = record.Rate
	} else if s.adminConfigEnabled("user_yqzc") {
		return errors.New("请输入邀请码")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_user (uuid, name, user, pass, email, grade_id, invite_grade_id, addprice, addtime, active, grade, money) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), '1', '0', 0)",
		parentUID, req.Nickname, req.Username, string(hashedPass), req.Email, inviteGradeID, inviteGradeID, addPrice,
	)
	if err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}
	return nil
}

func (s *Service) SendResetCode(email string) error {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE email = ?", email).Scan(&count)
	if count == 0 {
		return nil
	}

	if err := commonmodule.CheckVerificationRateLimit("reset_pwd", email); err != nil {
		return err
	}

	code, err := commonmodule.GenerateVerificationCode("reset_pwd", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	subject := siteName + " - 重置密码验证码"
	htmlBody := templateResetPassword(siteName, code, 10)

	if err := commonmodule.SendEmailWithType(email, subject, htmlBody, "reset"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

func (s *Service) ResetPassword(email, code, password string) error {
	if len(password) < 6 {
		return errors.New("新密码至少6位")
	}
	if !commonmodule.VerifyVerificationCode("reset_pwd", email, code) {
		return errors.New("验证码错误或已过期")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}
	result, err := database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE email = ?", string(hashedPass), email)
	if err != nil {
		return fmt.Errorf("重置密码失败: %v", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("该邮箱未绑定任何账号")
	}

	go func() {
		var username string
		database.DB.QueryRow("SELECT user FROM qingka_wangke_user WHERE email = ?", email).Scan(&username)
		siteName := s.getSiteName()
		htmlBody := templatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		if err := commonmodule.SendEmail(email, siteName+" - 密码修改通知", htmlBody); err != nil {
			obslogger.L().Warn("发送密码修改通知失败", "error", err)
		}
	}()

	return nil
}

// CheckLoginDevice 检测登录设备变化，新设备异步发安全提醒邮件。
func (s *Service) CheckLoginDevice(uid int, username, ip, ua string) {
	var email string
	database.DB.QueryRow("SELECT COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&email)
	if email == "" {
		return
	}

	fingerprint := fmt.Sprintf("%x", md5.Sum([]byte(ip+ua)))
	ctx := context.Background()
	deviceKey := fmt.Sprintf("login_devices:%d", uid)

	exists, _ := cache.RDB.SIsMember(ctx, deviceKey, fingerprint).Result()
	if exists {
		return
	}

	cache.RDB.SAdd(ctx, deviceKey, fingerprint)
	cache.RDB.Expire(ctx, deviceKey, 30*24*time.Hour)

	count, _ := cache.RDB.SCard(ctx, deviceKey).Result()
	if count <= 1 {
		return
	}

	go func() {
		siteName := s.getSiteName()
		loginTime := time.Now().Format("2006-01-02 15:04:05")
		htmlBody := templateLoginAlert(siteName, username, ip, ua, loginTime)
		if err := commonmodule.SendEmail(email, siteName+" - 登录安全提醒", htmlBody); err != nil {
			obslogger.L().Warn("发送登录提醒邮件失败", "error", err)
		}
	}()
}

func (s *Service) RefreshAccessToken(refreshToken string) (string, string, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Global.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("RefreshToken 无效或已过期")
	}

	var user model.User
	err = database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, name, money, grade, active FROM qingka_wangke_user WHERE uid = ?",
		claims.UID,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil {
		return "", "", errors.New("用户不存在")
	}

	if user.Active != "1" {
		return "", "", errors.New("账号已被禁用")
	}
	if claims.TokenType != "refresh" {
		return "", "", errors.New("RefreshToken 类型无效")
	}

	newAccessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL, "access")
	if err != nil {
		return "", "", errors.New("生成 Token 失败")
	}
	newRefreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL, "refresh")
	if err != nil {
		return "", "", errors.New("生成 RefreshToken 失败")
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *Service) GetUserInfo(uid int) (*model.VbenUserInfo, error) {
	user, err := s.users.GetByIDBasic(uid)
	if err != nil {
		return nil, err
	}

	return &model.VbenUserInfo{
		UserId:   fmt.Sprintf("%d", user.UID),
		Username: user.User,
		Avatar:   "",
		Desc:     fmt.Sprintf("余额: %.2f", user.Money),
		RealName: user.Name,
		HomePath: "/dashboard",
		Roles:    gradeToRoles(user.Grade),
	}, nil
}

func (s *Service) GetAccessCodes(grade string) []string {
	g, _ := strconvAtoi(grade)
	codes := []string{"AC_USER"}
	if g >= 1 {
		codes = append(codes, "AC_AGENT")
	}
	if g >= 2 {
		codes = append(codes, "admin", "AC_ADMIN", "AC_ADMIN_CONFIG", "AC_ADMIN_CLASS", "AC_ADMIN_STATS")
	}
	if g >= 3 {
		codes = append(codes, "super")
	}
	return codes
}

func (s *Service) Impersonate(operatorUID int, operatorGrade string, targetUID int, clientIP string) (*model.VbenLoginResponse, string, error) {
	if operatorGrade != "3" {
		return nil, "", errors.New("需要超级管理员权限")
	}
	if targetUID <= 0 {
		return nil, "", errors.New("目标用户不存在")
	}

	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, name, money, grade, active FROM qingka_wangke_user WHERE uid = ?",
		targetUID,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil {
		return nil, "", errors.New("目标用户不存在")
	}
	if user.Active != "1" {
		return nil, "", errors.New("目标用户已被禁用")
	}
	if user.Grade == "3" {
		return nil, "", errors.New("禁止伪装超级管理员账号")
	}

	accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL, "access")
	if err != nil {
		return nil, "", errors.New("生成 Token 失败")
	}
	refreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL, "refresh")
	if err != nil {
		return nil, "", errors.New("生成 RefreshToken 失败")
	}
	s.logImpersonation(operatorUID, targetUID, clientIP)

	return &model.VbenLoginResponse{
		AccessToken: accessToken,
		UserId:      fmt.Sprintf("%d", user.UID),
		Username:    user.User,
		RealName:    user.Name,
		Avatar:      "",
		Desc:        fmt.Sprintf("等级: %s", user.Grade),
		HomePath:    "/dashboard",
		Roles:       gradeToRoles(user.Grade),
	}, refreshToken, nil
}

func (s *Service) adminConfigEnabled(key string) bool {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf[key] == "1"
}

func (s *Service) getSiteName() string {
	conf, _ := commonmodule.GetAdminConfigMap()
	siteName := conf["sitename"]
	if siteName == "" {
		siteName = "System"
	}
	return siteName
}

func (s *Service) generateToken(user model.User, ttl int, tokenType string) (string, error) {
	claims := middleware.Claims{
		UID:       user.UID,
		User:      user.User,
		Grade:     user.Grade,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Global.JWT.Secret))
}

func (s *Service) logImpersonation(operatorUID, targetUID int, clientIP string) {
	if database.DB == nil {
		return
	}
	_, _ = database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, '管理员代登录', ?, '0', '', ?, NOW())",
		operatorUID,
		fmt.Sprintf("代登录目标UID=%d", targetUID),
		clientIP,
	)
}

// SetRefreshTokenCookie 设置 refresh_token 到 cookie。
func SetRefreshTokenCookie(w http.ResponseWriter, token string, ttl int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   ttl,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func gradeToRoles(grade string) []string {
	switch grade {
	case "3":
		return []string{"super", "admin", "user"}
	case "2":
		return []string{"admin", "user"}
	default:
		return []string{"user"}
	}
}

func strconvAtoi(v string) (int, error) {
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid integer")
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

func emailLayout(siteName, title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f5f7;font-family:'Segoe UI',Arial,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f5f7;padding:40px 0;">
<tr><td align="center">
<table width="520" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.06);overflow:hidden;">
  <tr><td style="background:linear-gradient(135deg,#1890ff,#096dd9);padding:28px 32px;">
    <h1 style="margin:0;color:#fff;font-size:20px;">%s</h1>
  </td></tr>
  <tr><td style="padding:32px;">
    <h2 style="margin:0 0 16px;color:#333;font-size:18px;">%s</h2>
    %s
  </td></tr>
  <tr><td style="padding:20px 32px;background:#fafafa;border-top:1px solid #f0f0f0;">
    <p style="margin:0;color:#999;font-size:12px;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, siteName, title, body)
}

func templateVerifyCode(siteName, code string, expireMinutes int) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您正在注册账号，请使用以下验证码完成注册：</p>
    <div style="text-align:center;margin:24px 0;">
      <span style="display:inline-block;padding:12px 32px;background:#f0f5ff;border:2px dashed #1890ff;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#1890ff;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;">验证码 %d 分钟内有效，请勿将验证码泄露给他人。</p>`, code, expireMinutes)
	return emailLayout(siteName, "注册验证码", body)
}

func templateResetPassword(siteName, code string, expireMinutes int) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您正在重置登录密码，请使用以下验证码：</p>
    <div style="text-align:center;margin:24px 0;">
      <span style="display:inline-block;padding:12px 32px;background:#fff7e6;border:2px dashed #fa8c16;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#fa8c16;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;">验证码 %d 分钟内有效。如非本人操作，请忽略此邮件。</p>`, code, expireMinutes)
	return emailLayout(siteName, "重置密码验证码", body)
}

func templateLoginAlert(siteName, username, ip, ua, loginTime string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号刚刚在一台新设备上登录，详情如下：</p>
    <table style="width:100%%;margin:16px 0;border-collapse:collapse;">
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;width:80px;">账号</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">登录IP</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">设备</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;word-break:break-all;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">时间</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
    </table>
    <p style="color:#ff4d4f;font-size:13px;">⚠️ 如果这不是您本人的操作，请立即修改密码。</p>`, username, ip, ua, loginTime)
	return emailLayout(siteName, "登录安全提醒", body)
}

func templatePasswordChanged(siteName, username, changeTime string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 的登录密码已于 <strong>%s</strong> 成功修改。</p>
    <p style="color:#ff4d4f;font-size:13px;">⚠️ 如果这不是您本人的操作，请立即联系管理员。</p>`, username, changeTime)
	return emailLayout(siteName, "密码修改通知", body)
}
