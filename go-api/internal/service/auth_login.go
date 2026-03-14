package service

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(req model.LoginRequest) (*model.VbenLoginResponse, string, error) {
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
			log.Printf("[Auth] 用户 uid=%d 使用明文密码登录，已自动迁移为 bcrypt", user.UID)
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

	accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL)
	if err != nil {
		return nil, "", errors.New("生成 Token 失败")
	}
	refreshToken, err := s.generateToken(user, config.Global.JWT.RefreshTTL)
	if err != nil {
		return nil, "", errors.New("生成 RefreshToken 失败")
	}

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

// CheckLoginDevice 检测登录设备变化，新设备异步发安全提醒邮件。
func (s *AuthService) CheckLoginDevice(uid int, username, ip, ua string) {
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
		if err := SendEmail(email, siteName+" - 登录安全提醒", htmlBody); err != nil {
			log.Printf("发送登录提醒邮件失败: %v", err)
		}
	}()
}

func (s *AuthService) getSiteName() string {
	return getConfiguredSiteName()
}
