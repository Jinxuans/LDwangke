package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Cache    CacheConfig    `yaml:"cache"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	License  LicenseConfig  `yaml:"license"`
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
		log.Fatalf("读取配置文件失败: %v", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	Global = cfg
	return cfg
}
