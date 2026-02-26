use serde::{Deserialize, Serialize};

// ===== 用户 =====

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: i64,
    pub username: String,
    #[serde(skip_serializing)]
    pub password_hash: String,
    pub role: i32,           // 0=超管, 1=授权商, 2=普通用户
    pub display_name: String,
    pub status: i32,         // 1=启用, 0=禁用
    pub max_licenses: i32,   // 授权商配额
    pub created_by: Option<i64>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Deserialize)]
pub struct LoginRequest {
    pub username: String,
    pub password: String,
}

#[derive(Debug, Deserialize)]
pub struct RegisterRequest {
    pub username: String,
    pub password: String,
    #[serde(default)]
    pub display_name: String,
}

#[derive(Debug, Deserialize)]
pub struct ChangePasswordRequest {
    pub old_password: String,
    pub new_password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub sub: i64,      // user id
    pub role: i32,     // 角色
    pub username: String,
    pub exp: usize,    // 过期时间
}

#[derive(Debug, Deserialize)]
pub struct CreateUserRequest {
    pub username: String,
    pub password: String,
    #[serde(default = "default_role")]
    pub role: i32,
    #[serde(default)]
    pub display_name: String,
    #[serde(default = "default_max_licenses")]
    pub max_licenses: i32,
}

fn default_role() -> i32 { 1 }
fn default_max_licenses() -> i32 { 100 }

#[derive(Debug, Deserialize)]
pub struct UpdateUserRequest {
    pub id: i64,
    pub display_name: Option<String>,
    pub password: Option<String>,
    pub role: Option<i32>,
    pub status: Option<i32>,
    pub max_licenses: Option<i32>,
}

// ===== 授权码 =====

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct License {
    pub id: i64,
    pub license_key: String,
    pub domain: String,
    pub machine_id: String,
    pub note: String,
    pub plan: String,
    pub max_users: i32,
    pub max_agents: i32,
    pub status: i32, // 1=正常 0=禁用 2=已过期
    pub expire_at: Option<String>,
    pub last_heartbeat: Option<String>,
    pub last_ip: String,
    pub version: String,
    pub bind_count: i32,
    pub max_bind: i32,
    pub dealer_id: i64,
    pub is_trial: i32,
    pub month_rebind_count: i32,
    pub last_rebind_month: String,
    pub created_at: String,
    pub updated_at: String,
}

// ===== 操作日志 =====

#[derive(Debug, Clone, Serialize)]
pub struct LicenseLog {
    pub id: i64,
    pub license_id: i64,
    pub action: String,
    pub ip: String,
    pub detail: String,
    pub created_at: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub license_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub note: Option<String>,
}

// ===== 公开接口请求 =====

#[derive(Debug, Deserialize)]
pub struct VerifyRequest {
    pub license_key: String,
    pub domain: String,
    pub machine_id: String,
    #[serde(default)]
    pub version: String,
    pub timestamp: i64,
    pub sign: String,
}

#[derive(Debug, Deserialize)]
pub struct HeartbeatRequest {
    pub license_key: String,
    pub machine_id: String,
    #[serde(default)]
    pub domain: String,
    #[serde(default)]
    pub version: String,
    pub timestamp: i64,
    pub sign: String,
    #[serde(default)]
    pub stats: Option<serde_json::Value>,
}

// ===== 公开接口响应 =====

#[derive(Debug, Clone, Serialize)]
pub struct VerifyResponse {
    pub valid: bool,
    pub plan: String,
    pub expire_at: Option<String>,
    pub max_users: i32,
    pub max_agents: i32,
    pub is_trial: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub message: Option<String>,
}

// ===== 管理接口请求 =====

#[derive(Debug, Deserialize)]
pub struct CreateLicenseRequest {
    #[serde(default = "default_domain")]
    pub domain: String,
    #[serde(default)]
    pub note: String,
    #[serde(default = "default_plan")]
    pub plan: String,
    #[serde(default)]
    pub max_users: i32,
    #[serde(default)]
    pub max_agents: i32,
    #[serde(default = "default_max_bind")]
    pub max_bind: i32,
    #[serde(default)]
    pub expire_days: i32, // 0=永久
    #[serde(default)]
    pub dealer_id: i64,   // 由中间件注入
}

fn default_domain() -> String { "*".into() }
fn default_plan() -> String { "standard".into() }
fn default_max_bind() -> i32 { 3 }

#[derive(Debug, Deserialize)]
pub struct UpdateLicenseRequest {
    pub id: i64,
    pub domain: Option<String>,
    pub note: Option<String>,
    pub plan: Option<String>,
    pub max_users: Option<i32>,
    pub max_agents: Option<i32>,
    pub max_bind: Option<i32>,
}

#[derive(Debug, Deserialize)]
pub struct RenewRequest {
    pub id: i64,
    pub expire_days: i32,
}

