use dashmap::DashMap;
use std::time::{Duration, Instant};

use crate::model::VerifyResponse;

pub struct VerifyCache {
    map: DashMap<String, (VerifyResponse, Instant)>,
    ttl: Duration,
}

impl VerifyCache {
    pub fn new(ttl_secs: u64) -> Self {
        Self {
            map: DashMap::new(),
            ttl: Duration::from_secs(ttl_secs),
        }
    }

    pub fn get(&self, key: &str) -> Option<VerifyResponse> {
        if let Some(entry) = self.map.get(key) {
            if entry.1.elapsed() < self.ttl {
                return Some(entry.0.clone());
            }
            drop(entry);
            self.map.remove(key);
        }
        None
    }

    pub fn set(&self, key: String, value: VerifyResponse) {
        self.map.insert(key, (value, Instant::now()));
    }

    pub fn invalidate(&self, key: &str) {
        self.map.remove(key);
    }

    /// 定期清理过期条目，防止内存泄漏
    pub fn cleanup(&self) {
        self.map.retain(|_, v| v.1.elapsed() < self.ttl);
    }
}

// ===== IP 速率限制器 =====

pub struct RateLimiter {
    /// IP -> (请求计数, 窗口起始时间)
    map: DashMap<String, (u32, Instant)>,
    max_requests: u32,
    window: Duration,
}

impl RateLimiter {
    pub fn new(max_requests: u32, window_secs: u64) -> Self {
        Self {
            map: DashMap::new(),
            max_requests,
            window: Duration::from_secs(window_secs),
        }
    }

    /// 检查是否允许请求，返回 true 表示放行，false 表示限流
    pub fn check(&self, ip: &str) -> bool {
        let mut entry = self.map.entry(ip.to_string()).or_insert((0, Instant::now()));
        let (count, start) = entry.value_mut();

        if start.elapsed() >= self.window {
            // 窗口已过，重置
            *count = 1;
            *start = Instant::now();
            return true;
        }

        if *count >= self.max_requests {
            return false;
        }

        *count += 1;
        true
    }

    /// 定期清理过期条目
    pub fn cleanup(&self) {
        self.map.retain(|_, v| v.1.elapsed() < self.window);
    }
}
