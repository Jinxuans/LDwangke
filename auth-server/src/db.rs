use rusqlite::params;
use r2d2::{Pool, PooledConnection};
use r2d2_sqlite::SqliteConnectionManager;

use crate::config::AppConfig;
use crate::model::*;

pub type DbPool = Pool<SqliteConnectionManager>;

fn conn(db: &DbPool) -> PooledConnection<SqliteConnectionManager> {
    db.get().expect("获取数据库连接失败")
}

const SCHEMA: &str = r#"
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 2,
    display_name TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    max_licenses INTEGER NOT NULL DEFAULT 100,
    created_by INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS license (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    license_key TEXT NOT NULL UNIQUE,
    domain TEXT NOT NULL DEFAULT '*',
    machine_id TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    plan TEXT NOT NULL DEFAULT 'standard',
    max_users INTEGER NOT NULL DEFAULT 0,
    max_agents INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    expire_at TEXT,
    last_heartbeat TEXT,
    last_ip TEXT NOT NULL DEFAULT '',
    version TEXT NOT NULL DEFAULT '',
    bind_count INTEGER NOT NULL DEFAULT 0,
    max_bind INTEGER NOT NULL DEFAULT 3,
    dealer_id INTEGER NOT NULL DEFAULT 0,
    is_trial INTEGER NOT NULL DEFAULT 0,
    month_rebind_count INTEGER NOT NULL DEFAULT 0,
    last_rebind_month TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS license_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    license_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    ip TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS notice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    notice_type TEXT NOT NULL DEFAULT 'info',
    target TEXT NOT NULL DEFAULT '*',
    active INTEGER NOT NULL DEFAULT 1,
    created_by INTEGER,
    start_at TEXT,
    end_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE INDEX IF NOT EXISTS idx_license_key ON license(license_key);
CREATE INDEX IF NOT EXISTS idx_license_status ON license(status);
CREATE INDEX IF NOT EXISTS idx_license_dealer ON license(dealer_id);
CREATE INDEX IF NOT EXISTS idx_license_created ON license(created_at);
CREATE INDEX IF NOT EXISTS idx_log_license_id ON license_log(license_id);
CREATE INDEX IF NOT EXISTS idx_log_action ON license_log(action);
CREATE INDEX IF NOT EXISTS idx_log_created ON license_log(created_at);
CREATE INDEX IF NOT EXISTS idx_notice_active ON notice(active);
"#;

/// 初始化数据库连接池，建表
pub fn init(config: &AppConfig) -> DbPool {
    if let Some(parent) = std::path::Path::new(&config.database.path).parent() {
        std::fs::create_dir_all(parent).ok();
    }

    let manager = SqliteConnectionManager::file(&config.database.path)
        .with_init(|c| {
            c.execute_batch(
                "PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;"
            )
        });
    let pool = Pool::builder()
        .max_size(8)  // 允许最多8个并发连接（SQLite WAL模式支持并发读）
        .build(manager)
        .unwrap_or_else(|e| panic!("创建连接池失败: {}", e));

    // 用一个连接初始化建表
    {
        let c = pool.get().expect("获取初始化连接失败");
        c.execute_batch(SCHEMA).expect("建表失败");
        // 兼容旧库：给 license 表加新列（如果不存在）
        c.execute_batch("ALTER TABLE license ADD COLUMN dealer_id INTEGER NOT NULL DEFAULT 0;").ok();
        c.execute_batch("ALTER TABLE license ADD COLUMN is_trial INTEGER NOT NULL DEFAULT 0;").ok();
        c.execute_batch("ALTER TABLE license ADD COLUMN month_rebind_count INTEGER NOT NULL DEFAULT 0;").ok();
        c.execute_batch("ALTER TABLE license ADD COLUMN last_rebind_month TEXT NOT NULL DEFAULT '';").ok();
    }

    pool
}

// ===== 查询 =====

const SELECT_COLS: &str = "id,license_key,domain,machine_id,note,plan,max_users,max_agents,status,expire_at,last_heartbeat,last_ip,version,bind_count,max_bind,dealer_id,is_trial,month_rebind_count,last_rebind_month,created_at,updated_at";

fn map_license(row: &rusqlite::Row) -> rusqlite::Result<License> {
    Ok(License {
        id: row.get(0)?,
        license_key: row.get(1)?,
        domain: row.get(2)?,
        machine_id: row.get(3)?,
        note: row.get(4)?,
        plan: row.get(5)?,
        max_users: row.get(6)?,
        max_agents: row.get(7)?,
        status: row.get(8)?,
        expire_at: row.get(9)?,
        last_heartbeat: row.get(10)?,
        last_ip: row.get(11)?,
        version: row.get(12)?,
        bind_count: row.get(13)?,
        max_bind: row.get(14)?,
        dealer_id: row.get(15)?,
        is_trial: row.get(16)?,
        month_rebind_count: row.get(17)?,
        last_rebind_month: row.get(18)?,
        created_at: row.get(19)?,
        updated_at: row.get(20)?,
    })
}

// ===== 用户操作 =====

fn map_user(row: &rusqlite::Row) -> rusqlite::Result<User> {
    Ok(User {
        id: row.get(0)?,
        username: row.get(1)?,
        password_hash: row.get(2)?,
        role: row.get(3)?,
        display_name: row.get(4)?,
        status: row.get(5)?,
        max_licenses: row.get(6)?,
        created_by: row.get(7)?,
        created_at: row.get(8)?,
        updated_at: row.get(9)?,
    })
}

const USER_COLS: &str = "id,username,password_hash,role,display_name,status,max_licenses,created_by,created_at,updated_at";

pub fn get_user_by_username(db: &DbPool, username: &str) -> Option<User> {
    let c = conn(db);
    let sql = format!("SELECT {} FROM user WHERE username=?1", USER_COLS);
    c.query_row(&sql, params![username], map_user).ok()
}

pub fn get_user_by_id(db: &DbPool, id: i64) -> Option<User> {
    let c = conn(db);
    let sql = format!("SELECT {} FROM user WHERE id=?1", USER_COLS);
    c.query_row(&sql, params![id], map_user).ok()
}

pub fn create_user(db: &DbPool, username: &str, password_hash: &str, role: i32, display_name: &str, max_licenses: i32, created_by: Option<i64>) -> Result<i64, String> {
    let c = conn(db);
    c.execute(
        "INSERT INTO user (username, password_hash, role, display_name, max_licenses, created_by) VALUES (?1,?2,?3,?4,?5,?6)",
        params![username, password_hash, role, display_name, max_licenses, created_by],
    ).map_err(|e| {
        if e.to_string().contains("UNIQUE") { "用户名已存在".into() }
        else { format!("创建用户失败: {}", e) }
    })?;
    Ok(c.last_insert_rowid())
}

pub fn update_user(db: &DbPool, req: &UpdateUserRequest, password_hash: Option<&str>) -> Result<(), String> {
    let c = conn(db);
    let mut sets = Vec::new();
    let mut vals: Vec<Box<dyn rusqlite::types::ToSql>> = Vec::new();

    if let Some(ref v) = req.display_name {
        sets.push("display_name=?");
        vals.push(Box::new(v.clone()));
    }
    if let Some(ref h) = password_hash {
        sets.push("password_hash=?");
        vals.push(Box::new(h.to_string()));
    }
    if let Some(v) = req.role {
        sets.push("role=?");
        vals.push(Box::new(v));
    }
    if let Some(v) = req.status {
        sets.push("status=?");
        vals.push(Box::new(v));
    }
    if let Some(v) = req.max_licenses {
        sets.push("max_licenses=?");
        vals.push(Box::new(v));
    }

    if sets.is_empty() {
        return Err("无更新字段".into());
    }

    sets.push("updated_at=datetime('now','localtime')");
    vals.push(Box::new(req.id));

    let sql = format!("UPDATE user SET {} WHERE id=?", sets.join(","));
    let params: Vec<&dyn rusqlite::types::ToSql> = vals.iter().map(|v| v.as_ref()).collect();
    c.execute(&sql, params.as_slice()).map_err(|e| format!("更新失败: {}", e))?;
    Ok(())
}

pub fn delete_user(db: &DbPool, id: i64) -> Result<(), String> {
    let c = conn(db);
    // 不允许删除超管
    let role: i32 = c.query_row("SELECT role FROM user WHERE id=?1", params![id], |r| r.get(0))
        .map_err(|_| "用户不存在".to_string())?;
    if role == 0 { return Err("不能删除超级管理员".into()); }
    c.execute("DELETE FROM user WHERE id=?1", params![id])
        .map_err(|e| format!("删除失败: {}", e))?;
    Ok(())
}

pub fn list_users(db: &DbPool, page: i64, limit: i64) -> (Vec<User>, i64) {
    let c = conn(db);
    let offset = (page - 1) * limit;
    let total: i64 = c.query_row("SELECT COUNT(*) FROM user", [], |r| r.get(0)).unwrap_or(0);
    let sql = format!("SELECT {} FROM user ORDER BY id ASC LIMIT ?1 OFFSET ?2", USER_COLS);
    let mut stmt = c.prepare(&sql).unwrap();
    let list = stmt.query_map(params![limit, offset], map_user).unwrap().filter_map(|r| r.ok()).collect();
    (list, total)
}

pub fn count_dealer_licenses(db: &DbPool, dealer_id: i64) -> i64 {
    let c = conn(db);
    c.query_row("SELECT COUNT(*) FROM license WHERE dealer_id=?1", params![dealer_id], |r| r.get(0)).unwrap_or(0)
}

pub fn ensure_super_admin(db: &DbPool, username: &str, password_hash: &str) {
    let c = conn(db);
    let exists: bool = c.query_row(
        "SELECT COUNT(*) FROM user WHERE role=0", [], |r| r.get::<_,i64>(0)
    ).unwrap_or(0) > 0;
    if !exists {
        c.execute(
            "INSERT OR IGNORE INTO user (username, password_hash, role, display_name, max_licenses) VALUES (?1,?2,0,'超级管理员',0)",
            params![username, password_hash],
        ).ok();
        tracing::info!("已创建超级管理员账号: {}", username);
    }
}

pub fn get_by_key(db: &DbPool, key: &str) -> Option<License> {
    let c = conn(db);
    let sql = format!("SELECT {} FROM license WHERE license_key=?1", SELECT_COLS);
    c.query_row(&sql, params![key], map_license).ok()
}

pub fn get_by_id(db: &DbPool, id: i64) -> Option<License> {
    let c = conn(db);
    let sql = format!("SELECT {} FROM license WHERE id=?1", SELECT_COLS);
    c.query_row(&sql, params![id], map_license).ok()
}

/// 分页列表（支持关键词搜索和状态筛选）
pub fn list_licenses(db: &DbPool, page: i64, limit: i64, keyword: &str, status_filter: Option<i32>, dealer_id: Option<i64>) -> (Vec<License>, i64) {
    let c = conn(db);
    let offset = (page - 1) * limit;
    let like = format!("%{}%", keyword);
    let status_val = status_filter.unwrap_or(-1);
    let did = dealer_id.unwrap_or(-1);

    let where_clause = "WHERE (?1 = '' OR license_key LIKE ?2 OR note LIKE ?2 OR domain LIKE ?2) AND (?3 = -1 OR status = ?3) AND (?6 = -1 OR dealer_id = ?6)";

    let count_sql = format!("SELECT COUNT(*) FROM license {}", where_clause);
    let total: i64 = c
        .query_row(&count_sql, params![keyword, &like, status_val, 0, 0, did], |r| r.get(0))
        .unwrap_or(0);

    let query_sql = format!(
        "SELECT {} FROM license {} ORDER BY id DESC LIMIT ?4 OFFSET ?5",
        SELECT_COLS, where_clause
    );
    let mut stmt = c.prepare(&query_sql).unwrap();
    let list = stmt
        .query_map(params![keyword, &like, status_val, limit, offset, did], map_license)
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();

    (list, total)
}

// ===== 创建 =====

fn duration_days(days: i64) -> chrono::TimeDelta {
    chrono::TimeDelta::try_days(days).expect("invalid days")
}

pub fn create_license(db: &DbPool, key: &str, req: &CreateLicenseRequest) -> Result<i64, String> {
    let c = conn(db);
    let expire_at: Option<String> = if req.expire_days > 0 {
        Some(
            (chrono::Local::now() + duration_days(req.expire_days as i64))
                .format("%Y-%m-%d %H:%M:%S")
                .to_string(),
        )
    } else {
        None
    };

    c.execute(
        "INSERT INTO license (license_key, domain, note, plan, max_users, max_agents, max_bind, expire_at, dealer_id) VALUES (?1,?2,?3,?4,?5,?6,?7,?8,?9)",
        params![key, req.domain, req.note, req.plan, req.max_users, req.max_agents, req.max_bind, expire_at, req.dealer_id],
    )
    .map_err(|e| format!("创建失败: {}", e))?;

    Ok(c.last_insert_rowid())
}

// ===== 更新 =====

pub fn update_license(db: &DbPool, req: &UpdateLicenseRequest) -> Result<(), String> {
    let c = conn(db);
    let mut sets = Vec::new();
    let mut vals: Vec<Box<dyn rusqlite::types::ToSql>> = Vec::new();

    if let Some(ref v) = req.domain {
        sets.push("domain=?");
        vals.push(Box::new(v.clone()));
    }
    if let Some(ref v) = req.note {
        sets.push("note=?");
        vals.push(Box::new(v.clone()));
    }
    if let Some(ref v) = req.plan {
        sets.push("plan=?");
        vals.push(Box::new(v.clone()));
    }
    if let Some(v) = req.max_users {
        sets.push("max_users=?");
        vals.push(Box::new(v));
    }
    if let Some(v) = req.max_agents {
        sets.push("max_agents=?");
        vals.push(Box::new(v));
    }
    if let Some(v) = req.max_bind {
        sets.push("max_bind=?");
        vals.push(Box::new(v));
    }

    if sets.is_empty() {
        return Err("无更新字段".into());
    }

    sets.push("updated_at=datetime('now','localtime')");
    vals.push(Box::new(req.id));

    let sql = format!("UPDATE license SET {} WHERE id=?", sets.join(","));
    let params: Vec<&dyn rusqlite::types::ToSql> = vals.iter().map(|v| v.as_ref()).collect();
    c.execute(&sql, params.as_slice())
        .map_err(|e| format!("更新失败: {}", e))?;
    Ok(())
}

// ===== 状态变更 =====

pub fn set_status(db: &DbPool, id: i64, status: i32) {
    let c = conn(db);
    c.execute(
        "UPDATE license SET status=?1, updated_at=datetime('now','localtime') WHERE id=?2",
        params![status, id],
    )
    .ok();
}

// ===== 机器码绑定 =====

pub fn bind_machine(db: &DbPool, id: i64, machine_id: &str) -> Result<(), String> {
    let c = conn(db);
    c.execute(
        "UPDATE license SET machine_id=?1, bind_count=bind_count+1, updated_at=datetime('now','localtime') WHERE id=?2",
        params![machine_id, id],
    )
    .map_err(|e| format!("绑定失败: {}", e))?;
    Ok(())
}

pub fn unbind_machine(db: &DbPool, id: i64) -> Result<(), String> {
    let c = conn(db);
    // 读取当月换绑信息
    let (month_rebind_count, last_rebind_month): (i32, String) = c
        .query_row("SELECT month_rebind_count, last_rebind_month FROM license WHERE id=?1", params![id], |r| {
            Ok((r.get(0)?, r.get::<_, String>(1)?))
        })
        .map_err(|_| "授权码不存在".to_string())?;

    let current_month = chrono::Local::now().format("%Y-%m").to_string();
    let used = if last_rebind_month == current_month { month_rebind_count } else { 0 };

    if used >= 2 {
        return Err(format!("本月换绑次数已用完 ({}/2)", used));
    }

    let new_count = used + 1;
    c.execute(
        "UPDATE license SET machine_id='', month_rebind_count=?1, last_rebind_month=?2, bind_count=bind_count+1, updated_at=datetime('now','localtime') WHERE id=?3",
        params![new_count, &current_month, id],
    )
    .map_err(|e| format!("解绑失败: {}", e))?;
    Ok(())
}

// ===== 心跳更新 =====

pub fn update_heartbeat(db: &DbPool, id: i64, ip: &str, version: &str) {
    let c = conn(db);
    c.execute(
        "UPDATE license SET last_heartbeat=datetime('now','localtime'), last_ip=?1, version=?2, updated_at=datetime('now','localtime') WHERE id=?3",
        params![ip, version, id],
    )
    .ok();
}

// ===== 续期 =====

pub fn renew_license(db: &DbPool, id: i64, expire_days: i32) -> Result<String, String> {
    let c = conn(db);

    // 读取当前过期时间
    let current_expire: Option<String> = c
        .query_row("SELECT expire_at FROM license WHERE id=?1", params![id], |r| r.get(0))
        .map_err(|_| "授权码不存在".to_string())?;

    // 从当前过期时间或现在开始续期（已过期则从现在算起）
    let now = chrono::Local::now().naive_local();
    let base = if let Some(ref exp) = current_expire {
        let exp_time = chrono::NaiveDateTime::parse_from_str(exp, "%Y-%m-%d %H:%M:%S")
            .unwrap_or(now);
        if exp_time < now { now } else { exp_time }
    } else {
        now
    };

    let new_expire = (base + duration_days(expire_days as i64))
        .format("%Y-%m-%d %H:%M:%S")
        .to_string();

    c.execute(
        "UPDATE license SET expire_at=?1, status=1, updated_at=datetime('now','localtime') WHERE id=?2",
        params![&new_expire, id],
    )
    .map_err(|e| format!("续期失败: {}", e))?;

    Ok(new_expire)
}

// ===== 删除 =====

pub fn delete_license(db: &DbPool, id: i64) -> Result<(), String> {
    let c = conn(db);
    c.execute("DELETE FROM license_log WHERE license_id=?1", params![id]).ok();
    c.execute("DELETE FROM license WHERE id=?1", params![id])
        .map_err(|e| format!("删除失败: {}", e))?;
    Ok(())
}

// ===== 日志 =====

pub fn add_log(db: &DbPool, license_id: i64, action: &str, ip: &str, detail: &str) {
    let c = conn(db);
    c.execute(
        "INSERT INTO license_log (license_id, action, ip, detail) VALUES (?1,?2,?3,?4)",
        params![license_id, action, ip, detail],
    )
    .ok();
}

pub fn list_logs(db: &DbPool, page: i64, limit: i64, license_id: Option<i64>, action: Option<&str>, dealer_id: Option<i64>) -> (Vec<LicenseLog>, i64) {
    let c = conn(db);
    let offset = (page - 1) * limit;
    let lid = license_id.unwrap_or(-1);
    let act = action.unwrap_or("");
    let did = dealer_id.unwrap_or(-1);

    let where_clause = "WHERE (?1 = -1 OR l.license_id = ?1) AND (?2 = '' OR l.action = ?2) AND (?5 = -1 OR lc.dealer_id = ?5)";

    let count_sql = format!("SELECT COUNT(*) FROM license_log l LEFT JOIN license lc ON lc.id = l.license_id {}", where_clause);
    let total: i64 = c
        .query_row(&count_sql, params![lid, act, 0, 0, did], |r| r.get(0))
        .unwrap_or(0);

    let query_sql = format!(
        "SELECT l.id, l.license_id, l.action, l.ip, l.detail, l.created_at, lc.license_key, lc.note \
         FROM license_log l LEFT JOIN license lc ON lc.id = l.license_id {} ORDER BY l.id DESC LIMIT ?3 OFFSET ?4",
        where_clause
    );
    let mut stmt = c.prepare(&query_sql).unwrap();
    let list = stmt
        .query_map(params![lid, act, limit, offset, did], |row| {
            Ok(LicenseLog {
                id: row.get(0)?,
                license_id: row.get(1)?,
                action: row.get(2)?,
                ip: row.get(3)?,
                detail: row.get(4)?,
                created_at: row.get(5)?,
                license_key: row.get(6)?,
                note: row.get(7)?,
            })
        })
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();

    (list, total)
}

// ===== 统计看板 =====

pub fn dashboard(db: &DbPool, offline_threshold_secs: i64, dealer_id: Option<i64>) -> Dashboard {
    let c = conn(db);
    let did = dealer_id.unwrap_or(-1);
    let w = "(?1 = -1 OR dealer_id = ?1)";

    let total: i64 = c.query_row(&format!("SELECT COUNT(*) FROM license WHERE {}", w), params![did], |r| r.get(0)).unwrap_or(0);
    let active: i64 = c.query_row(&format!("SELECT COUNT(*) FROM license WHERE status=1 AND {}", w), params![did], |r| r.get(0)).unwrap_or(0);
    let expired: i64 = c.query_row(&format!("SELECT COUNT(*) FROM license WHERE status=2 AND {}", w), params![did], |r| r.get(0)).unwrap_or(0);
    let revoked: i64 = c.query_row(&format!("SELECT COUNT(*) FROM license WHERE status=0 AND {}", w), params![did], |r| r.get(0)).unwrap_or(0);

    let threshold = format!("-{} seconds", offline_threshold_secs);
    let online_now: i64 = c
        .query_row(
            &format!("SELECT COUNT(*) FROM license WHERE status=1 AND last_heartbeat IS NOT NULL AND last_heartbeat >= datetime('now','localtime', ?2) AND {}", w),
            params![did, threshold],
            |r| r.get(0),
        )
        .unwrap_or(0);

    Dashboard { total, active, expired, revoked, online_now }
}

// ===== 过期扫描 =====

pub fn expire_scan(db: &DbPool) -> Vec<String> {
    let c = conn(db);
    // 先查出即将被标记为过期的授权码 key，用于缓存失效
    let mut stmt = c.prepare(
        "SELECT license_key FROM license WHERE status=1 AND expire_at IS NOT NULL AND expire_at < datetime('now','localtime')"
    ).unwrap();
    let keys: Vec<String> = stmt.query_map([], |row| row.get(0))
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();

    // 释放读连接后用新连接写入（减少锁持有时间）
    drop(stmt);
    drop(c);

    if !keys.is_empty() {
        let c = conn(db);
        c.execute(
            "UPDATE license SET status=2, updated_at=datetime('now','localtime') WHERE status=1 AND expire_at IS NOT NULL AND expire_at < datetime('now','localtime')",
            [],
        ).ok();
    }
    keys
}

// ===== 日志清理 =====

pub fn cleanup_old_logs(db: &DbPool, retain_days: i64) -> usize {
    let c = conn(db);
    c.execute(
        "DELETE FROM license_log WHERE created_at < datetime('now','localtime', ?1)",
        params![format!("-{} days", retain_days)],
    ).unwrap_or(0)
}

// ===== 批量操作 =====

pub fn batch_create(db: &DbPool, keys: &[String], req: &crate::model::BatchCreateRequest) -> Result<Vec<(i64, String)>, String> {
    let c = conn(db);
    let expire_at: Option<String> = if req.expire_days > 0 {
        Some(
            (chrono::Local::now() + duration_days(req.expire_days as i64))
                .format("%Y-%m-%d %H:%M:%S")
                .to_string(),
        )
    } else {
        None
    };

    let tx = c.unchecked_transaction().map_err(|e| format!("事务开始失败: {}", e))?;
    let mut results = Vec::new();
    for key in keys {
        tx.execute(
            "INSERT INTO license (license_key, domain, note, plan, max_users, max_agents, max_bind, expire_at, dealer_id, is_trial) VALUES (?1,?2,?3,?4,?5,?6,?7,?8,?9,0)",
            params![key, req.domain, req.note, req.plan, req.max_users, req.max_agents, req.max_bind, expire_at, req.dealer_id],
        ).map_err(|e| format!("创建失败: {}", e))?;
        let id = tx.last_insert_rowid();
        results.push((id, key.clone()));
    }
    tx.commit().map_err(|e| format!("事务提交失败: {}", e))?;
    Ok(results)
}

pub fn batch_set_status(db: &DbPool, ids: &[i64], status: i32) -> Vec<String> {
    let c = conn(db);
    let mut keys = Vec::new();
    for &id in ids {
        if let Ok(key) = c.query_row("SELECT license_key FROM license WHERE id=?1", params![id], |r| r.get::<_, String>(0)) {
            keys.push(key);
        }
        c.execute(
            "UPDATE license SET status=?1, updated_at=datetime('now','localtime') WHERE id=?2",
            params![status, id],
        ).ok();
    }
    keys
}

pub fn batch_renew(db: &DbPool, ids: &[i64], expire_days: i32) -> Result<Vec<(i64, String)>, String> {
    let c = conn(db);
    let mut results = Vec::new();
    let now = chrono::Local::now().naive_local();
    for &id in ids {
        let current_expire: Option<String> = c
            .query_row("SELECT expire_at FROM license WHERE id=?1", params![id], |r| r.get(0))
            .map_err(|_| format!("授权码 {} 不存在", id))?;
        let base = if let Some(ref exp) = current_expire {
            let exp_time = chrono::NaiveDateTime::parse_from_str(exp, "%Y-%m-%d %H:%M:%S").unwrap_or(now);
            if exp_time < now { now } else { exp_time }
        } else { now };
        let new_expire = (base + duration_days(expire_days as i64)).format("%Y-%m-%d %H:%M:%S").to_string();
        c.execute(
            "UPDATE license SET expire_at=?1, status=1, updated_at=datetime('now','localtime') WHERE id=?2",
            params![&new_expire, id],
        ).map_err(|e| format!("续期失败: {}", e))?;
        results.push((id, new_expire));
    }
    Ok(results)
}

// ===== 试用授权 =====

pub fn check_trial_exists(db: &DbPool, machine_id: &str) -> bool {
    let c = conn(db);
    let count: i64 = c.query_row(
        "SELECT COUNT(*) FROM license WHERE is_trial=1 AND machine_id=?1",
        params![machine_id],
        |r| r.get(0),
    ).unwrap_or(0);
    count > 0
}

pub fn create_trial_license(db: &DbPool, key: &str, machine_id: &str, domain: &str, trial_days: i64) -> Result<i64, String> {
    let c = conn(db);
    let expire_at = (chrono::Local::now() + duration_days(trial_days))
        .format("%Y-%m-%d %H:%M:%S")
        .to_string();
    c.execute(
        "INSERT INTO license (license_key, domain, machine_id, note, plan, max_users, max_agents, max_bind, expire_at, is_trial, bind_count) VALUES (?1,?2,?3,'试用授权','trial',10,5,1,?4,1,1)",
        params![key, domain, machine_id, &expire_at],
    ).map_err(|e| format!("创建试用授权失败: {}", e))?;
    Ok(c.last_insert_rowid())
}

// ===== 公告管理 =====

const NOTICE_COLS: &str = "id,title,content,notice_type,target,active,created_by,start_at,end_at,created_at,updated_at";

fn map_notice(row: &rusqlite::Row) -> rusqlite::Result<Notice> {
    Ok(Notice {
        id: row.get(0)?,
        title: row.get(1)?,
        content: row.get(2)?,
        notice_type: row.get(3)?,
        target: row.get(4)?,
        active: row.get(5)?,
        created_by: row.get(6)?,
        start_at: row.get(7)?,
        end_at: row.get(8)?,
        created_at: row.get(9)?,
        updated_at: row.get(10)?,
    })
}

pub fn create_notice(db: &DbPool, req: &CreateNoticeRequest, created_by: i64) -> Result<i64, String> {
    let c = conn(db);
    c.execute(
        "INSERT INTO notice (title, content, notice_type, target, created_by, start_at, end_at) VALUES (?1,?2,?3,?4,?5,?6,?7)",
        params![req.title, req.content, req.notice_type, req.target, created_by, req.start_at, req.end_at],
    ).map_err(|e| format!("创建公告失败: {}", e))?;
    Ok(c.last_insert_rowid())
}

pub fn update_notice(db: &DbPool, req: &UpdateNoticeRequest) -> Result<(), String> {
    let c = conn(db);
    let mut sets = Vec::new();
    let mut vals: Vec<Box<dyn rusqlite::types::ToSql>> = Vec::new();

    if let Some(ref v) = req.title { sets.push("title=?"); vals.push(Box::new(v.clone())); }
    if let Some(ref v) = req.content { sets.push("content=?"); vals.push(Box::new(v.clone())); }
    if let Some(ref v) = req.notice_type { sets.push("notice_type=?"); vals.push(Box::new(v.clone())); }
    if let Some(ref v) = req.target { sets.push("target=?"); vals.push(Box::new(v.clone())); }
    if let Some(v) = req.active { sets.push("active=?"); vals.push(Box::new(v)); }
    if let Some(ref v) = req.start_at { sets.push("start_at=?"); vals.push(Box::new(v.clone())); }
    if let Some(ref v) = req.end_at { sets.push("end_at=?"); vals.push(Box::new(v.clone())); }

    if sets.is_empty() { return Err("无更新字段".into()); }
    sets.push("updated_at=datetime('now','localtime')");
    vals.push(Box::new(req.id));

    let sql = format!("UPDATE notice SET {} WHERE id=?", sets.join(","));
    let params: Vec<&dyn rusqlite::types::ToSql> = vals.iter().map(|v| v.as_ref()).collect();
    c.execute(&sql, params.as_slice()).map_err(|e| format!("更新失败: {}", e))?;
    Ok(())
}

pub fn delete_notice(db: &DbPool, id: i64) -> Result<(), String> {
    let c = conn(db);
    c.execute("DELETE FROM notice WHERE id=?1", params![id])
        .map_err(|e| format!("删除失败: {}", e))?;
    Ok(())
}

pub fn list_notices(db: &DbPool, page: i64, limit: i64) -> (Vec<Notice>, i64) {
    let c = conn(db);
    let offset = (page - 1) * limit;
    let total: i64 = c.query_row("SELECT COUNT(*) FROM notice", [], |r| r.get(0)).unwrap_or(0);
    let sql = format!("SELECT {} FROM notice ORDER BY id DESC LIMIT ?1 OFFSET ?2", NOTICE_COLS);
    let mut stmt = c.prepare(&sql).unwrap();
    let list = stmt.query_map(params![limit, offset], map_notice).unwrap().filter_map(|r| r.ok()).collect();
    (list, total)
}

pub fn get_active_notices(db: &DbPool, plan: &str) -> Vec<NoticePublic> {
    let c = conn(db);
    let sql = "SELECT id, title, content, notice_type FROM notice WHERE active=1 \
               AND (start_at IS NULL OR start_at <= datetime('now','localtime')) \
               AND (end_at IS NULL OR end_at >= datetime('now','localtime')) \
               AND (target='*' OR target=?1) \
               ORDER BY id DESC LIMIT 10";
    let mut stmt = c.prepare(sql).unwrap();
    let target = format!("plan:{}", plan);
    stmt.query_map(params![target], |row| {
        Ok(NoticePublic {
            id: row.get(0)?,
            title: row.get(1)?,
            content: row.get(2)?,
            notice_type: row.get(3)?,
        })
    }).unwrap().filter_map(|r| r.ok()).collect()
}

// ===== 统计报表 =====

pub fn trend_stats(db: &DbPool, days: i64, offline_threshold_secs: i64, dealer_id: Option<i64>) -> Vec<TrendItem> {
    let c = conn(db);
    let did = dealer_id.unwrap_or(-1);
    let threshold = format!("-{} seconds", offline_threshold_secs);
    let sql = format!(
        "WITH RECURSIVE dates(d) AS (
            SELECT date('now','localtime','-{} days')
            UNION ALL
            SELECT date(d, '+1 day') FROM dates WHERE d < date('now','localtime')
        )
        SELECT dates.d,
            (SELECT COUNT(*) FROM license WHERE date(created_at) = dates.d AND (?1 = -1 OR dealer_id = ?1)),
            (SELECT COUNT(*) FROM license WHERE status=2 AND expire_at IS NOT NULL AND date(expire_at) = dates.d AND (?1 = -1 OR dealer_id = ?1)),
            (SELECT COUNT(*) FROM license WHERE status=1 AND last_heartbeat IS NOT NULL AND last_heartbeat >= datetime(dates.d || ' 00:00:00', ?2) AND (?1 = -1 OR dealer_id = ?1))
        FROM dates ORDER BY dates.d",
        days
    );
    let mut stmt = c.prepare(&sql).unwrap();
    stmt.query_map(params![did, threshold], |row| {
        Ok(TrendItem {
            day: row.get(0)?,
            created: row.get(1)?,
            expired: row.get(2)?,
            online: row.get(3)?,
        })
    }).unwrap().filter_map(|r| r.ok()).collect()
}