#[derive(Debug, Deserialize)]
pub struct IdRequest {
    pub id: i64,
}

#[derive(Debug, Deserialize)]
pub struct IdsRequest {
    pub ids: Vec<i64>,
}

#[derive(Debug, Deserialize)]
pub struct BatchCreateRequest {
    #[serde(default = "default_batch_count")]
    pub count: i32,
    #[serde(default = "default_domain")]
    pub domain: String,
    #[serde(default)]
    pub note: String,
    #[serde(default = "default_plan")]
    pub plan: String,
    #[serde(default)]
    pub max_users: i32,
    #[serde(default)]
    pub max_agents: i32,
    #[serde(default = "default_max_bind")]
    pub max_bind: i32,
    #[serde(default)]
    pub expire_days: i32,
    #[serde(default)]
    pub dealer_id: i64,
}

fn default_batch_count() -> i32 { 1 }

#[derive(Debug, Deserialize)]
pub struct BatchRenewRequest {
    pub ids: Vec<i64>,
    pub expire_days: i32,
}

#[derive(Debug, Deserialize)]
pub struct TrialRequest {
    pub machine_id: String,
    #[serde(default)]
    pub domain: String,
    pub timestamp: i64,
    pub sign: String,
}

// ===== 统计看板 =====

#[derive(Debug, Serialize)]
pub struct Dashboard {
    pub total: i64,
    pub active: i64,
    pub expired: i64,
    pub revoked: i64,
    pub online_now: i64,
}

#[derive(Debug, Serialize)]
pub struct TrendItem {
    pub day: String,
    pub created: i64,
    pub expired: i64,
    pub online: i64,
}

#[derive(Debug, Serialize)]
pub struct PlanDistribution {
    pub plan: String,
    pub count: i64,
}

// ===== 公告 =====

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notice {
    pub id: i64,
    pub title: String,
    pub content: String,
    pub notice_type: String,
    pub target: String,
    pub active: i32,
    pub created_by: Option<i64>,
    pub start_at: Option<String>,
    pub end_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize)]
pub struct NoticePublic {
    pub id: i64,
    pub title: String,
    pub content: String,
    #[serde(rename = "type")]
    pub notice_type: String,
}

#[derive(Debug, Deserialize)]
pub struct CreateNoticeRequest {
    pub title: String,
    #[serde(default)]
    pub content: String,
    #[serde(default = "default_notice_type")]
    pub notice_type: String,
    #[serde(default = "default_target")]
    pub target: String,
    pub start_at: Option<String>,
    pub end_at: Option<String>,
}

fn default_notice_type() -> String { "info".into() }
fn default_target() -> String { "*".into() }

#[derive(Debug, Deserialize)]
pub struct UpdateNoticeRequest {
    pub id: i64,
    pub title: Option<String>,
    pub content: Option<String>,
    pub notice_type: Option<String>,
    pub target: Option<String>,
    pub active: Option<i32>,
    pub start_at: Option<String>,
    pub end_at: Option<String>,
}

// ===== 分页查询参数 =====

#[derive(Debug, Deserialize)]
pub struct PageQuery {
    #[serde(default = "default_page")]
    pub page: i64,
    #[serde(default = "default_limit")]
    pub limit: i64,
    pub keyword: Option<String>,
    pub status: Option<i32>,
}

fn default_page() -> i64 { 1 }
fn default_limit() -> i64 { 20 }

#[derive(Debug, Deserialize)]
pub struct LogQuery {
    #[serde(default = "default_page")]
    pub page: i64,
    #[serde(default = "default_limit")]
    pub limit: i64,
    pub license_id: Option<i64>,
    pub action: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct TrendQuery {
    #[serde(default = "default_trend_days")]
    pub days: i64,
}

fn default_trend_days() -> i64 { 30 }

#[derive(Debug, Deserialize)]
pub struct ExportQuery {
    pub status: Option<i32>,
}

// ===== 通用 API 响应 =====

#[derive(Debug, Serialize)]
pub struct ApiResponse {
    pub code: i32,
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<serde_json::Value>,
    pub timestamp: i64,
}

impl ApiResponse {
    pub fn success(data: impl Serialize) -> Self {
        Self {
            code: 0,
            message: "success".into(),
            data: Some(serde_json::to_value(data).unwrap_or_default()),
            timestamp: chrono::Utc::now().timestamp(),
        }
    }

    pub fn success_msg(msg: &str) -> Self {
        Self {
            code: 0,
            message: msg.into(),
            data: None,
            timestamp: chrono::Utc::now().timestamp(),
        }
    }

    pub fn error(code: i32, msg: &str) -> Self {
        Self {
            code,
            message: msg.into(),
            data: None,
            timestamp: chrono::Utc::now().timestamp(),
        }
    }
}
