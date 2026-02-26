use std::net::SocketAddr;
use std::sync::Arc;

use axum::extract::{ConnectInfo, Extension, Json, Path, Query, Request, State};
use axum::http::HeaderMap;
use axum::middleware::Next;
use axum::response::{IntoResponse, Response};

use crate::auth;
use crate::cache::{RateLimiter, VerifyCache};
use crate::config::AppConfig;
use crate::db::{self, DbPool};
use crate::model::*;

// ===== 应用状态 =====

pub struct AppState {
    pub db: DbPool,
    pub config: AppConfig,
    pub cache: VerifyCache,
    pub rate_limiter: RateLimiter,
}

// ===== 工具函数 =====

fn get_client_ip(headers: &HeaderMap, addr: &SocketAddr) -> String {
    headers
        .get("x-forwarded-for")
        .or_else(|| headers.get("x-real-ip"))
        .and_then(|v| v.to_str().ok())
        .map(|s| s.split(',').next().unwrap_or("").trim().to_string())
        .unwrap_or_else(|| addr.ip().to_string())
}

// ===== 公开接口 =====

/// POST /api/v1/license/verify
/// 客户端启动时调用，验证授权码有效性，首次自动绑定机器码
pub async fn verify(
    State(state): State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<VerifyRequest>,
) -> Json<ApiResponse> {
    // 0. 速率限制
    let ip = get_client_ip(&headers, &addr);
    if !state.rate_limiter.check(&ip) {
        return Json(ApiResponse::error(429, "请求过于频繁，请稍后重试"));
    }

    // 1. 验证签名
    if let Err(msg) = auth::verify_request_sign(
        &state.config.security.client_secret,
        &req,
        state.config.security.sign_window_secs,
    ) {
        return Json(ApiResponse::error(403, msg));
    }

    // 2. 检查缓存
    if let Some(cached) = state.cache.get(&req.license_key) {
        return Json(ApiResponse::success(cached));
    }

    // 3. 数据库验证（阻塞操作放到 spawn_blocking）
    let db = state.db.clone();
    let key = req.license_key.clone();
    let domain = req.domain.clone();
    let machine_id = req.machine_id.clone();
    let version = req.version.clone();
    let ip = ip.clone();

    let (resp, cache_resp) = tokio::task::spawn_blocking(move || {
        let license = match db::get_by_key(&db, &key) {
            Some(l) => l,
            None => {
                return (ApiResponse::error(404, "授权码不存在"), None);
            }
        };

        // 状态检查
        if license.status == 0 {
            db::add_log(&db, license.id, "verify_fail", &ip, "授权码已被禁用");
            return (ApiResponse::error(403, "授权码已被禁用"), None);
        }
        if license.status == 2 {
            db::add_log(&db, license.id, "verify_fail", &ip, "授权码已过期");
            return (ApiResponse::error(403, "授权码已过期"), None);
        }

        // 过期检查（实时，使用时间类型比较）
        if let Some(ref expire) = license.expire_at {
            let expired = chrono::NaiveDateTime::parse_from_str(expire, "%Y-%m-%d %H:%M:%S")
                .map(|exp_time| chrono::Local::now().naive_local() > exp_time)
                .unwrap_or(false);
            if expired {
                db::set_status(&db, license.id, 2);
                db::add_log(&db, license.id, "expire", &ip, "验证时发现已过期");
                return (ApiResponse::error(403, "授权码已过期"), None);
            }
        }

        // 域名检查
        if license.domain != "*" && license.domain != domain {
            db::add_log(
                &db,
                license.id,
                "verify_fail",
                &ip,
                &format!("域名不匹配: {} != {}", domain, license.domain),
            );
            return (ApiResponse::error(403, "域名不匹配"), None);
        }

        // 机器码绑定检查
        if license.machine_id.is_empty() {
            // 首次激活，自动绑定
            if let Err(e) = db::bind_machine(&db, license.id, &machine_id) {
                return (ApiResponse::error(500, &e), None);
            }
            db::add_log(
                &db,
                license.id,
                "bind",
                &ip,
                &format!("首次激活绑定: {}", &machine_id),
            );
        } else if license.machine_id != machine_id {
            db::add_log(
                &db,
                license.id,
                "verify_fail",
                &ip,
                &format!("机器码不匹配: {} != {}", machine_id, license.machine_id),
            );
            return (
                ApiResponse::error(403, "机器码不匹配，请联系管理员解绑"),
                None,
            );
        }

        // 更新心跳信息
        db::update_heartbeat(&db, license.id, &ip, &version);
        db::add_log(&db, license.id, "verify", &ip, "验证通过");

        let verify_resp = VerifyResponse {
            valid: true,
            plan: license.plan.clone(),
            expire_at: license.expire_at,
            max_users: license.max_users,
            max_agents: license.max_agents,
            is_trial: license.is_trial == 1,
            message: None,
        };

        let notices = db::get_active_notices(&db, &license.plan);

        (ApiResponse::success(serde_json::json!({
            "valid": verify_resp.valid,
            "plan": verify_resp.plan,
            "expire_at": verify_resp.expire_at,
            "max_users": verify_resp.max_users,
            "max_agents": verify_resp.max_agents,
            "is_trial": verify_resp.is_trial,
            "notices": notices,
        })), Some(verify_resp))
    })
    .await
    .unwrap();

    // 4. 写入缓存
    if let Some(cr) = cache_resp {
        state.cache.set(req.license_key, cr);
    }

    Json(resp)
}

