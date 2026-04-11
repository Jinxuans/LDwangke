package config

import (
	"os"
	"strconv"
	"strings"

	obslogger "go-api/internal/observability/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Cache     CacheConfig     `yaml:"cache"`
	License   LicenseConfig   `yaml:"license"`
	Security  SecurityConfig  `yaml:"security"`
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
}

type SecurityConfig struct {
	AllowLegacyPlaintextPasswords *bool `yaml:"allow_legacy_plaintext_passwords"`
}

type BootstrapConfig struct {
	CreateDefaultAdmin *bool  `yaml:"create_default_admin"`
	AutoMigrate        *bool  `yaml:"auto_migrate"`
	MigrationsDir      string `yaml:"migrations_dir"`
}

type LicenseConfig struct {
	ServerURL   string `yaml:"server_url"`   // 授权站地址，如 https://29.colnt.com
	LicenseKey  string `yaml:"license_key"`  // 授权码（也可从 key_file 读取）
	KeyFile     string `yaml:"key_file"`     // 授权码文件路径，优先级低于 license_key
	Domain      string `yaml:"domain"`       // 当前站点域名
	CacheFile   string `yaml:"cache_file"`   // 离线缓存文件路径，默认 .sys_state
	SecretsFile string `yaml:"secrets_file"` // 密钥文件路径，默认 .secrets（宝塔插件生成）
}

type SMTPConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	FromName   string `yaml:"from_name"`
	Encryption string `yaml:"encryption"` // ssl | starttls | none
}

type ServerConfig struct {
	Port         string `yaml:"port"`
	Mode         string `yaml:"mode"`
	PhpBackend   string `yaml:"php_backend"`    // PHP 后端地址（Go代理用），如 http://127.0.0.1:9000
	PhpPublicURL string `yaml:"php_public_url"` // PHP 公网地址（浏览器iframe用），留空=同域相对路径
	BridgeSecret string `yaml:"bridge_secret"`  // iframe 认证桥接签名密钥
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	AccessTTL  int    `yaml:"access_ttl"`
	RefreshTTL int    `yaml:"refresh_ttl"`
}

type CacheConfig struct {
	OrderListTTL int `yaml:"order_list_ttl"`
	ClassListTTL int `yaml:"class_list_ttl"`
}

var Global *Config

func Load(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		obslogger.Fatal("读取配置文件失败", "path", path, "error", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		obslogger.Fatal("解析配置文件失败", "path", path, "error", err)
	}

	applyEnvOverrides(cfg)
	Global = cfg
	return cfg
}

func (c *Config) AllowLegacyPlaintextPasswords() bool {
	if v, ok := envBool("GO_API_SECURITY_ALLOW_LEGACY_PLAINTEXT_PASSWORDS"); ok {
		return v
	}
	if c.Security.AllowLegacyPlaintextPasswords != nil {
		return *c.Security.AllowLegacyPlaintextPasswords
	}
	return false
}

func (c *Config) CreateDefaultAdminEnabled() bool {
	if v, ok := envBool("GO_API_BOOTSTRAP_CREATE_DEFAULT_ADMIN"); ok {
		return v
	}
	if c.Bootstrap.CreateDefaultAdmin != nil {
		return *c.Bootstrap.CreateDefaultAdmin
	}
	return false
}

func (c *Config) AutoMigrateEnabled() bool {
	if v, ok := envBool("GO_API_BOOTSTRAP_AUTO_MIGRATE"); ok {
		return v
	}
	if c.Bootstrap.AutoMigrate != nil {
		return *c.Bootstrap.AutoMigrate
	}
	return true
}

func (c *Config) MigrationsDirValue() string {
	if val, ok := os.LookupEnv("GO_API_BOOTSTRAP_MIGRATIONS_DIR"); ok {
		return strings.TrimSpace(val)
	}
	return strings.TrimSpace(c.Bootstrap.MigrationsDir)
}

