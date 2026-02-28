package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-api/internal/database"

	"github.com/go-sql-driver/mysql"
)

// ===== 数据同步工具：从其他29系统数据库同步数据到当前系统 =====

// SyncRequest 同步请求参数
type SyncRequest struct {
	Host           string `json:"host" binding:"required"`
	Port           int    `json:"port"`
	DBName         string `json:"db_name" binding:"required"`
	User           string `json:"user" binding:"required"`
	Password       string `json:"password"`
	UpdateExisting bool   `json:"update_existing"`
}

// SyncResult 同步结果
type SyncResult struct {
	SyncTime   string           `json:"sync_time"`
	Success    bool             `json:"success"`
	Details    []SyncTableInfo  `json:"details"`
	Errors     []string         `json:"errors"`
	Summary    string           `json:"summary"`
}

// SyncTableInfo 每张表的同步统计
type SyncTableInfo struct {
	Table    string `json:"table"`
	Label    string `json:"label"`
	Total    int    `json:"total"`
	Inserted int    `json:"inserted"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
}

// SyncTestResult 连接测试结果
type SyncTestResult struct {
	Connected bool              `json:"connected"`
	Tables    map[string]int    `json:"tables"`
	Error     string            `json:"error,omitempty"`
}

type DBSyncService struct{}

func NewDBSyncService() *DBSyncService {
	return &DBSyncService{}
}

// connectExternal 连接外部数据库
func (s *DBSyncService) connectExternal(req SyncRequest) (*sql.DB, error) {
	if req.Port == 0 {
		req.Port = 3306
	}
	cfg := mysql.NewConfig()
	cfg.User = req.User
	cfg.Passwd = req.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", req.Host, req.Port)
	cfg.DBName = req.DBName
	cfg.ParseTime = true
	cfg.Collation = "utf8mb4_general_ci"
	cfg.Params = map[string]string{"sql_mode": "''"}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("无法连接到数据库: %v", err)
	}
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(2 * time.Minute)
	return db, nil
}

// TestConnection 测试外部数据库连接，返回各表行数
func (s *DBSyncService) TestConnection(req SyncRequest) (*SyncTestResult, error) {
	extDB, err := s.connectExternal(req)
	if err != nil {
		return &SyncTestResult{Connected: false, Error: err.Error()}, nil
	}
	defer extDB.Close()

	tables := map[string]string{
		"qingka_wangke_huoyuan": "货源",
		"qingka_wangke_user":    "用户",
		"qingka_wangke_fenlei":  "分类",
		"qingka_wangke_class":   "商品",
		"qingka_wangke_order":   "订单",
	}

	counts := make(map[string]int)
	for tbl := range tables {
		var cnt int
		err := extDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tbl)).Scan(&cnt)
		if err != nil {
			counts[tbl] = -1 // 表不存在
		} else {
			counts[tbl] = cnt
		}
	}

	return &SyncTestResult{Connected: true, Tables: counts}, nil
}

// Execute 执行完整数据同步
func (s *DBSyncService) Execute(req SyncRequest) (*SyncResult, error) {
	extDB, err := s.connectExternal(req)
	if err != nil {
		return nil, err
	}
	defer extDB.Close()

	result := &SyncResult{
		SyncTime: time.Now().Format("2006-01-02 15:04:05"),
		Success:  true,
	}

	// 按顺序同步：货源 → 用户 → 分类 → 商品 → 订单
	syncFuncs := []struct {
		name  string
		label string
		fn    func(*sql.DB, bool) (*SyncTableInfo, error)
	}{
		{"qingka_wangke_huoyuan", "货源", s.syncHuoyuan},
		{"qingka_wangke_user", "用户", s.syncUsers},
		{"qingka_wangke_fenlei", "分类", s.syncFenlei},
		{"qingka_wangke_class", "商品", s.syncClass},
		{"qingka_wangke_order", "订单", s.syncOrders},
	}

	totalInserted, totalUpdated := 0, 0
	for _, sf := range syncFuncs {
		info, err := sf.fn(extDB, req.UpdateExisting)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s同步失败: %v", sf.label, err))
			result.Details = append(result.Details, SyncTableInfo{
				Table: sf.name, Label: sf.label, Failed: -1,
			})
			log.Printf("[DBSync] %s同步失败: %v", sf.label, err)
			continue
		}
		info.Table = sf.name
		info.Label = sf.label
		result.Details = append(result.Details, *info)
		totalInserted += info.Inserted
		totalUpdated += info.Updated
		log.Printf("[DBSync] %s: 共%d条, 新增%d, 更新%d, 跳过%d",
			sf.label, info.Total, info.Inserted, info.Updated, info.Skipped)
	}

	if result.Errors == nil {
		result.Errors = []string{}
	}

	result.Summary = fmt.Sprintf("同步完成，共新增 %d 条、更新 %d 条数据", totalInserted, totalUpdated)
	if len(result.Errors) > 0 {
		result.Success = false
		result.Summary += fmt.Sprintf("，%d 项出错", len(result.Errors))
	}

	return result, nil
}

// ========== 各表同步逻辑 ==========

// syncHuoyuan 同步货源表
func (s *DBSyncService) syncHuoyuan(extDB *sql.DB, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	rows, err := extDB.Query("SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(ip,''), COALESCE(cookie,''), COALESCE(money,'0'), COALESCE(status,'1'), COALESCE(addtime,'') FROM qingka_wangke_huoyuan")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hid int
		var pt, name, url, user, pass, token, ip, cookie, money, status, addtime string
		if err := rows.Scan(&hid, &pt, &name, &url, &user, &pass, &token, &ip, &cookie, &money, &status, &addtime); err != nil {
			info.Failed++
			continue
		}
		info.Total++

		// 检查是否已存在（按 hid）
		var exists int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_huoyuan WHERE hid = ?", hid).Scan(&exists)

		if exists > 0 {
			if updateExisting {
				_, err := database.DB.Exec(
					"UPDATE qingka_wangke_huoyuan SET pt=?, name=?, url=?, user=?, pass=?, token=?, ip=?, cookie=?, money=?, status=?, addtime=? WHERE hid=?",
					pt, name, url, user, pass, token, ip, cookie, money, status, addtime, hid)
				if err != nil {
					info.Failed++
				} else {
					info.Updated++
				}
			} else {
				info.Skipped++
			}
		} else {
			_, err := database.DB.Exec(
				"INSERT INTO qingka_wangke_huoyuan (hid, pt, name, url, user, pass, token, ip, cookie, money, status, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				hid, pt, name, url, user, pass, token, ip, cookie, money, status, addtime)
			if err != nil {
				info.Failed++
			} else {
				info.Inserted++
			}
		}
	}
	return info, nil
}

// syncUsers 同步用户表
func (s *DBSyncService) syncUsers(extDB *sql.DB, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	rows, err := extDB.Query(`SELECT uid, COALESCE(uuid,0), COALESCE(user,''), COALESCE(pass,''),
		COALESCE(name,''), COALESCE(money,0), COALESCE(grade,'0'), COALESCE(active,'1'),
		COALESCE(addprice,1), COALESCE(` + "`key`" + `,''), COALESCE(yqm,''), COALESCE(yqprice,'0'),
		COALESCE(email,''), COALESCE(tuisongtoken,''), COALESCE(addtime,'')
		FROM qingka_wangke_user`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid, uuid int
		var user, pass, name, grade, active, key, yqm, yqprice, email, tuisongtoken, addtime string
		var money, addprice float64
		if err := rows.Scan(&uid, &uuid, &user, &pass, &name, &money, &grade, &active,
			&addprice, &key, &yqm, &yqprice, &email, &tuisongtoken, &addtime); err != nil {
			info.Failed++
			continue
		}
		info.Total++

		var exists int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&exists)

		if exists > 0 {
			if updateExisting {
				_, err := database.DB.Exec(
					"UPDATE qingka_wangke_user SET uuid=?, user=?, pass=?, name=?, money=?, grade=?, active=?, addprice=?, `key`=?, yqm=?, yqprice=?, email=?, tuisongtoken=?, addtime=? WHERE uid=?",
					uuid, user, pass, name, money, grade, active, addprice, key, yqm, yqprice, email, tuisongtoken, addtime, uid)
				if err != nil {
					info.Failed++
				} else {
					info.Updated++
				}
			} else {
				info.Skipped++
			}
		} else {
			_, err := database.DB.Exec(
				"INSERT INTO qingka_wangke_user (uid, uuid, user, pass, name, money, grade, active, addprice, `key`, yqm, yqprice, email, tuisongtoken, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				uid, uuid, user, pass, name, money, grade, active, addprice, key, yqm, yqprice, email, tuisongtoken, addtime)
			if err != nil {
				info.Failed++
			} else {
				info.Inserted++
			}
		}
	}
	return info, nil
}

// syncFenlei 同步分类表
func (s *DBSyncService) syncFenlei(extDB *sql.DB, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	rows, err := extDB.Query("SELECT id, COALESCE(sort,0), COALESCE(name,''), COALESCE(status,'1'), COALESCE(time,''), COALESCE(recommend,0) FROM qingka_wangke_fenlei")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, sort, recommend int
		var name, status, timeStr string
		if err := rows.Scan(&id, &sort, &name, &status, &timeStr, &recommend); err != nil {
			info.Failed++
			continue
		}
		info.Total++

		var exists int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE id = ?", id).Scan(&exists)

		if exists > 0 {
			if updateExisting {
				_, err := database.DB.Exec(
					"UPDATE qingka_wangke_fenlei SET sort=?, name=?, status=?, time=?, recommend=? WHERE id=?",
					sort, name, status, timeStr, recommend, id)
				if err != nil {
					info.Failed++
				} else {
					info.Updated++
				}
			} else {
				info.Skipped++
			}
		} else {
			_, err := database.DB.Exec(
				"INSERT INTO qingka_wangke_fenlei (id, sort, name, status, time, recommend) VALUES (?, ?, ?, ?, ?, ?)",
				id, sort, name, status, timeStr, recommend)
			if err != nil {
				info.Failed++
			} else {
				info.Inserted++
			}
		}
	}
	return info, nil
}

// syncClass 同步商品/课程表
func (s *DBSyncService) syncClass(extDB *sql.DB, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	rows, err := extDB.Query(`SELECT cid, COALESCE(name,''), COALESCE(noun,''), COALESCE(getnoun,''),
		COALESCE(docking,'0'), COALESCE(price,'0'), COALESCE(yunsuan,'*'),
		COALESCE(content,''), COALESCE(fenlei,'0'), COALESCE(status,0),
		COALESCE(addtime,''), COALESCE(sort,0)
		FROM qingka_wangke_class`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, status, sort int
		var name, noun, getnoun, docking, price, yunsuan, content, fenlei, addtime string
		if err := rows.Scan(&cid, &name, &noun, &getnoun, &docking, &price, &yunsuan, &content, &fenlei, &status, &addtime, &sort); err != nil {
			info.Failed++
			continue
		}
		info.Total++

		var exists int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE cid = ?", cid).Scan(&exists)

		if exists > 0 {
			if updateExisting {
				_, err := database.DB.Exec(
					"UPDATE qingka_wangke_class SET name=?, noun=?, getnoun=?, docking=?, price=?, yunsuan=?, content=?, fenlei=?, status=?, addtime=?, sort=? WHERE cid=?",
					name, noun, getnoun, docking, price, yunsuan, content, fenlei, status, addtime, sort, cid)
				if err != nil {
					info.Failed++
				} else {
					info.Updated++
				}
			} else {
				info.Skipped++
			}
		} else {
			_, err := database.DB.Exec(
				"INSERT INTO qingka_wangke_class (cid, name, noun, getnoun, docking, price, yunsuan, content, fenlei, status, addtime, sort) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				cid, name, noun, getnoun, docking, price, yunsuan, content, fenlei, status, addtime, sort)
			if err != nil {
				info.Failed++
			} else {
				info.Inserted++
			}
		}
	}
	return info, nil
}

// syncOrders 同步订单表
func (s *DBSyncService) syncOrders(extDB *sql.DB, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	rows, err := extDB.Query(`SELECT oid, COALESCE(uid,0), COALESCE(cid,0), COALESCE(hid,0),
		COALESCE(ptname,''), COALESCE(school,''), COALESCE(name,''),
		COALESCE(user,''), COALESCE(pass,''), COALESCE(kcid,''), COALESCE(kcname,''),
		COALESCE(fees,'0'), COALESCE(noun,''), COALESCE(addtime,''), COALESCE(ip,''),
		COALESCE(dockstatus,0), COALESCE(status,''), COALESCE(process,''),
		COALESCE(yid,'')
		FROM qingka_wangke_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oid, uid, cid, hid, dockstatus int
		var ptname, school, name, user, pass, kcid, kcname, fees, noun, addtime, ip, status, process, yid string
		if err := rows.Scan(&oid, &uid, &cid, &hid, &ptname, &school, &name, &user, &pass,
			&kcid, &kcname, &fees, &noun, &addtime, &ip, &dockstatus, &status, &process, &yid); err != nil {
			info.Failed++
			continue
		}
		info.Total++

		var exists int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&exists)

		if exists > 0 {
			if updateExisting {
				_, err := database.DB.Exec(
					`UPDATE qingka_wangke_order SET uid=?, cid=?, hid=?, ptname=?, school=?, name=?,
					user=?, pass=?, kcid=?, kcname=?, fees=?, noun=?, addtime=?, ip=?,
					dockstatus=?, status=?, process=?, yid=? WHERE oid=?`,
					uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, fees, noun,
					addtime, ip, dockstatus, status, process, yid, oid)
				if err != nil {
					info.Failed++
				} else {
					info.Updated++
				}
			} else {
				info.Skipped++
			}
		} else {
			_, err := database.DB.Exec(
				`INSERT INTO qingka_wangke_order (oid, uid, cid, hid, ptname, school, name,
				user, pass, kcid, kcname, fees, noun, addtime, ip, dockstatus, status, process, yid)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				oid, uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, fees, noun,
				addtime, ip, dockstatus, status, process, yid)
			if err != nil {
				info.Failed++
			} else {
				info.Inserted++
			}
		}
	}
	return info, nil
}
