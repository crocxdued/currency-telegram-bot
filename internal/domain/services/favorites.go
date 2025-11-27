package services

import (
	"context"

	"github.com/crocxdued/currency-telegram-bot/internal/domain/entities"
)

// FavoritesRepository определяет операции для работы с избранными парами
type FavoritesRepository interface {
	AddFavorite(ctx context.Context, userID int64, fromCurrency, toCurrency string) error
	GetUserFavorites(ctx context.Context, userID int64) ([]entities.UserFavorite, error)
	RemoveFavorite(ctx context.Context, userID int64, fromCurrency, toCurrency string) error
}