/// POST /api/v1/license/heartbeat
/// 客户端定时调用，上报运行状态
pub async fn heartbeat(
    State(state): State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<HeartbeatRequest>,
) -> Json<ApiResponse> {
    // 速率限制
    let ip = get_client_ip(&headers, &addr);
    if !state.rate_limiter.check(&ip) {
        return Json(ApiResponse::error(429, "请求过于频繁，请稍后重试"));
    }

    // 验证签名
    if let Err(msg) = auth::verify_heartbeat_sign(
        &state.config.security.client_secret,
        &req,
        state.config.security.sign_window_secs,
    ) {
        return Json(ApiResponse::error(403, msg));
    }

    let db = state.db.clone();
    let key = req.license_key.clone();
    let machine_id = req.machine_id.clone();
    let domain = req.domain.clone();
    let version = req.version.clone();
    let ip = ip.clone();
    let stats = req.stats.clone();

    let resp = tokio::task::spawn_blocking(move || {
        let license = match db::get_by_key(&db, &key) {
            Some(l) => l,
            None => return ApiResponse::error(404, "授权码不存在"),
        };

        if license.status != 1 {
            return ApiResponse::error(403, "授权码无效");
        }

        if !license.machine_id.is_empty() && license.machine_id != machine_id {
            return ApiResponse::error(403, "机器码不匹配");
        }

        // 域名校验（非通配符时检查）
        if !domain.is_empty() && license.domain != "*" && license.domain != domain {
            return ApiResponse::error(403, "域名不匹配");
        }

        db::update_heartbeat(&db, license.id, &ip, &version);

        // 心跳不逐条写日志（性能），仅在有 stats 时记录
        if let Some(ref s) = stats {
            if !s.is_null() {
                db::add_log(
                    &db,
                    license.id,
                    "heartbeat",
                    &ip,
                    &serde_json::to_string(s).unwrap_or_default(),
                );
            }
        }

        let notices = db::get_active_notices(&db, &license.plan);
        ApiResponse::success(serde_json::json!({
            "status": "ok",
            "notices": notices,
        }))
    })
    .await
    .unwrap();

    Json(resp)
}

// ===== 登录接口 =====

