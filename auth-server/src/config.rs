use serde::Deserialize;
use std::fs;

#[derive(Debug, Clone, Deserialize)]
pub struct AppConfig {
    pub server: ServerConfig,
    pub security: SecurityConfig,
    pub database: DatabaseConfig,
    #[serde(default = "default_super_admin")]
    pub super_admin: SuperAdminConfig,
}

fn default_super_admin() -> SuperAdminConfig {
    SuperAdminConfig {
        username: default_sa_user(),
        password: default_sa_pass(),
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub host: String,
    pub port: u16,
    #[serde(default = "default_jwt_secret")]
    pub jwt_secret: String,
    #[serde(default = "default_jwt_expire")]
    pub jwt_expire_hours: u64,
}

fn default_jwt_secret() -> String { "license-server-jwt-secret-change-me".into() }
fn default_jwt_expire() -> u64 { 72 }

#[derive(Debug, Clone, Deserialize)]
pub struct SuperAdminConfig {
    #[serde(default = "default_sa_user")]
    pub username: String,
    #[serde(default = "default_sa_pass")]
    pub password: String,
}

fn default_sa_user() -> String { "admin".into() }
fn default_sa_pass() -> String { "admin123".into() }

#[derive(Debug, Clone, Deserialize)]
pub struct SecurityConfig {
    pub client_secret: String,
    pub sign_window_secs: i64,
    pub verify_cache_secs: u64,
    pub offline_threshold_secs: i64,
    pub max_default_bind: i32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DatabaseConfig {
    pub path: String,
}

impl AppConfig {
    pub fn load(path: &str) -> Self {
        let content = fs::read_to_string(path)
            .unwrap_or_else(|_| panic!("无法读取配置文件: {}", path));
        toml::from_str(&content)
            .unwrap_or_else(|e| panic!("解析配置文件失败: {}", e))
    }
}
