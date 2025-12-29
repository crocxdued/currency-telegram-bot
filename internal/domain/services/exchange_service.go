package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/crocxdued/currency-telegram-bot/internal/interfaces/repository/cache"
)

type ExchangeServiceImpl struct {
	providers []ExchangeProvider
	cache     *cache.RatesCache
}

func NewExchangeService(providers []ExchangeProvider, cache *cache.RatesCache) *ExchangeServiceImpl {
	return &ExchangeServiceImpl{
		providers: providers,
		cache:     cache,
	}
}

// GetRate получает курс обмена с использованием кэша и fallback провайдеров
func (s *ExchangeServiceImpl) GetRate(ctx context.Context, from, to string) (float64, error) {
	// 1. Максимально жесткая очистка входящих строк
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	// 2. Проверка на пустые строки (чтобы не ходить в кэш с мусором)
	if from == "" || to == "" {
		return 0, fmt.Errorf("invalid currency codes: from='%s', to='%s'", from, to)
	}

	// Пробуем получить курс от доступных провайдеров
	var lastErr error
	for _, provider := range s.providers {
		if !provider.IsAvailable() {
			continue
		}

		rate, err := provider.GetRate(ctx, from, to)
		if err != nil {
			lastErr = err
			continue
		}

		// Сохраняем в кэш при успешном получении
		s.cache.Set(from, to, rate)
		return rate, nil
	}

	return 0, fmt.Errorf("failed to get exchange rate: %w", lastErr)
}

// ConvertAmount конвертирует сумму из одной валюты в другую
func (s *ExchangeServiceImpl) ConvertAmount(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

// GetSupportedCurrencies возвращает список поддерживаемых валют
func (s *ExchangeServiceImpl) GetSupportedCurrencies(ctx context.Context) (map[string]string, error) {
	// Основные валюты, поддерживаемые нашим сервисом
	currencies := map[string]string{
		"USD": "United States Dollar",
		"EUR": "Euro",
		"RUB": "Russian Ruble",
		"GBP": "British Pound",
		"JPY": "Japanese Yen",
		"CNY": "Chinese Yuan",
		"CAD": "Canadian Dollar",
		"CHF": "Swiss Franc",
		"AUD": "Australian Dollar",
		"TRY": "Turkish Lira",
		"KZT": "Kazakhstani Tenge",
		"UAH": "Ukrainian Hryvnia",
		"BYN": "Belarusian Ruble",
	}

	return currencies, nil
}