pub async fn login(
    State(state): State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<LoginRequest>,
) -> Json<ApiResponse> {
    // 登录接口速率限制（防暴力破解）
    let ip = get_client_ip(&headers, &addr);
    if !state.rate_limiter.check(&ip) {
        return Json(ApiResponse::error(429, "请求过于频繁，请稍后重试"));
    }

    let db = state.db.clone();
    let username = req.username.clone();
    let password = req.password.clone();
    let jwt_secret = state.config.server.jwt_secret.clone();
    let jwt_hours = state.config.server.jwt_expire_hours;

    let default_sa_pass = state.config.super_admin.password.clone();

    let result = tokio::task::spawn_blocking(move || {
        let user = db::get_user_by_username(&db, &username)
            .ok_or("用户名或密码错误")?;
        if user.status != 1 {
            return Err("账号已被禁用");
        }
        if !auth::verify_password(&password, &user.password_hash) {
            return Err("用户名或密码错误");
        }
        // 超管使用默认密码时强制修改
        let need_reset = user.role == 0 && password == default_sa_pass;
        let token = auth::create_jwt(&jwt_secret, user.id, user.role, &user.username, jwt_hours)
            .map_err(|_| "Token生成失败")?;
        Ok(serde_json::json!({
            "token": token,
            "need_reset": need_reset,
            "user": {
                "id": user.id,
                "username": user.username,
                "role": user.role,
                "display_name": user.display_name,
            }
        }))
    })
    .await
    .unwrap();

    match result {
        Ok(data) => Json(ApiResponse::success(data)),
        Err(e) => Json(ApiResponse::error(401, e)),
    }
}

// ===== 注册接口 =====

/// POST /api/v1/auth/register
pub async fn register(
    State(state): State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<RegisterRequest>,
) -> Json<ApiResponse> {
    let ip = get_client_ip(&headers, &addr);
    if !state.rate_limiter.check(&ip) {
        return Json(ApiResponse::error(429, "请求过于频繁，请稍后重试"));
    }

    if req.username.len() < 3 || req.username.len() > 32 {
        return Json(ApiResponse::error(400, "用户名长度应为3-32位"));
    }
    if req.password.len() < 6 {
        return Json(ApiResponse::error(400, "密码长度不能少于6位"));
    }

    let db = state.db.clone();
    let display = if req.display_name.is_empty() { req.username.clone() } else { req.display_name.clone() };

    let result = tokio::task::spawn_blocking(move || {
        let hash = auth::hash_password(&req.password)?;
        // 注册为普通用户(role=2)
        let id = db::create_user(&db, &req.username, &hash, 2, &display, 0, None)?;
        Ok::<_, String>(serde_json::json!({ "id": id }))
    }).await.unwrap();

    match result {
        Ok(data) => Json(ApiResponse::success(data)),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

// ===== 修改密码接口 =====

/// POST /api/v1/admin/change_password
pub async fn change_password(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<ChangePasswordRequest>,
) -> Json<ApiResponse> {
    if req.new_password.len() < 6 {
        return Json(ApiResponse::error(400, "新密码长度不能少于6位"));
    }

    let db = state.db.clone();
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || {
        let user = db::get_user_by_id(&db, uid).ok_or("用户不存在")?;
        if !auth::verify_password(&req.old_password, &user.password_hash) {
            return Err("原密码错误");
        }
        let new_hash = auth::hash_password(&req.new_password).map_err(|_| "密码加密失败")?;
        db::update_user(&db, &UpdateUserRequest {
            id: uid,
            display_name: None,
            password: None,
            role: None,
            status: None,
            max_licenses: None,
        }, Some(&new_hash)).map_err(|_| "更新失败")?;
        Ok(())
    }).await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("密码修改成功")),
        Err(e) => Json(ApiResponse::error(400, e)),
    }
}

// ===== 日志清理接口 =====

/// POST /api/v1/admin/license/cleanup_logs
pub async fn admin_cleanup_logs(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();

    let deleted = tokio::task::spawn_blocking(move || db::cleanup_old_logs(&db, 30))
        .await.unwrap();

    Json(ApiResponse::success(serde_json::json!({
        "deleted": deleted,
        "message": format!("已清理 {} 条30天前的日志", deleted),
    })))
}

// ===== JWT认证中间件 =====