pub fn plan_distribution(db: &DbPool, dealer_id: Option<i64>) -> Vec<PlanDistribution> {
    let c = conn(db);
    let did = dealer_id.unwrap_or(-1);
    let sql = "SELECT plan, COUNT(*) as cnt FROM license WHERE (?1 = -1 OR dealer_id = ?1) GROUP BY plan ORDER BY cnt DESC";
    let mut stmt = c.prepare(sql).unwrap();
    stmt.query_map(params![did], |row| {
        Ok(PlanDistribution {
            plan: row.get(0)?,
            count: row.get(1)?,
        })
    }).unwrap().filter_map(|r| r.ok()).collect()
}

// ===== 导出 =====

pub fn export_licenses(db: &DbPool, status_filter: Option<i32>, dealer_id: Option<i64>) -> Vec<License> {
    let c = conn(db);
    let status_val = status_filter.unwrap_or(-1);
    let did = dealer_id.unwrap_or(-1);
    let sql = format!(
        "SELECT {} FROM license WHERE (?1 = -1 OR status = ?1) AND (?2 = -1 OR dealer_id = ?2) ORDER BY id DESC",
        SELECT_COLS
    );
    let mut stmt = c.prepare(&sql).unwrap();
    stmt.query_map(params![status_val, did], map_license).unwrap().filter_map(|r| r.ok()).collect()
}

