package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const cfgPath = "./config"

type Server struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ShutdowTimeout int    `mapstructure:"shutdown_timeout"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	IdleTimeout    int    `mapstructure:"idle_timeout"`
	Debug          bool   `mapstructure:"debug"`
}

type Postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type Kafka struct {
	Network     string   `mapstructure:"network"`
	Brokers     []string `mapstructure:"brokers"`
	Topic       string   `mapstructure:"topic"`
	GroupID     string   `mapstructure:"group_id"`
	PollTimeout int      `mapstructure:"poll_timeout"`
}

type Cache struct {
	Capacity int `mapstructure:"capacity"`
	Ttl      int `mapstructure:"ttl"`
}
type Config struct {
	Serv  Server   `mapstructure:"server"`
	Db    Postgres `mapstructure:"postgres"`
	Kafka `mapstructure:"kafka"`
	Cache `mapstructure:"cache"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	v.AddConfigPath(cfgPath)
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %w", err)
	}

	return &cfg, nil
}

func GetDbConnString(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Db.User, cfg.Db.Password, cfg.Db.Host, cfg.Db.Port, cfg.Db.Database, cfg.Db.SSLMode)
}

func GetServerAddr(cfg *Config) string {
	return fmt.Sprintf("%s:%d", cfg.Serv.Host, cfg.Serv.Port)
}
