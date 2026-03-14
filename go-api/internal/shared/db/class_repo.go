package db

import (
	"database/sql"
	"fmt"

	"go-api/internal/database"
	"go-api/internal/model"
)

type ClassRepo struct{}

func NewClassRepo() *ClassRepo {
	return &ClassRepo{}
}

func (r *ClassRepo) GetFullByCID(cid int) (*model.ClassFull, error) {
	var cls model.ClassFull
	err := database.DB.QueryRow(
		"SELECT cid, COALESCE(name,''), COALESCE(noun,''), COALESCE(price,'0'), COALESCE(docking,'0'), COALESCE(fenlei,''), COALESCE(status,0), COALESCE(yunsuan,'*'), COALESCE(content,'') FROM qingka_wangke_class WHERE cid = ?",
		cid,
	).Scan(&cls.CID, &cls.Name, &cls.Noun, &cls.Price, &cls.Docking, &cls.Fenlei, &cls.Status, &cls.Yunsuan, &cls.Content)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("课程不存在")
	}
	if err != nil {
		return nil, err
	}
	return &cls, nil
}
