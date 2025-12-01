package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App         AppConfig       `mapstructure:"app"`
	Database    DatabaseConfig  `mapstructure:"database"`
	JWT         JWTConfig       `mapstructure:"jwt"`
	MinIO       MinioConfig     `mapstructure:"minio"`
	Services    ServicesConfig  `mapstructure:"services"`
	Logging     LoggingConfig   `mapstructure:"logging"`
	Server      ServerConfig    `mapstructure:"server"`
	SendGrid    SendGrid        `mapstructure:"sendgrid"`
	Redis       RedisConfig     `mapstructure:"redis"`
	FileService FileServiceConf `mapstructure:"file_service"`
}

// APP
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// DATABASE
type DatabaseConfig struct {
	Host     string       `mapstructure:"host"`
	Port     int          `mapstructure:"port"`
	User     string       `mapstructure:"user"`
	Password string       `mapstructure:"password"`
	Name     string       `mapstructure:"name"`
	SSLMode  string       `mapstructure:"sslmode"`
	Pool     DBPoolConfig `mapstructure:"pool"`
}

type DBPoolConfig struct {
	MaxConnections    int    `mapstructure:"max_connections"`
	MinConnections    int    `mapstructure:"min_connections"`
	MaxConnLifetime   string `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   string `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod string `mapstructure:"health_check_period"`
}

// JWT
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expiry time.Duration `mapstructure:"expiry"`
	Issuer string        `mapstructure:"issuer"`
}

// MINIO / S3
type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

// SERVICES
type ServicesConfig struct {
	Admin        ServiceConfig `mapstructure:"admin"`
	Auth         ServiceConfig `mapstructure:"auth"`
	Users        ServiceConfig `mapstructure:"users"`
	File         ServiceConfig `mapstructure:"file"`
	Chat         ServiceConfig `mapstructure:"chat"`
	Notification ServiceConfig `mapstructure:"notification"`
	Gateway      ServiceConfig `mapstructure:"gateway"`
	Front        ServiceConfig `mapstructure:"front"`
}

type ServiceConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// LOGGING
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// HTTP SERVER CONFIG
type ServerConfig struct {
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	IdleTimeout     string `mapstructure:"idle_timeout"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
}

// SENDGRID
type SendGrid struct {
	Key        string `mapstructure:"api_key"`
	TemplateId string `mapstructure:"template_id"`
}

// REDIS (NEW)
type RedisConfig struct {
	Host     string   `mapstructure:"host"`
	Port     string   `mapstructure:"port"`
	DB       int      `mapstructure:"db"`
	Password string   `mapstructure:"password"`
	TTL      RedisTTL `mapstructure:"ttl"`
}

type RedisTTL struct {
	UploadSession time.Duration `mapstructure:"upload_session"`
	Presign       time.Duration `mapstructure:"presign"`
}

// FILE SERVICE EXTRA CONFIG (NEW)
type FileServiceConf struct {
	UploadURLExpiry   time.Duration   `mapstructure:"upload_url_expiry"`
	DownloadURLExpiry time.Duration   `mapstructure:"download_url_expiry"`
	MaxUploadSize     string          `mapstructure:"max_upload_size"`
	AllowedMimeTypes  []string        `mapstructure:"allowed_mime_types"`
	Antivirus         AntivirusConfig `mapstructure:"antivirus"`
}

type AntivirusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
}

// LOAD FUNCTION
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
