package db

import (
	"database/sql"

	"go-api/internal/database"
)

type ConfigRepo struct{}

func NewConfigRepo() *ConfigRepo {
	return &ConfigRepo{}
}

// GetLegacyValue 读取旧配置表中以 v 为键、k 为值的配置项。
func (r *ConfigRepo) GetLegacyValue(key string) (string, error) {
	var value string
	err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = ? LIMIT 1", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (r *ConfigRepo) GetStructuredValue(skey string) (string, error) {
	var value string
	err := database.DB.QueryRow("SELECT COALESCE(svalue,'') FROM qingka_wangke_config WHERE skey = ? LIMIT 1", skey).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}
