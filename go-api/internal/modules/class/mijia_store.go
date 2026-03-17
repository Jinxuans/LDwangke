package class

import (
	"database/sql"
	"fmt"
	"strings"

	"go-api/internal/database"
)

func normalizeMiJiaMode(mode string, defaultMode string) (string, error) {
	value := strings.TrimSpace(mode)
	if value == "" {
		value = defaultMode
	}

	switch value {
	case "0", "1", "2", "4":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported mijia mode: %s", mode)
	}
}

func upsertMiJiaRecordTx(tx *sql.Tx, uid, cid int, mode, price string) error {
	var mid int
	err := tx.QueryRow(
		"SELECT mid FROM qingka_wangke_mijia WHERE uid = ? AND cid = ? ORDER BY mid ASC LIMIT 1",
		uid, cid,
	).Scan(&mid)
	switch err {
	case nil:
		if _, err := tx.Exec("UPDATE qingka_wangke_mijia SET mode = ?, price = ? WHERE mid = ?", mode, price, mid); err != nil {
			return err
		}
		_, err = tx.Exec("DELETE FROM qingka_wangke_mijia WHERE uid = ? AND cid = ? AND mid <> ?", uid, cid, mid)
		return err
	case sql.ErrNoRows:
		_, err = tx.Exec(
			"INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())",
			uid, cid, mode, price,
		)
		return err
	default:
		return err
	}
}

func saveMiJiaRecord(uid, cid int, mode, price string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertMiJiaRecordTx(tx, uid, cid, mode, price); err != nil {
		return err
	}
	return tx.Commit()
}

func updateMiJiaRecord(mid, uid, cid int, mode, price string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"UPDATE qingka_wangke_mijia SET uid = ?, cid = ?, mode = ?, price = ? WHERE mid = ?",
		uid, cid, mode, price, mid,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"DELETE FROM qingka_wangke_mijia WHERE uid = ? AND cid = ? AND mid <> ?",
		uid, cid, mid,
	); err != nil {
		return err
	}

	return tx.Commit()
}
