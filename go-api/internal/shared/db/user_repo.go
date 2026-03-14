package db

import (
	"database/sql"
	"fmt"

	"go-api/internal/database"
	"go-api/internal/model"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) GetByIDBasic(uid int) (*model.User, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, name, money, grade, active FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.QueryRow(
		"SELECT uid, uuid, user, pass, IFNULL(pass2,''), name, money, grade, active FROM qingka_wangke_user WHERE user = ?",
		username,
	).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Pass2, &user.Name, &user.Money, &user.Grade, &user.Active)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
