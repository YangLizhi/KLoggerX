package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	MinIO      MinIOConfig      `mapstructure:"minio"`
	Log        LogConfig        `mapstructure:"log"`
	OnlyOffice OnlyOfficeConfig `mapstructure:"onlyoffice"`
	Qdrant     QdrantConfig     `mapstructure:"qdrant"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expire time.Duration `mapstructure:"expire"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type OnlyOfficeConfig struct {
	ServerURL    string `mapstructure:"server_url"`     // OnlyOffice DocumentServer URL (e.g., http://localhost:8082)
	JWTSecret    string `mapstructure:"jwt_secret"`     // JWT secret for signing requests
	FileBaseURL  string `mapstructure:"file_base_url"`  // Base URL for file downloads (this server)
	CallbackURL  string `mapstructure:"callback_url"`   // Callback URL for OnlyOffice to save changes
}

type QdrantConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Collection string `mapstructure:"collection"`
	APIKey     string `mapstructure:"api_key"` // Optional API key for Qdrant Cloud
}

var Cfg *Config

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	Cfg = &Config{}
	return viper.Unmarshal(Cfg)
}

// GetJWTSecret 优先从环境变量读取 JWT Secret，否则使用配置文件
func GetJWTSecret() string {
	if secret := os.Getenv("KLOGGERX_JWT_SECRET"); secret != "" {
		return secret
	}
	return Cfg.JWT.Secret
}

// GetMinIOAccessKey 优先从环境变量读取 MinIO AccessKey
func GetMinIOAccessKey() string {
	if key := os.Getenv("KLOGGERX_MINIO_ACCESS_KEY"); key != "" {
		return key
	}
	return Cfg.MinIO.AccessKey
}

// GetMinIOSecretKey 优先从环境变量读取 MinIO SecretKey
func GetMinIOSecretKey() string {
	if key := os.Getenv("KLOGGERX_MINIO_SECRET_KEY"); key != "" {
		return key
	}
	return Cfg.MinIO.SecretKey
}
