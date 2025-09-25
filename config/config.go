package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/reusedev/draw-hub/internal/consts"
	"gopkg.in/yaml.v3"
)

var GConfig *Config

func Init(data []byte) {
	initFromYaml(data)
	err := GConfig.Verify()
	if err != nil {
		panic(err)
	}
}

func initFromYaml(config []byte) {
	err := yaml.Unmarshal(config, &GConfig)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	// 日志配置
	LogLevel      string `yaml:"log_level"`
	LogFile       string `yaml:"log_file"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`

	LocalStorageDomain    string `yaml:"local_storage_domain"`
	LocalStorageDirectory string `yaml:"local_storage_directory"`
	CloudStorageEnabled   bool   `yaml:"cloud_storage_enabled"`
	CloudStorageSupplier  string `yaml:"cloud_storage_supplier"`
	URLExpires            string `yaml:"url_expires"`
	AliOss                `yaml:"ali_oss"`
	MySQL                 `yaml:"mysql"`
	RequestOrder          `yaml:"request_order"`
}

func (c *Config) Verify() error {
	// 验证日志级别
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	logLevelValid := false
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			logLevelValid = true
			break
		}
	}
	if !logLevelValid {
		return fmt.Errorf("log_level must be one of: %v", validLogLevels)
	}

	if _, err := url.Parse(c.LocalStorageDomain); err != nil {
		return fmt.Errorf("local_storage_domain is not a valid URL: %v", err)
	}
	if !strings.HasSuffix(c.LocalStorageDirectory, "/") {
		return fmt.Errorf("local_storage_directory must end with '/'")
	}
	if c.CloudStorageEnabled {
		if c.CloudStorageSupplier != "ali_oss" {
			return fmt.Errorf("storage_supplier must be ali_oss")
		}
		if !strings.HasSuffix(c.AliOss.Directory, "/") {
			return fmt.Errorf("ali_oss.directory must end with '/'")
		}
	}
	_, err := time.ParseDuration(c.URLExpires)
	if err != nil {
		return err
	}
	return nil
}

type AliOss struct {
	AccessKeyId     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	Bucket          string `yaml:"bucket"`
	Directory       string `yaml:"directory"`
}

type MySQL struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type RequestOrder struct {
	SlowSpeed       [][]Request `yaml:"slow_speed"`
	FastSpeed       [][]Request `yaml:"fast_speed"`
	DeepSearch      [][]Request `yaml:"deepsearch"`
	Gemini25Flash   [][]Request `yaml:"gemini_25_flash"`
	Gemini25FlashHD [][]Request `yaml:"gemini_25_flash_hd"`
	JiMengV40       [][]Request `yaml:"jimeng_v40"`
}

type Request struct {
	Supplier string `json:"supplier"`
	Token    string `json:"token"`
	Desc     string `json:"desc"`
	Model    string `json:"model"`
}

func GetRequestOrder(model consts.Model) []Request {
	return nil
}
