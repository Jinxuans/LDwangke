package service

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/middleware"
	"go-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// RefreshAccessToken 从 cookie 中的 refresh_token 刷新。
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

// Impersonate 管理员免登录进入代理界面。
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
