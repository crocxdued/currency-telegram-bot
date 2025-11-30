package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	BotToken        string `mapstructure:"BOT_TOKEN"`
	DBURL           string
	LogLevel        string `mapstructure:"LOG_LEVEL"`
	CacheTTLMinutes int    `mapstructure:"CACHE_TTL_MINUTES"`
}

func Load() (*Config, error) {
	// Только окружение! Файл .env НЕ читаем
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("CACHE_TTL_MINUTES", 5)
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_SSLMODE", "disable")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_DB", "currency_bot")

	var c Config

	// Собираем DB_URL из отдельных переменных
	c.DBURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_DB"),
		viper.GetString("POSTGRES_SSLMODE"),
	)

	c.BotToken = viper.GetString("BOT_TOKEN")
	c.LogLevel = viper.GetString("LOG_LEVEL")
	c.CacheTTLMinutes = viper.GetInt("CACHE_TTL_MINUTES")

	if c.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	return &c, nil
}
