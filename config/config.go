package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var GConfig *Config

func Init(filePath string) {
	config, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	initFromYaml(config)
	err = GConfig.Verify()
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
	StorageEnabled  bool   `yaml:"storage_enabled"`
	StorageSupplier string `yaml:"storage_supplier"`
	URLExpires      string `yaml:"url_expires"`
	AliOss          `yaml:"ali_oss"`
	MySQL           `yaml:"mysql"`
	Geek            `yaml:"geek"`
	Tuzi            `yaml:"tuzi"`
	V3              `yaml:"v3"`
	RequestOrder    `yaml:"request_order"`
}

func (c *Config) Verify() error {
	// todo 支持不转存
	if !c.StorageEnabled {
		return fmt.Errorf("storage_enabled must be true")
	}
	if c.StorageSupplier != "ali_oss" {
		return fmt.Errorf("storage_supplier must be ali_oss")
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

type Geek struct {
	LowPriceToken      string `yaml:"low_price_token"`
	BalanceToken       string `yaml:"balance_token"`
	HighAvailableToken string `yaml:"high_available_token"`
}

type V3 struct {
	Token string `yaml:"token"`
}

type Tuzi struct {
	DefaultChannelToken string `yaml:"default_channel_token"`
	OpenaiChannelToken  string `yaml:"openai_channel_token"`
}

type RequestOrder struct {
	SlowSpeed []Request `yaml:"slow_speed"`
	FastSpeed []Request `yaml:"fast_speed"`
}

type Request struct {
	Supplier  string `json:"supplier"`
	TokenName string `json:"token_name"`
	Model     string `json:"model"`
}
