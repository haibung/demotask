package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	str2duration "github.com/xhit/go-str2duration/v2"
)

var c *Config

const (
	Development = "Development"
	Production  = "Production"
)

type Config struct {
	Env        string `yaml:"env"`
	Port       int    `yaml:"port"`
	MaxFailure int    `yaml:"maxFailure"`
	Postgres   struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		SSLMode  string `yaml:"sslMode"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		DBName   string `yaml:"dbName"`
	} `yaml:"postgres"`
	Redis struct {
		Host      string `yaml:"host"`
		Port      string `yaml:"port"`
		Password  string `yaml:"password"`
		DefaultDB int    `yaml:"defaultDb"`
	} `yaml:"redis"`
	Auth struct {
		ExpireAccessToken          string `yaml:"expireAccessToken"`
		ExpireRefreshToken         string `yaml:"expireRefreshToken"`
		Secret                     string `yaml:"secret"`
		SecretClaim                string `yaml:"secretClaim"`
		ExpireAccessTokenDuration  time.Duration
		ExpireRefreshTokenDuration time.Duration
	} `yaml:"auth"`
	PayPal struct {
		ClientID     string `yaml:"clientId"`
		ClientSecret string `yaml:"clientSecret"`
		RedirectURL  string `yaml:"redirectUrl"`
		WebhookID    string `yaml:"webhookId"`
		// WebhookSecret      string `yaml:"webhookSecret"`
		Environment        string `yaml:"environment"`
		SandboxURL         string `yaml:"sandboxUrl"`
		CheckoutSuccessURL string `yaml:"checkoutSuccessUrl"`
		CheckoutCancelURL  string `yaml:"checkoutCancelUrl"`
		BaseURL            string `yaml:"baseUrl"`
		AuthBaseURL        string `yaml:"authBaseUrl"`
	} `yaml:"paypal"`
}

func Get() *Config {
	return c
}

func SetConfig() {
	var err error

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config.yml")

	if err := viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	if err := viper.Unmarshal(&c); err != nil {
		panic(err.Error())
	}

	c.Auth.ExpireAccessTokenDuration, err = str2duration.ParseDuration(c.Auth.ExpireAccessToken)
	if err != nil {
		panic(fmt.Sprintf("config auth access expired duration string not valid: %s", err.Error()))
	}

	c.Auth.ExpireRefreshTokenDuration, err = str2duration.ParseDuration(c.Auth.ExpireRefreshToken)
	if err != nil {
		panic(fmt.Sprintf("config auth refresh token duration string not valid: %s", err.Error()))
	}

	// Set PayPal BaseURL based on environment
	if c.PayPal.Environment == "live" {
		c.PayPal.BaseURL = "https://api-m.paypal.com"
	} else {
		c.PayPal.BaseURL = "https://api-m.sandbox.paypal.com"
	}

	viper.WatchConfig()
}
