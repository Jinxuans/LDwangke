use std::net::SocketAddr;
use std::sync::Arc;

use axum::response::Html;
use axum::routing::{delete, get, post};
use axum::{middleware, Router};
use tower_http::cors::CorsLayer;
use tower_http::trace::TraceLayer;

mod auth;
mod cache;
mod config;
mod db;
mod handler;
mod model;

#[tokio::main]
async fn main() {
    // 初始化日志
    tracing_subscriber::fmt()
        .with_env_filter("license_server=info,tower_http=info")
        .init();

    // 加载配置
    let cfg = config::AppConfig::load("config.toml");
    let addr = format!("{}:{}", cfg.server.host, cfg.server.port);

    // 初始化数据库
    let db_pool = db::init(&cfg);

    // 从 JSON 迁移（如果文件存在）
    if std::path::Path::new("licenses_db.json").exists() {
        match db::migrate_from_json(&db_pool, "licenses_db.json") {
            Ok(n) => {
                tracing::info!("从 JSON 迁移了 {} 条授权记录", n);
                std::fs::rename("licenses_db.json", "licenses_db.json.bak").ok();
            }
            Err(e) => tracing::error!("JSON 迁移失败: {}", e),
        }
    }

    // 初始化缓存
    let verify_cache = cache::VerifyCache::new(cfg.security.verify_cache_secs);

    // 自动创建超级管理员
    let sa_hash = auth::hash_password(&cfg.super_admin.password)
        .expect("密码哈希失败");
    db::ensure_super_admin(&db_pool, &cfg.super_admin.username, &sa_hash);

    // 初始化速率限制器（每个 IP 每分钟最多 30 次请求）
    let rate_limiter = cache::RateLimiter::new(30, 60);

    // 构建应用状态
    let state = Arc::new(handler::AppState {
        db: db_pool.clone(),
        config: cfg.clone(),
        cache: verify_cache,
        rate_limiter,
    });

    // 管理路由（JWT 认证）
    let admin_routes = Router::new()
        .route("/licenses", get(handler::admin_list))
        .route("/license/create", post(handler::admin_create))
        .route("/license/update", post(handler::admin_update))
        .route("/license/revoke", post(handler::admin_revoke))
        .route("/license/enable", post(handler::admin_enable))
        .route("/license/unbind", post(handler::admin_unbind))
        .route("/license/renew", post(handler::admin_renew))
        .route("/license/delete/:id", delete(handler::admin_delete))
        .route("/license/logs", get(handler::admin_logs))
        .route("/license/dashboard", get(handler::admin_dashboard))
        .route("/users", get(handler::admin_user_list))
        .route("/user/create", post(handler::admin_user_create))
        .route("/user/update", post(handler::admin_user_update))
        .route("/user/delete/:id", delete(handler::admin_user_delete))
        .route("/me", get(handler::admin_me))
        .route_layer(middleware::from_fn_with_state(
            state.clone(),
            handler::jwt_auth_middleware,
        ));

    // 应用路由
    let app = Router::new()
        .route("/admin", get(admin_page))
        .route("/api/v1/auth/login", post(handler::login))
        .route("/api/v1/license/verify", post(handler::verify))
        .route("/api/v1/license/heartbeat", post(handler::heartbeat))
        .nest("/api/v1/admin", admin_routes)
        .layer(CorsLayer::permissive())
        .layer(TraceLayer::new_for_http())
        .with_state(state.clone());

    // 启动过期扫描定时任务（每小时）
    let scan_db = db_pool.clone();
    let scan_state = state.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(3600));
        loop {
            interval.tick().await;
            let db = scan_db.clone();
            let expired_keys = tokio::task::spawn_blocking(move || db::expire_scan(&db))
                .await
                .unwrap_or_default();
            if !expired_keys.is_empty() {
                tracing::info!("过期扫描：标记了 {} 个过期授权码", expired_keys.len());
                for key in &expired_keys {
                    scan_state.cache.invalidate(key);
                }
            }
            // 清理 90 天前的旧日志
            let db2 = scan_db.clone();
            let deleted = tokio::task::spawn_blocking(move || db::cleanup_old_logs(&db2, 90))
                .await
                .unwrap_or(0);
            if deleted > 0 {
                tracing::info!("日志清理：删除了 {} 条 90 天前的日志", deleted);
            }
            // 清理速率限制器过期条目
            scan_state.rate_limiter.cleanup();
            // 清理验证缓存过期条目（防止内存泄漏）
            scan_state.cache.cleanup();
        }
    });

    // 启动服务
    let listener = tokio::net::TcpListener::bind(&addr)
        .await
        .unwrap_or_else(|e| panic!("绑定 {} 失败: {}", addr, e));
    tracing::info!("授权站启动于 {}", addr);
    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
    .await
    .unwrap();
}

async fn admin_page() -> Html<&'static str> {
    Html(include_str!("../static/admin.html"))
}
