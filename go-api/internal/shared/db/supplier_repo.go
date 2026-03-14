package db

import (
	"database/sql"
	"fmt"

	"go-api/internal/database"
	"go-api/internal/model"
)

type SupplierRepo struct{}

func NewSupplierRepo() *SupplierRepo {
	return &SupplierRepo{}
}

func (r *SupplierRepo) GetFullByHID(hid int) (*model.SupplierFull, error) {
	var sup model.SupplierFull
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(ip,''), COALESCE(cookie,''), COALESCE(money,'0'), COALESCE(status,'1') FROM qingka_wangke_huoyuan WHERE hid = ?",
		hid,
	).Scan(&sup.HID, &sup.PT, &sup.Name, &sup.URL, &sup.User, &sup.Pass, &sup.Token, &sup.IP, &sup.Cookie, &sup.Money, &sup.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("供应商不存在")
	}
	if err != nil {
		return nil, err
	}
	return &sup, nil
}
