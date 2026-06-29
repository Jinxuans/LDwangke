package class

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go-api/internal/database"
)

func normalizeMiJiaMode(mode string, defaultMode string) (string, error) {
	value := strings.TrimSpace(mode)
	if value == "" {
		value = defaultMode
	}

	// 新枚举只保留 0/1/2/3，4 只在读取旧数据时做兼容解释。
	switch value {
	case "0", "1", "2", "3":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported mijia mode: %s", mode)
	}
}

func validateMiJiaPrice(mode, price string) (string, error) {
	value := strings.TrimSpace(price)
	if value == "" {
		return "", fmt.Errorf("密价值不能为空")
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", fmt.Errorf("密价值必须是数字")
	}
	if parsed < 0 {
		return "", fmt.Errorf("密价值不能小于 0")
	}
	if mode == "3" {
		if parsed <= 0 {
			return "", fmt.Errorf("倍率必须大于 0")
		}
		if parsed > 1 {
			return "", fmt.Errorf("密价倍率不能大于 1")
		}
	}
	return value, nil
}

func normalizeMiJiaScopeRecord(scopeType string, scopeID int, cid int) (string, int, int) {
	scopeType = strings.ToLower(strings.TrimSpace(scopeType))
	switch scopeType {
	case "category":
		return "category", scopeID, 0
	case "product":
		if scopeID <= 0 {
			scopeID = cid
		}
		return "product", scopeID, scopeID
	default:
		if cid > 0 {
			return "product", cid, cid
		}
		return "category", scopeID, 0
	}
}

func validateMiJiaTarget(uid int, scopeType string, scopeID int) error {
	if uid <= 0 {
		return fmt.Errorf("用户 UID 无效")
	}
	if scopeID <= 0 {
		return fmt.Errorf("密价范围无效")
	}

	var exists int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("用户不存在")
	}

	switch scopeType {
	case "product":
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE cid = ?", scopeID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("商品不存在")
		}
	case "category":
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE id = ? AND status != 3", scopeID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("分类不存在或已删除")
		}
	default:
		return fmt.Errorf("不支持的密价范围")
	}

	return nil
}

func upsertMiJiaRecordTx(tx *sql.Tx, uid, cid int, mode, price string) error {
	return upsertMiJiaScopeRecordTx(tx, uid, cid, "product", cid, mode, price)
}

func upsertMiJiaScopeRecordTx(tx *sql.Tx, uid, cid int, scopeType string, scopeID int, mode, price string) error {
	scopeType, scopeID, cid = normalizeMiJiaScopeRecord(scopeType, scopeID, cid)
	var err error
	price, err = validateMiJiaPrice(mode, price)
	if err != nil {
		return err
	}
	if err := validateMiJiaTarget(uid, scopeType, scopeID); err != nil {
		return err
	}

	var mid int
	if scopeType == "category" {
		err = tx.QueryRow(
			"SELECT mid FROM qingka_wangke_mijia WHERE uid = ? AND scope_type = 'category' AND scope_id = ? ORDER BY mid ASC LIMIT 1",
			uid, scopeID,
		).Scan(&mid)
	} else {
		err = tx.QueryRow(
			"SELECT mid FROM qingka_wangke_mijia WHERE uid = ? AND ((scope_type = 'product' AND scope_id = ?) OR (scope_type = '' AND cid = ?)) ORDER BY mid ASC LIMIT 1",
			uid, scopeID, scopeID,
		).Scan(&mid)
	}

	switch err {
	case nil:
		if _, err := tx.Exec("UPDATE qingka_wangke_mijia SET uid = ?, cid = ?, scope_type = ?, scope_id = ?, mode = ?, price = ? WHERE mid = ?", uid, cid, scopeType, scopeID, mode, price, mid); err != nil {
			return err
		}
		if scopeType == "category" {
			_, err = tx.Exec("DELETE FROM qingka_wangke_mijia WHERE uid = ? AND scope_type = 'category' AND scope_id = ? AND mid <> ?", uid, scopeID, mid)
			return err
		}
		_, err = tx.Exec("DELETE FROM qingka_wangke_mijia WHERE uid = ? AND ((scope_type = 'product' AND scope_id = ?) OR (scope_type = '' AND cid = ?)) AND mid <> ?", uid, scopeID, scopeID, mid)
		return err
	case sql.ErrNoRows:
		_, err = tx.Exec(
			"INSERT INTO qingka_wangke_mijia (uid, cid, scope_type, scope_id, mode, price, addtime) VALUES (?, ?, ?, ?, ?, ?, NOW())",
			uid, cid, scopeType, scopeID, mode, price,
		)
		return err
	default:
		return err
	}
}

func saveMiJiaRecord(uid, cid int, mode, price string) error {
	return saveMiJiaScopeRecord(uid, cid, "product", cid, mode, price)
}

func saveMiJiaScopeRecord(uid, cid int, scopeType string, scopeID int, mode, price string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertMiJiaScopeRecordTx(tx, uid, cid, scopeType, scopeID, mode, price); err != nil {
		return err
	}
	return tx.Commit()
}

func updateMiJiaRecord(mid, uid, cid int, mode, price string) error {
	return updateMiJiaScopeRecord(mid, uid, cid, "product", cid, mode, price)
}

func updateMiJiaScopeRecord(mid, uid, cid int, scopeType string, scopeID int, mode, price string) error {
	scopeType, scopeID, cid = normalizeMiJiaScopeRecord(scopeType, scopeID, cid)
	var err error
	price, err = validateMiJiaPrice(mode, price)
	if err != nil {
		return err
	}
	if err := validateMiJiaTarget(uid, scopeType, scopeID); err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"UPDATE qingka_wangke_mijia SET uid = ?, cid = ?, scope_type = ?, scope_id = ?, mode = ?, price = ? WHERE mid = ?",
		uid, cid, scopeType, scopeID, mode, price, mid,
	); err != nil {
		return err
	}
	if scopeType == "category" {
		if _, err := tx.Exec(
			"DELETE FROM qingka_wangke_mijia WHERE uid = ? AND scope_type = 'category' AND scope_id = ? AND mid <> ?",
			uid, scopeID, mid,
		); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(
			"DELETE FROM qingka_wangke_mijia WHERE uid = ? AND ((scope_type = 'product' AND scope_id = ?) OR (scope_type = '' AND cid = ?)) AND mid <> ?",
			uid, scopeID, scopeID, mid,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
