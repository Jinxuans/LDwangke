package service

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/middleware"
	"go-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.VbenLoginResponse, string, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, IFNULL(pass2,''), name, money, grade, active FROM qingka_wangke_user WHERE user = ?",
		req.Username,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Pass2, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil && strings.Contains(err.Error(), "pass2") {
		// pass2 列不存在，fallback 不查 pass2
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

	// 密码验证：支持 bcrypt 和明文（自动迁移）
	if strings.HasPrefix(user.Pass, "$2a$") || strings.HasPrefix(user.Pass, "$2b$") {
		// bcrypt 哈希验证
		if err := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(req.Password)); err != nil {
			return nil, "", errors.New("密码错误")
		}
	} else {
		// 明文比对（兼容旧数据）
		if user.Pass != req.Password {
			return nil, "", errors.New("密码错误")
		}
		// 自动迁移：将明文密码升级为 bcrypt
		if hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost); err == nil {
			database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", string(hashed), user.UID)
		}
	}

	// 二级密码开关：查询系统配置 pass2_kg，默认开启("1")
	pass2Enabled := true
	var pass2Kg string
	if err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v`='pass2_kg'").Scan(&pass2Kg); err == nil {
		if pass2Kg == "0" {
			pass2Enabled = false
		}
	}

	if pass2Enabled && user.Grade == "3" {
		// 如果是管理员且未提供二级密码，返回特殊错误码触发前端弹窗
		if req.Pass2 == "" {
			return nil, "", errors.New("NEED_ADMIN_AUTH")
		}
		// 如果提供了二级密码，校验是否正确（使用独立的 pass2 字段）
		if user.Pass2 == "" {
			// 尚未设置二级密码，跳过校验（首次登录自动通过）
		} else if strings.HasPrefix(user.Pass2, "$2a$") || strings.HasPrefix(user.Pass2, "$2b$") {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Pass2), []byte(req.Pass2)); err != nil {
				return nil, "", errors.New("管理员验证失败")
			}
		} else if user.Pass2 != req.Pass2 {
			return nil, "", errors.New("管理员验证失败")
		}
	}

	accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL)
	if err != nil {
		return nil, "", errors.New("生成 Token 失败")
	}

	refreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL)
	if err != nil {
		return nil, "", errors.New("生成 RefreshToken 失败")
	}

	// 更新最后登录时间（endtime 用于迁移7天检查，与旧系统一致）
	_, _ = database.DB.Exec("UPDATE qingka_wangke_user SET lasttime = NOW(), endtime = NOW() WHERE uid = ?", user.UID)

	roles := s.gradeToRoles(user.Grade)

	return &model.VbenLoginResponse{
		AccessToken: accessToken,
		UserId:      fmt.Sprintf("%d", user.UID),
		Username:    user.User,
		RealName:    user.User,
		Avatar:      "",
		Desc:        fmt.Sprintf("等级: %s", user.Grade),
		HomePath:    "/dashboard",
		Roles:       roles,
	}, refreshToken, nil
}

// SendRegisterCode 发送注册验证码到邮箱
func (s *AuthService) SendRegisterCode(email string) error {
	vs := NewVerificationService()
	if err := vs.RateLimitCheck("register", email); err != nil {
		return err
	}

	code, err := vs.GenerateCode("register", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	ts := NewEmailTemplateService()
	vars := map[string]string{
		"site_name":      siteName,
		"code":           code,
		"expire_minutes": "10",
		"email":          email,
		"time":           time.Now().Format("2006-01-02 15:04:05"),
	}
	subject, htmlBody, tplErr := ts.RenderByCode("register", vars)
	if tplErr != nil {
		subject = siteName + " - 注册验证码"
		htmlBody = TemplateVerifyCode(siteName, code, 10)
	}

	es := NewEmailService()
	if err := es.SendEmailWithType(email, subject, htmlBody, "register"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}

func (s *AuthService) Register(req model.RegisterRequest) error {
	// 校验账号是否为QQ号: 5-11位纯数字，首位不为0
	qqRe := regexp.MustCompile(`^[1-9][0-9]{4,10}$`)
	if !qqRe.MatchString(req.Username) {
		return errors.New("账号必须为有效的 QQ 号（5-11位数字，首位不为0）")
	}

	// 读取系统配置
	conf, _ := NewAdminService().GetConfig()

	// 如果开启了邮箱验证，校验验证码
	if conf["login_email_verify"] == "1" {
		if req.Email == "" {
			return errors.New("请填写邮箱地址")
		}
		if req.VerifyCode == "" {
			return errors.New("请输入邮箱验证码")
		}
		vs := NewVerificationService()
		if !vs.VerifyCode("register", req.Email, req.VerifyCode) {
			return errors.New("验证码错误或已过期")
		}
	}

	// 检查账号是否已存在
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE user = ?", req.Username).Scan(&count)
	if count > 0 {
		return errors.New("该账号已存在")
	}

	// 处理邀请码 (旧系统通过邀请码获取上级 UID 和费率)
	var parentUID int = 1
	var addPrice float64 = 1.0
	if req.Invite != "" {
		err := database.DB.QueryRow("SELECT uid, addprice FROM qingka_wangke_user WHERE yqm = ?", req.Invite).Scan(&parentUID, &addPrice)
		if err != nil {
			return errors.New("邀请码无效")
		}
	} else if conf["user_yqzc"] == "1" {
		// 如果开启了邀请码注册但未提供邀请码
		return errors.New("请输入邀请码")
	}

	// 注册时使用 bcrypt 加密密码
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	emailVal := req.Email
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_user (uuid, name, user, pass, email, addprice, addtime, active, grade, money) VALUES (?, ?, ?, ?, ?, ?, NOW(), '1', '0', 0)",
		parentUID, req.Nickname, req.Username, string(hashedPass), emailVal, addPrice,
	)
	if err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}

	return nil
}

// SendResetCode 发送重置密码验证码
func (s *AuthService) SendResetCode(email string) error {
	// 检查邮箱是否关联了用户
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE email = ?", email).Scan(&count)
	if count == 0 {
		return errors.New("该邮箱未绑定任何账号")
	}

	vs := NewVerificationService()
	if err := vs.RateLimitCheck("reset_pwd", email); err != nil {
		return err
	}

	code, err := vs.GenerateCode("reset_pwd", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	ts := NewEmailTemplateService()
	vars := map[string]string{
		"site_name":      siteName,
		"code":           code,
		"expire_minutes": "10",
		"email":          email,
		"time":           time.Now().Format("2006-01-02 15:04:05"),
	}
	subject, htmlBody, tplErr := ts.RenderByCode("reset_password", vars)
	if tplErr != nil {
		subject = siteName + " - 重置密码验证码"
		htmlBody = TemplateResetPassword(siteName, code, 10)
	}

	es := NewEmailService()
	if err := es.SendEmailWithType(email, subject, htmlBody, "reset"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}

// ResetPassword 通过邮箱验证码重置密码
func (s *AuthService) ResetPassword(email, code, newPass string) error {
	if len(newPass) < 6 {
		return errors.New("新密码至少6位")
	}

	vs := NewVerificationService()
	if !vs.VerifyCode("reset_pwd", email, code) {
		return errors.New("验证码错误或已过期")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
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

	// 异步发送密码修改通知
	go func() {
		var username string
		database.DB.QueryRow("SELECT user FROM qingka_wangke_user WHERE email = ?", email).Scan(&username)
		siteName := s.getSiteName()
		htmlBody := TemplatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		NewEmailService().SendEmail(email, siteName+" - 密码修改通知", htmlBody)
	}()

	return nil
}

// CheckLoginDevice 检测登录设备变化，新设备异步发安全提醒邮件
func (s *AuthService) CheckLoginDevice(uid int, username, ip, ua string) {
	// 查用户邮箱
	var email string
	database.DB.QueryRow("SELECT COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&email)
	if email == "" {
		return
	}

	// 设备指纹 = md5(ip + ua)
	fingerprint := fmt.Sprintf("%x", md5.Sum([]byte(ip+ua)))
	ctx := context.Background()
	deviceKey := fmt.Sprintf("login_devices:%d", uid)

	// 检查是否为已知设备
	exists, _ := cache.RDB.SIsMember(ctx, deviceKey, fingerprint).Result()
	if exists {
		return
	}

	// 添加到已知设备集合，设置30天过期
	cache.RDB.SAdd(ctx, deviceKey, fingerprint)
	cache.RDB.Expire(ctx, deviceKey, 30*24*time.Hour)

	// 如果是首次登录（设备集合刚创建），不发提醒
	count, _ := cache.RDB.SCard(ctx, deviceKey).Result()
	if count <= 1 {
		return
	}

	// 异步发送安全提醒邮件
	go func() {
		siteName := s.getSiteName()
		loginTime := time.Now().Format("2006-01-02 15:04:05")
		htmlBody := TemplateLoginAlert(siteName, username, ip, ua, loginTime)
		if err := NewEmailService().SendEmail(email, siteName+" - 登录安全提醒", htmlBody); err != nil {
			log.Printf("发送登录提醒邮件失败: %v", err)
		}
	}()
}

// getSiteName 从系统配置获取站点名称
func (s *AuthService) getSiteName() string {
	conf, _ := NewAdminService().GetConfig()
	name := conf["sitename"]
	if name == "" {
		name = "System"
	}
	return name
}

// RefreshAccessToken 从 cookie 中的 refresh_token 刷新，返回新的 accessToken 裸字符串
func (s *AuthService) RefreshAccessToken(refreshToken string) (string, string, error) {
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

	newAccessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL)
	if err != nil {
		return "", "", errors.New("生成 Token 失败")
	}

	newRefreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL)
	if err != nil {
		return "", "", errors.New("生成 RefreshToken 失败")
	}

	return newAccessToken, newRefreshToken, nil
}

// GetUserInfo 根据 uid 返回 Vben 格式的用户信息
func (s *AuthService) GetUserInfo(uid int) (*model.VbenUserInfo, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, name, money, grade, active FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	roles := s.gradeToRoles(user.Grade)

	return &model.VbenUserInfo{
		UserId:   fmt.Sprintf("%d", user.UID),
		Username: user.User,
		Avatar:   "",
		Desc:     fmt.Sprintf("余额: %.2f", user.Money),
		RealName: user.Name,
		HomePath: "/dashboard",
		Roles:    roles,
	}, nil
}

// GetAccessCodes 根据用户等级返回权限码
func (s *AuthService) GetAccessCodes(grade string) []string {
	g, _ := strconv.Atoi(grade)
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

func (s *AuthService) gradeToRoles(grade string) []string {
	switch grade {
	case "3":
		return []string{"super", "admin", "user"}
	case "2":
		return []string{"admin", "user"}
	default:
		return []string{"user"}
	}
}

// Impersonate 管理员免登录进入代理界面（生成目标用户的 token）
func (s *AuthService) Impersonate(targetUID int) (*model.VbenLoginResponse, string, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, name, money, grade, active FROM qingka_wangke_user WHERE uid = ?",
		targetUID,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err != nil {
		return nil, "", errors.New("目标用户不存在")
	}

	accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL)
	if err != nil {
		return nil, "", errors.New("生成 Token 失败")
	}
	refreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL)
	if err != nil {
		return nil, "", errors.New("生成 RefreshToken 失败")
	}

	roles := s.gradeToRoles(user.Grade)

	return &model.VbenLoginResponse{
		AccessToken: accessToken,
		UserId:      fmt.Sprintf("%d", user.UID),
		Username:    user.User,
		RealName:    user.Name,
		Avatar:      "",
		Desc:        fmt.Sprintf("等级: %s", user.Grade),
		HomePath:    "/dashboard",
		Roles:       roles,
	}, refreshToken, nil
}

// SetRefreshTokenCookie 设置 refresh_token 到 cookie
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

func (s *AuthService) generateToken(user model.User, ttl int) (string, error) {
	claims := middleware.Claims{
		UID:   user.UID,
		User:  user.User,
		Grade: user.Grade,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Global.JWT.Secret))
}
