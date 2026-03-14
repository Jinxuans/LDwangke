package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserCenterService) ChangePassword(uid int, oldPass, newPass string) error {
	var dbPass string
	err := database.DB.QueryRow("SELECT pass FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&dbPass)
	if err != nil {
		return errors.New("用户不存在")
	}

	if oldPass != dbPass {
		return errors.New("旧密码错误")
	}
	if len(newPass) < 6 {
		return errors.New("新密码至少6位")
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, uid)
	if err != nil {
		return err
	}

	go func() {
		var username, email string
		database.DB.QueryRow("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&username, &email)
		if email == "" {
			return
		}
		siteName := getConfiguredSiteName()
		htmlBody := templatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		SendEmail(email, siteName+" - 密码修改通知", htmlBody)
	}()

	return nil
}

func (s *UserCenterService) ChangePass2(uid int, oldPass2, newPass2 string) error {
	var grade, dbPass2 string
	err := database.DB.QueryRow("SELECT grade, IFNULL(pass2,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&grade, &dbPass2)
	if err != nil && strings.Contains(err.Error(), "pass2") {
		err = database.DB.QueryRow("SELECT grade FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&grade)
	}
	if err != nil {
		return errors.New("用户不存在")
	}
	if grade != "3" {
		return errors.New("仅管理员可设置二级密码")
	}
	if dbPass2 != "" {
		if strings.HasPrefix(dbPass2, "$2a$") || strings.HasPrefix(dbPass2, "$2b$") {
			if err := bcrypt.CompareHashAndPassword([]byte(dbPass2), []byte(oldPass2)); err != nil {
				return errors.New("旧二级密码错误")
			}
		} else if dbPass2 != oldPass2 {
			return errors.New("旧二级密码错误")
		}
	}
	if len(newPass2) < 6 {
		return errors.New("新二级密码至少6位")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPass2), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("加密失败")
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET pass2 = ? WHERE uid = ?", string(hashed), uid)
	return err
}

func (s *UserCenterService) SendChangeEmailCode(uid int, newEmail string) error {
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if err := CheckVerificationRateLimit("change_email", identifier); err != nil {
		return err
	}

	code, err := GenerateVerificationCode("change_email", identifier, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := getConfiguredSiteName()
	htmlBody := templateChangeEmailCode(siteName, code, 10)

	if err := SendEmail(newEmail, siteName+" - 邮箱变更验证码", htmlBody); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

func (s *UserCenterService) ChangeEmail(uid int, newEmail, code string) error {
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if !VerifyVerificationCode("change_email", identifier, code) {
		return errors.New("验证码错误或已过期")
	}

	var oldEmail, username string
	database.DB.QueryRow("SELECT COALESCE(email,''), COALESCE(user,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&oldEmail, &username)

	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET email = ? WHERE uid = ?", newEmail, uid)
	if err != nil {
		return fmt.Errorf("更新邮箱失败: %v", err)
	}

	if oldEmail != "" && oldEmail != newEmail {
		go func() {
			siteName := getConfiguredSiteName()
			htmlBody := templateEmailChanged(siteName, username, newEmail, time.Now().Format("2006-01-02 15:04:05"))
			SendEmail(oldEmail, siteName+" - 邮箱变更通知", htmlBody)
		}()
	}

	return nil
}