pub async fn jwt_auth_middleware(
    State(state): State<Arc<AppState>>,
    mut request: Request,
    next: Next,
) -> Response {
    let auth_header = request
        .headers()
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");

    if !auth_header.starts_with("Bearer ") {
        return Json(ApiResponse::error(401, "认证失败")).into_response();
    }

    let token = &auth_header[7..];
    match auth::verify_jwt(&state.config.server.jwt_secret, token) {
        Ok(claims) => {
            request.extensions_mut().insert(claims);
            next.run(request).await
        }
        Err(_) => Json(ApiResponse::error(401, "Token无效或已过期")).into_response(),
    }
}

/// 从 Claims 获取 dealer_id：超管看全部，授权商只看自己的
fn dealer_filter(claims: &Claims) -> Option<i64> {
    if claims.role == 0 { None } else { Some(claims.sub) }
}

// ===== 管理接口 =====

/// GET /api/v1/admin/licenses
pub async fn admin_list(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<PageQuery>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let keyword = query.keyword.clone().unwrap_or_default();
    let status = query.status;
    let page = query.page;
    let limit = query.limit;
    let did = dealer_filter(&claims);

    let (list, total) = tokio::task::spawn_blocking(move || {
        db::list_licenses(&db, page, limit, &keyword, status, did)
    })
    .await
    .unwrap();

    Json(ApiResponse::success(serde_json::json!({
        "list": list,
        "total": total,
    })))
}

/// POST /api/v1/admin/license/create
pub async fn admin_create(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(mut req): Json<CreateLicenseRequest>,
) -> Json<ApiResponse> {
    // 普通用户不能创建
    if claims.role == 2 {
        return Json(ApiResponse::error(403, "无权操作"));
    }
    // 授权商检查配额
    if claims.role == 1 {
        req.dealer_id = claims.sub;
        let db = state.db.clone();
        let uid = claims.sub;
        let (count, max) = tokio::task::spawn_blocking(move || {
            let c = db::count_dealer_licenses(&db, uid);
            let u = db::get_user_by_id(&db, uid).map(|u| u.max_licenses).unwrap_or(0);
            (c, u)
        }).await.unwrap();
        if count >= max as i64 {
            return Json(ApiResponse::error(403, &format!("已达配额上限 {}/{}", count, max)));
        }
    }

    let db = state.db.clone();
    let key = auth::generate_license_key();
    let key_clone = key.clone();
    let who = claims.username.clone();

    let result = tokio::task::spawn_blocking(move || {
        let id = db::create_license(&db, &key_clone, &req)?;
        db::add_log(&db, id, "create", &who, &format!("创建授权码: {}", key_clone));
        Ok::<_, String>(serde_json::json!({
            "id": id,
            "license_key": key_clone,
        }))
    })
    .await
    .unwrap();

    match result {
        Ok(data) => Json(ApiResponse::success(data)),
        Err(e) => Json(ApiResponse::error(500, &e)),
    }
}

/// POST /api/v1/admin/license/update
pub async fn admin_update(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<UpdateLicenseRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let id = req.id;
    let who = claims.username.clone();
    let role = claims.role;
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || -> Result<String, String> {
        // 授权商只能改自己的
        if role == 1 {
            if let Some(lic) = db::get_by_id(&db, id) {
                if lic.dealer_id != uid { return Err("无权操作此授权码".into()); }
            }
        }
        db::update_license(&db, &req)?;
        db::add_log(&db, id, "update", &who, "更新授权码信息");
        if let Some(lic) = db::get_by_id(&db, id) {
            return Ok(lic.license_key);
        }
        Ok(String::new())
    })
    .await
    .unwrap();

    match result {
        Ok(key) => {
            if !key.is_empty() {
                state.cache.invalidate(&key);
            }
            Json(ApiResponse::success_msg("更新成功"))
        }
        Err(e) => Json(ApiResponse::error(500, &e)),
    }
}

