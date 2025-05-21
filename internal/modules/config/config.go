package config

type Config struct {
	StorageEnabled  bool   `yaml:"storage_enabled"`
	StorageSupplier string `yaml:"storage_supplier"`
}

type AliOss struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyId     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Bucket          string `yaml:"bucket"`
	Directory       string `yaml:"directory"`
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
	LowSpeed  []string `yaml:"low_speed"`
	HighSpeed []string `yaml:"high_speed"`
}