func applyEnvOverrides(cfg *Config) {
	cfg.Server.Port = envString("GO_API_SERVER_PORT", cfg.Server.Port)
	cfg.Server.Mode = envString("GO_API_SERVER_MODE", cfg.Server.Mode)
	cfg.Server.PhpBackend = envString("GO_API_SERVER_PHP_BACKEND", cfg.Server.PhpBackend)
	cfg.Server.PhpPublicURL = envString("GO_API_SERVER_PHP_PUBLIC_URL", cfg.Server.PhpPublicURL)
	cfg.Server.BridgeSecret = envString("GO_API_SERVER_BRIDGE_SECRET", cfg.Server.BridgeSecret)

	cfg.Database.Host = envString("GO_API_DATABASE_HOST", cfg.Database.Host)
	cfg.Database.Port = envInt("GO_API_DATABASE_PORT", cfg.Database.Port)
	cfg.Database.User = envString("GO_API_DATABASE_USER", cfg.Database.User)
	cfg.Database.Password = envString("GO_API_DATABASE_PASSWORD", cfg.Database.Password)
	cfg.Database.DBName = envString("GO_API_DATABASE_DBNAME", cfg.Database.DBName)
	cfg.Database.MaxOpenConns = envInt("GO_API_DATABASE_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxIdleConns = envInt("GO_API_DATABASE_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)

	cfg.Redis.Host = envString("GO_API_REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = envInt("GO_API_REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = envString("GO_API_REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Redis.DB = envInt("GO_API_REDIS_DB", cfg.Redis.DB)

	cfg.JWT.Secret = envString("GO_API_JWT_SECRET", cfg.JWT.Secret)
	cfg.JWT.AccessTTL = envInt("GO_API_JWT_ACCESS_TTL", cfg.JWT.AccessTTL)
	cfg.JWT.RefreshTTL = envInt("GO_API_JWT_REFRESH_TTL", cfg.JWT.RefreshTTL)

	cfg.Cache.OrderListTTL = envInt("GO_API_CACHE_ORDER_LIST_TTL", cfg.Cache.OrderListTTL)
	cfg.Cache.ClassListTTL = envInt("GO_API_CACHE_CLASS_LIST_TTL", cfg.Cache.ClassListTTL)

	cfg.License.ServerURL = envString("GO_API_LICENSE_SERVER_URL", cfg.License.ServerURL)
	cfg.License.LicenseKey = envString("GO_API_LICENSE_LICENSE_KEY", cfg.License.LicenseKey)
	cfg.License.KeyFile = envString("GO_API_LICENSE_KEY_FILE", cfg.License.KeyFile)
	cfg.License.Domain = envString("GO_API_LICENSE_DOMAIN", cfg.License.Domain)
	cfg.License.CacheFile = envString("GO_API_LICENSE_CACHE_FILE", cfg.License.CacheFile)
	cfg.License.SecretsFile = envString("GO_API_LICENSE_SECRETS_FILE", cfg.License.SecretsFile)

	if v, ok := envBool("GO_API_SECURITY_ALLOW_LEGACY_PLAINTEXT_PASSWORDS"); ok {
		cfg.Security.AllowLegacyPlaintextPasswords = &v
	}
	if v, ok := envBool("GO_API_BOOTSTRAP_CREATE_DEFAULT_ADMIN"); ok {
		cfg.Bootstrap.CreateDefaultAdmin = &v
	}
	if v, ok := envBool("GO_API_BOOTSTRAP_AUTO_MIGRATE"); ok {
		cfg.Bootstrap.AutoMigrate = &v
	}
	cfg.Bootstrap.MigrationsDir = envString("GO_API_BOOTSTRAP_MIGRATIONS_DIR", cfg.Bootstrap.MigrationsDir)
}

func envString(key, current string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return current
}

func envInt(key string, current int) int {
	if val, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(strings.TrimSpace(val))
		if err == nil {
			return parsed
		}
		obslogger.L().Warn("忽略无效整数环境变量", "key", key, "value", val)
	}
	return current
}

func envBool(key string) (bool, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		obslogger.L().Warn("忽略无效布尔环境变量", "key", key, "value", val)
		return false, false
	}
	return parsed, true
}