/// POST /api/v1/admin/license/revoke
pub async fn admin_revoke(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<IdRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let id = req.id;
    let who = claims.username.clone();
    let role = claims.role;
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || -> Result<Option<String>, String> {
        let key = if let Some(lic) = db::get_by_id(&db, id) {
            if role == 1 && lic.dealer_id != uid { return Err("无权操作此授权码".into()); }
            Some(lic.license_key)
        } else { None };
        db::set_status(&db, id, 0);
        db::add_log(&db, id, "revoke", &who, "吊销授权码");
        Ok(key)
    }).await.unwrap();

    match result {
        Ok(key) => {
            if let Some(k) = key { state.cache.invalidate(&k); }
            Json(ApiResponse::success_msg("已吊销"))
        }
        Err(e) => Json(ApiResponse::error(403, &e)),
    }
}

/// POST /api/v1/admin/license/enable
pub async fn admin_enable(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<IdRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let id = req.id;
    let who = claims.username.clone();
    let role = claims.role;
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || -> Result<(), String> {
        if role == 1 {
            if let Some(lic) = db::get_by_id(&db, id) {
                if lic.dealer_id != uid { return Err("无权操作此授权码".into()); }
            }
        }
        db::set_status(&db, id, 1);
        db::add_log(&db, id, "enable", &who, "启用授权码");
        Ok(())
    }).await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("已启用")),
        Err(e) => Json(ApiResponse::error(403, &e)),
    }
}

