package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Очищаем Viper
	viper.Reset()
	defer viper.Reset()

	// Устанавливаем все нужные переменные
	os.Setenv("BOT_TOKEN", "123:test_token")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DB", "testdb")
	os.Setenv("POSTGRES_SSLMODE", "disable")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("CACHE_TTL_MINUTES", "10")

	// Сбрасываем старые значения (на всякий случай)
	defer func() {
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DB")
		os.Unsetenv("POSTGRES_SSLMODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("CACHE_TTL_MINUTES")
	}()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.Equal(t, "123:test_token", cfg.BotToken)
	assert.Equal(t, "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable", cfg.DBURL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 10, cfg.CacheTTLMinutes)
}

func TestLoadConfig_MissingBotToken(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Убираем обязательные переменные
	os.Unsetenv("BOT_TOKEN")
	os.Unsetenv("DB_URL")

	// Сбрасываем Viper чтобы он перечитал environment variables
	viper.Reset()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing required environment variables")
	}

}

func TestLoadConfigDefaults(t *testing.T) {
	// Сохраняем оригинальные значения
	originalBotToken := os.Getenv("BOT_TOKEN")
	originalDBURL := os.Getenv("DB_URL")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Устанавливаем только обязательные значения
	os.Setenv("BOT_TOKEN", "test_token")
	os.Setenv("DB_URL", "postgres://test:test@localhost/test")
	os.Unsetenv("LOG_LEVEL")

	// Сбрасываем Viper чтобы он перечитал environment variables
	viper.Reset()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Проверяем значения по умолчанию
	if config.LogLevel != "info" {
		t.Errorf("Expected default LOG_LEVEL 'info', got '%s'", config.LogLevel)
	}

	if config.CacheTTLMinutes != 5 {
		t.Errorf("Expected default CACHE_TTL_MINUTES 5, got %d", config.CacheTTLMinutes)
	}

	// Восстанавливаем оригинальные значения
	if originalBotToken != "" {
		os.Setenv("BOT_TOKEN", originalBotToken)
	} else {
		os.Unsetenv("BOT_TOKEN")
	}
	if originalDBURL != "" {
		os.Setenv("DB_URL", originalDBURL)
	} else {
		os.Unsetenv("DB_URL")
	}
	if originalLogLevel != "" {
		os.Setenv("LOG_LEVEL", originalLogLevel)
	}
}
