package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置根结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Lsky     LskyConfig     `mapstructure:"lsky"`
	Admin    AdminConfig    `mapstructure:"admin"`
	Site     SiteConfig     `mapstructure:"site"`
	Upload   UploadConfig   `mapstructure:"upload"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessExpire  time.Duration `mapstructure:"access_expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
}

// SMTPConfig 邮件服务配置
type SMTPConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	From       string `mapstructure:"from"`
	AdminEmail string `mapstructure:"admin_email"`
	UseTLS     bool   `mapstructure:"use_tls"`
}

// LskyConfig 兰空图床配置
type LskyConfig struct {
	BaseURL  string `mapstructure:"base_url"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	Token    string `mapstructure:"token"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Email string `mapstructure:"email"`
}

// SiteConfig 站点信息配置
type SiteConfig struct {
	Name      string `mapstructure:"name"`
	StartDate string `mapstructure:"start_date"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxSizeMB int `mapstructure:"max_size_mb"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

// LoadConfig 从 YAML 文件加载配置，支持环境变量覆盖
//
// 加载顺序（后加载的覆盖先加载的）：
//  1. 默认值
//  2. config.yaml（通过 path 参数或默认路径搜索）
//  3. 环境变量（前缀 GM_，如 GM_SMTP_PASSWORD）
//
// path 参数说明：
//   - 非空：加载指定路径的配置文件
//   - 空字符串：按默认路径搜索 config.yaml（当前目录和 ./backend/）
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		// 检查指定路径的文件是否存在
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
		}
		// 文件不存在则跳过，使用默认值+环境变量
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./backend")
	}

	// 设置默认值
	setDefaults(v)

	// 环境变量支持
	// GM_SMTP_PASSWORD -> smtp.password
	// GM_JWT_ACCESS_SECRET -> jwt.access_secret
	v.SetEnvPrefix("GM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 配置文件存在但无法读取，返回错误
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// 配置文件不存在，使用默认值 + 环境变量
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置所有配置项的默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 1323)
	v.SetDefault("server.host", "0.0.0.0")

	v.SetDefault("database.path", "gm_site.db")

	v.SetDefault("jwt.access_expire", "15m")
	v.SetDefault("jwt.refresh_expire", "168h")

	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.use_tls", true)

	v.SetDefault("site.name", "顾夏")

	v.SetDefault("upload.max_size_mb", 10)

	v.SetDefault("cors.allowed_origins", "http://localhost:5173")
}