// ===== 从 JSON 迁移 =====

pub fn migrate_from_json(db: &DbPool, json_path: &str) -> Result<usize, String> {
    let content = std::fs::read_to_string(json_path)
        .map_err(|e| format!("读取文件失败: {}", e))?;
    let data: serde_json::Value =
        serde_json::from_str(&content).map_err(|e| format!("解析 JSON 失败: {}", e))?;

    let obj = data.as_object().ok_or("JSON 格式错误")?;
    let c = conn(db);
    let mut count = 0;

    for (key, info) in obj {
        let domain = info["domain"].as_str().unwrap_or("*");
        let note = info["note"].as_str().unwrap_or("");
        let active = info["active"].as_bool().unwrap_or(false);
        let machine_id = info["machine_id"].as_str().unwrap_or("");
        let created = info["created"].as_str().unwrap_or("");
        let status: i32 = if active { 1 } else { 0 };
        let bind_count: i32 = if machine_id.is_empty() { 0 } else { 1 };

        c.execute(
            "INSERT OR IGNORE INTO license (license_key, domain, machine_id, note, status, bind_count, created_at, updated_at) VALUES (?1,?2,?3,?4,?5,?6,?7,?7)",
            params![key, domain, machine_id, note, status, bind_count, created],
        )
        .ok();
        count += 1;
    }

    Ok(count)
}