/// POST /api/v1/admin/license/unbind
pub async fn admin_unbind(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<IdRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let id = req.id;
    let who = claims.username.clone();
    let role = claims.role;
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || {
        let key = if let Some(lic) = db::get_by_id(&db, id) {
            if role == 1 && lic.dealer_id != uid { return Err("无权操作此授权码".into()); }
            Some(lic.license_key)
        } else { None };
        db::unbind_machine(&db, id)?;
        db::add_log(&db, id, "unbind", &who, "解绑机器码");
        Ok::<_, String>(key)
    })
    .await
    .unwrap();

    match result {
        Ok(key) => {
            if let Some(k) = key {
                state.cache.invalidate(&k);
            }
            Json(ApiResponse::success_msg("已解绑"))
        }
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// POST /api/v1/admin/license/renew
pub async fn admin_renew(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<RenewRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let id = req.id;
    let days = req.expire_days;
    let who = claims.username.clone();
    let role = claims.role;
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || {
        let key = if let Some(lic) = db::get_by_id(&db, id) {
            if role == 1 && lic.dealer_id != uid { return Err("无权操作此授权码".into()); }
            Some(lic.license_key)
        } else { None };
        let new_expire = db::renew_license(&db, id, days)?;
        db::add_log(&db, id, "renew", &who, &format!("续期 {} 天，新到期时间: {}", days, new_expire));
        Ok::<_, String>((new_expire, key))
    })
    .await
    .unwrap();

    match result {
        Ok((new_expire, key)) => {
            if let Some(k) = key {
                state.cache.invalidate(&k);
            }
            Json(ApiResponse::success(serde_json::json!({
                "expire_at": new_expire,
            })))
        }
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// DELETE /api/v1/admin/license/delete/{id}
pub async fn admin_delete(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Path(id): Path<i64>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let role = claims.role;
    let uid = claims.sub;

    // 权限检查 + 删除 合并为一次 spawn_blocking
    let result = tokio::task::spawn_blocking(move || -> Result<Option<String>, String> {
        // 权限检查 + 获取 key
        let key = if let Some(lic) = db::get_by_id(&db, id) {
            if role == 1 && lic.dealer_id != uid {
                return Err("无权操作此授权码".into());
            }
            Some(lic.license_key)
        } else {
            None
        };
        db::delete_license(&db, id)?;
        Ok(key)
    }).await.unwrap();

    match result {
        Ok(key) => {
            if let Some(k) = key {
                state.cache.invalidate(&k);
            }
            Json(ApiResponse::success_msg("已删除"))
        }
        Err(e) => Json(ApiResponse::error(if e.contains("无权") { 403 } else { 500 }, &e)),
    }
}

/// GET /api/v1/admin/license/logs
pub async fn admin_logs(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<LogQuery>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let page = query.page;
    let limit = query.limit;
    let license_id = query.license_id;
    let action = query.action.clone();
    let did = dealer_filter(&claims);

    let (list, total) = tokio::task::spawn_blocking(move || {
        db::list_logs(&db, page, limit, license_id, action.as_deref(), did)
    })
    .await
    .unwrap();

    Json(ApiResponse::success(serde_json::json!({
        "list": list,
        "total": total,
    })))
}

/// GET /api/v1/admin/license/dashboard
pub async fn admin_dashboard(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let threshold = state.config.security.offline_threshold_secs;
    let did = dealer_filter(&claims);

    let dashboard = tokio::task::spawn_blocking(move || db::dashboard(&db, threshold, did))
        .await
        .unwrap();

    Json(ApiResponse::success(dashboard))
}

// ===== 用户管理接口（仅超管） =====

/// GET /api/v1/admin/users
pub async fn admin_user_list(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<PageQuery>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();
    let page = query.page;
    let limit = query.limit;

    let (list, total) = tokio::task::spawn_blocking(move || db::list_users(&db, page, limit))
        .await.unwrap();

    Json(ApiResponse::success(serde_json::json!({ "list": list, "total": total })))
}

/// POST /api/v1/admin/user/create
pub async fn admin_user_create(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<CreateUserRequest>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();
    let created_by = claims.sub;

    let result = tokio::task::spawn_blocking(move || {
        let hash = auth::hash_password(&req.password)?;
        let display = if req.display_name.is_empty() { req.username.clone() } else { req.display_name.clone() };
        let id = db::create_user(&db, &req.username, &hash, req.role, &display, req.max_licenses, Some(created_by))?;
        Ok::<_, String>(serde_json::json!({ "id": id }))
    }).await.unwrap();

    match result {
        Ok(data) => Json(ApiResponse::success(data)),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// POST /api/v1/admin/user/update
pub async fn admin_user_update(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<UpdateUserRequest>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();

    let result = tokio::task::spawn_blocking(move || {
        let pw_hash = if let Some(ref pw) = req.password {
            Some(auth::hash_password(pw)?)
        } else { None };
        db::update_user(&db, &req, pw_hash.as_deref())
    }).await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("更新成功")),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// DELETE /api/v1/admin/user/delete/{id}
pub async fn admin_user_delete(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Path(id): Path<i64>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();

    let result = tokio::task::spawn_blocking(move || db::delete_user(&db, id))
        .await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("已删除")),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// GET /api/v1/admin/me
pub async fn admin_me(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let uid = claims.sub;
    let user = tokio::task::spawn_blocking(move || db::get_user_by_id(&db, uid))
        .await.unwrap();
    match user {
        Some(u) => Json(ApiResponse::success(serde_json::json!({
            "id": u.id, "username": u.username, "role": u.role,
            "display_name": u.display_name, "max_licenses": u.max_licenses,
        }))),
        None => Json(ApiResponse::error(404, "用户不存在")),
    }
}

// ===== 批量操作接口 =====

/// POST /api/v1/admin/license/batch_create
pub async fn admin_batch_create(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(mut req): Json<BatchCreateRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    if req.count < 1 || req.count > 100 {
        return Json(ApiResponse::error(400, "批量数量应在 1-100 之间"));
    }
    if claims.role == 1 {
        req.dealer_id = claims.sub;
        let db = state.db.clone();
        let uid = claims.sub;
        let count_needed = req.count as i64;
        let (existing, max) = tokio::task::spawn_blocking(move || {
            let c = db::count_dealer_licenses(&db, uid);
            let u = db::get_user_by_id(&db, uid).map(|u| u.max_licenses).unwrap_or(0);
            (c, u)
        }).await.unwrap();
        if existing + count_needed > max as i64 {
            return Json(ApiResponse::error(403, &format!("配额不足 {}/{}", existing, max)));
        }
    }

    let keys: Vec<String> = (0..req.count).map(|_| auth::generate_license_key()).collect();
    let db = state.db.clone();
    let who = claims.username.clone();
    let keys_clone = keys.clone();

    let result = tokio::task::spawn_blocking(move || {
        let results = db::batch_create(&db, &keys_clone, &req)?;
        for (id, key) in &results {
            db::add_log(&db, *id, "create", &who, &format!("批量创建授权码: {}", key));
        }
        Ok::<_, String>(results)
    }).await.unwrap();

    match result {
        Ok(results) => {
            let keys: Vec<&str> = results.iter().map(|(_, k)| k.as_str()).collect();
            Json(ApiResponse::success(serde_json::json!({
                "count": results.len(),
                "keys": keys,
            })))
        }
        Err(e) => Json(ApiResponse::error(500, &e)),
    }
}

/// POST /api/v1/admin/license/batch_revoke
pub async fn admin_batch_revoke(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<IdsRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let who = claims.username.clone();

    let keys = tokio::task::spawn_blocking(move || {
        let keys = db::batch_set_status(&db, &req.ids, 0);
        for &id in &req.ids {
            db::add_log(&db, id, "revoke", &who, "批量吊销");
        }
        keys
    }).await.unwrap();

    for k in &keys { state.cache.invalidate(k); }
    Json(ApiResponse::success(serde_json::json!({ "count": keys.len() })))
}

/// POST /api/v1/admin/license/batch_renew
pub async fn admin_batch_renew(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<BatchRenewRequest>,
) -> Json<ApiResponse> {
    if claims.role == 2 { return Json(ApiResponse::error(403, "无权操作")); }
    let db = state.db.clone();
    let who = claims.username.clone();
    let days = req.expire_days;

    let result = tokio::task::spawn_blocking(move || {
        let results = db::batch_renew(&db, &req.ids, days)?;
        for (id, new_expire) in &results {
            db::add_log(&db, *id, "renew", &who, &format!("批量续期 {} 天，新到期: {}", days, new_expire));
        }
        Ok::<_, String>(results)
    }).await.unwrap();

    match result {
        Ok(results) => Json(ApiResponse::success(serde_json::json!({ "count": results.len() }))),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

// ===== 试用授权（公开接口） =====

/// POST /api/v1/license/trial
pub async fn request_trial(
    State(state): State<Arc<AppState>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<TrialRequest>,
) -> Json<ApiResponse> {
    let ip = get_client_ip(&headers, &addr);
    if !state.rate_limiter.check(&ip) {
        return Json(ApiResponse::error(429, "请求过于频繁，请稍后重试"));
    }

    // 验证签名（复用heartbeat签名格式: machine_id+timestamp+secret）
    let expected = auth::hmac_sign(
        &state.config.security.client_secret,
        &format!("{}{}", req.machine_id, req.timestamp),
    );
    if !auth::constant_time_eq(req.sign.as_bytes(), expected.as_bytes()) {
        return Json(ApiResponse::error(403, "签名无效"));
    }

    let db = state.db.clone();
    let machine_id = req.machine_id.clone();
    let domain = if req.domain.is_empty() { "*".to_string() } else { req.domain.clone() };
    let ip_clone = ip.clone();

    let result = tokio::task::spawn_blocking(move || {
        if db::check_trial_exists(&db, &machine_id) {
            return Err("该设备已申请过试用");
        }
        let key = auth::generate_license_key();
        let id = db::create_trial_license(&db, &key, &machine_id, &domain, 7)
            .map_err(|_| "创建试用授权失败")?;
        db::add_log(&db, id, "trial", &ip_clone, &format!("试用申请: {}", machine_id));
        Ok(serde_json::json!({
            "license_key": key,
            "expire_days": 7,
            "plan": "trial",
        }))
    }).await.unwrap();

    match result {
        Ok(data) => Json(ApiResponse::success(data)),
        Err(e) => Json(ApiResponse::error(400, e)),
    }
}

// ===== 公告管理接口 =====

/// GET /api/v1/admin/notices
pub async fn admin_notice_list(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<PageQuery>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();
    let page = query.page;
    let limit = query.limit;

    let (list, total) = tokio::task::spawn_blocking(move || db::list_notices(&db, page, limit))
        .await.unwrap();

    Json(ApiResponse::success(serde_json::json!({ "list": list, "total": total })))
}

/// POST /api/v1/admin/notice/create
pub async fn admin_notice_create(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<CreateNoticeRequest>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();
    let uid = claims.sub;

    let result = tokio::task::spawn_blocking(move || db::create_notice(&db, &req, uid))
        .await.unwrap();

    match result {
        Ok(id) => Json(ApiResponse::success(serde_json::json!({ "id": id }))),
        Err(e) => Json(ApiResponse::error(500, &e)),
    }
}

/// POST /api/v1/admin/notice/update
pub async fn admin_notice_update(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Json(req): Json<UpdateNoticeRequest>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();

    let result = tokio::task::spawn_blocking(move || db::update_notice(&db, &req))
        .await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("更新成功")),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

/// DELETE /api/v1/admin/notice/delete/{id}
pub async fn admin_notice_delete(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Path(id): Path<i64>,
) -> Json<ApiResponse> {
    if claims.role != 0 { return Json(ApiResponse::error(403, "仅超管可操作")); }
    let db = state.db.clone();

    let result = tokio::task::spawn_blocking(move || db::delete_notice(&db, id))
        .await.unwrap();

    match result {
        Ok(()) => Json(ApiResponse::success_msg("已删除")),
        Err(e) => Json(ApiResponse::error(400, &e)),
    }
}

// ===== 统计报表接口 =====

/// GET /api/v1/admin/stats/trend
pub async fn admin_stats_trend(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<TrendQuery>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let days = query.days.min(90).max(7);
    let threshold = state.config.security.offline_threshold_secs;
    let did = dealer_filter(&claims);

    let trend = tokio::task::spawn_blocking(move || db::trend_stats(&db, days, threshold, did))
        .await.unwrap();

    Json(ApiResponse::success(trend))
}

/// GET /api/v1/admin/stats/distribution
pub async fn admin_stats_distribution(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
) -> Json<ApiResponse> {
    let db = state.db.clone();
    let did = dealer_filter(&claims);

    let dist = tokio::task::spawn_blocking(move || db::plan_distribution(&db, did))
        .await.unwrap();

    Json(ApiResponse::success(dist))
}

// ===== 导出接口 =====

/// GET /api/v1/admin/licenses/export
pub async fn admin_export(
    State(state): State<Arc<AppState>>,
    Extension(claims): Extension<Claims>,
    Query(query): Query<ExportQuery>,
) -> impl IntoResponse {
    let db = state.db.clone();
    let did = dealer_filter(&claims);
    let status = query.status;

    let licenses = tokio::task::spawn_blocking(move || db::export_licenses(&db, status, did))
        .await.unwrap();

    // 生成 CSV
    let mut csv = String::from("ID,授权码,域名,备注,套餐,状态,试用,到期时间,最后心跳,最后IP,版本,创建时间\n");
    for l in &licenses {
        let status_text = match l.status { 1 => "活跃", 0 => "吊销", 2 => "过期", _ => "未知" };
        let trial_text = if l.is_trial == 1 { "是" } else { "否" };
        csv.push_str(&format!(
            "{},{},{},{},{},{},{},{},{},{},{},{}\n",
            l.id, l.license_key, l.domain,
            l.note.replace(',', "，"),
            l.plan, status_text, trial_text,
            l.expire_at.as_deref().unwrap_or("永久"),
            l.last_heartbeat.as_deref().unwrap_or("-"),
            l.last_ip, l.version, l.created_at,
        ));
    }

    (
        [
            (axum::http::header::CONTENT_TYPE, "text/csv; charset=utf-8"),
            (axum::http::header::CONTENT_DISPOSITION, "attachment; filename=licenses.csv"),
        ],
        csv,
    )
}
