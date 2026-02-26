use hmac::{Hmac, Mac};
use sha2::Sha256;
use jsonwebtoken::{encode, decode, Header, Validation, EncodingKey, DecodingKey};

use crate::model::{Claims, HeartbeatRequest, VerifyRequest};

type HmacSha256 = Hmac<Sha256>;

/// 验证 verify 请求的 HMAC-SHA256 签名
/// 签名算法：对参数按字母序排列拼接后，用 client_secret 做 HMAC-SHA256
pub fn verify_request_sign(
    client_secret: &str,
    req: &VerifyRequest,
    window_secs: i64,
) -> Result<(), &'static str> {
    check_timestamp(req.timestamp, window_secs)?;

    let sign_str = format!(
        "domain={}&license_key={}&machine_id={}&timestamp={}&version={}",
        req.domain, req.license_key, req.machine_id, req.timestamp, req.version
    );

    verify_hmac(client_secret, &sign_str, &req.sign)
}

/// 验证 heartbeat 请求的 HMAC-SHA256 签名
pub fn verify_heartbeat_sign(
    client_secret: &str,
    req: &HeartbeatRequest,
    window_secs: i64,
) -> Result<(), &'static str> {
    check_timestamp(req.timestamp, window_secs)?;

    let sign_str = format!(
        "license_key={}&machine_id={}&timestamp={}&version={}",
        req.license_key, req.machine_id, req.timestamp, req.version
    );

    verify_hmac(client_secret, &sign_str, &req.sign)
}

/// 生成授权码 QK-{32位大写HEX}
pub fn generate_license_key() -> String {
    use rand::Rng;
    let mut rng = rand::thread_rng();
    let bytes: Vec<u8> = (0..16).map(|_| rng.gen()).collect();
    format!("QK-{}", hex::encode_upper(bytes))
}

// ===== JWT =====

pub fn create_jwt(secret: &str, user_id: i64, role: i32, username: &str, expire_hours: u64) -> Result<String, String> {
    let exp = chrono::Utc::now().timestamp() as usize + (expire_hours as usize * 3600);
    let claims = Claims {
        sub: user_id,
        role,
        username: username.to_string(),
        exp,
    };
    encode(&Header::default(), &claims, &EncodingKey::from_secret(secret.as_bytes()))
        .map_err(|e| format!("JWT签发失败: {}", e))
}

pub fn verify_jwt(secret: &str, token: &str) -> Result<Claims, String> {
    decode::<Claims>(token, &DecodingKey::from_secret(secret.as_bytes()), &Validation::default())
        .map(|data| data.claims)
        .map_err(|e| format!("JWT验证失败: {}", e))
}

// ===== 密码 =====

pub fn hash_password(password: &str) -> Result<String, String> {
    bcrypt::hash(password, 10).map_err(|e| format!("密码哈希失败: {}", e))
}

pub fn verify_password(password: &str, hash: &str) -> bool {
    bcrypt::verify(password, hash).unwrap_or(false)
}

// ===== 内部工具 =====

fn check_timestamp(timestamp: i64, window_secs: i64) -> Result<(), &'static str> {
    let now = chrono::Utc::now().timestamp();
    if (now - timestamp).abs() > window_secs {
        return Err("签名已过期");
    }
    Ok(())
}

fn verify_hmac(secret: &str, data: &str, sign: &str) -> Result<(), &'static str> {
    let mut mac =
        HmacSha256::new_from_slice(secret.as_bytes()).map_err(|_| "HMAC 初始化失败")?;
    mac.update(data.as_bytes());
    let expected = hex::encode(mac.finalize().into_bytes());

    if constant_time_eq(expected.as_bytes(), sign.as_bytes()) {
        Ok(())
    } else {
        Err("签名验证失败")
    }
}

/// 常量时间比较，防止时序攻击
fn constant_time_eq(a: &[u8], b: &[u8]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    let mut diff = 0u8;
    for (x, y) in a.iter().zip(b.iter()) {
        diff |= x ^ y;
    }
    diff == 0
}
