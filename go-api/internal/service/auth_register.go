package service

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// SendRegisterCode 发送注册验证码到邮箱。
func (s *AuthService) SendRegisterCode(email string) error {
	if err := CheckVerificationRateLimit("register", email); err != nil {
		return err
	}

	code, err := GenerateVerificationCode("register", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	vars := map[string]string{
		"site_name":      siteName,
		"code":           code,
		"expire_minutes": "10",
		"email":          email,
		"time":           time.Now().Format("2006-01-02 15:04:05"),
	}
	subject, htmlBody, tplErr := renderEmailTemplateByCode("register", vars)
	if tplErr != nil {
		subject = siteName + " - 注册验证码"
		htmlBody = templateVerifyCode(siteName, code, 10)
	}

	if err := SendEmailWithType(email, subject, htmlBody, "register"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

func (s *AuthService) Register(req model.RegisterRequest) error {
	qqRe := regexp.MustCompile(`^[1-9][0-9]{4,10}$`)
	if !qqRe.MatchString(req.Username) {
		return errors.New("账号必须为有效的 QQ 号（5-11位数字，首位不为0）")
	}

	if adminConfigEnabled("login_email_verify") {
		if req.Email == "" {
			return errors.New("请填写邮箱地址")
		}
		if req.VerifyCode == "" {
			return errors.New("请输入邮箱验证码")
		}
		if !VerifyVerificationCode("register", req.Email, req.VerifyCode) {
			return errors.New("验证码错误或已过期")
		}
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE user = ?", req.Username).Scan(&count)
	if count > 0 {
		return errors.New("该账号已存在")
	}

	parentUID := 1
	addPrice := 1.0
	if req.Invite != "" {
		err := database.DB.QueryRow("SELECT uid, addprice FROM qingka_wangke_user WHERE yqm = ?", req.Invite).Scan(&parentUID, &addPrice)
		if err != nil {
			return errors.New("邀请码无效")
		}
	} else if adminConfigEnabled("user_yqzc") {
		return errors.New("请输入邀请码")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_user (uuid, name, user, pass, email, addprice, addtime, active, grade, money) VALUES (?, ?, ?, ?, ?, ?, NOW(), '1', '0', 0)",
		parentUID, req.Nickname, req.Username, string(hashedPass), req.Email, addPrice,
	)
	if err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}
	return nil
}

// SendResetCode 发送重置密码验证码。
func (s *AuthService) SendResetCode(email string) error {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE email = ?", email).Scan(&count)
	if count == 0 {
		return errors.New("该邮箱未绑定任何账号")
	}

	if err := CheckVerificationRateLimit("reset_pwd", email); err != nil {
		return err
	}

	code, err := GenerateVerificationCode("reset_pwd", email, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	vars := map[string]string{
		"site_name":      siteName,
		"code":           code,
		"expire_minutes": "10",
		"email":          email,
		"time":           time.Now().Format("2006-01-02 15:04:05"),
	}
	subject, htmlBody, tplErr := renderEmailTemplateByCode("reset_password", vars)
	if tplErr != nil {
		subject = siteName + " - 重置密码验证码"
		htmlBody = templateResetPassword(siteName, code, 10)
	}

	if err := SendEmailWithType(email, subject, htmlBody, "reset"); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

// ResetPassword 通过邮箱验证码重置密码。
func (s *AuthService) ResetPassword(email, code, newPass string) error {
	if len(newPass) < 6 {
		return errors.New("新密码至少6位")
	}
	if !VerifyVerificationCode("reset_pwd", email, code) {
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

	go func() {
		var username string
		database.DB.QueryRow("SELECT user FROM qingka_wangke_user WHERE email = ?", email).Scan(&username)
		siteName := s.getSiteName()
		htmlBody := templatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		if err := SendEmail(email, siteName+" - 密码修改通知", htmlBody); err != nil {
			log.Printf("发送密码修改通知失败: %v", err)
		}
	}()

	return nil
}
