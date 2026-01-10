package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/crocxdued/currency-telegram-bot/internal/domain/entities"
	"github.com/jmoiron/sqlx"
)

type FavoritesRepository struct {
	db *sqlx.DB
}

func NewFavoritesRepository(db *sqlx.DB) *FavoritesRepository {
	return &FavoritesRepository{db: db}
}

func (r *FavoritesRepository) AddFavorite(ctx context.Context, userID int64, fromCurrency, toCurrency string) error {
	query := `
		INSERT INTO user_favorites (user_id, from_currency, to_currency)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, from_currency, to_currency) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, fromCurrency, toCurrency)
	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

func (r *FavoritesRepository) GetUserFavorites(ctx context.Context, userID int64) ([]entities.UserFavorite, error) {
	var favorites []entities.UserFavorite

	query := `
		SELECT id, user_id, from_currency, to_currency, created_at
		FROM user_favorites 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &favorites, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user favorites: %w", err)
	}

	return favorites, nil
}

func (r *FavoritesRepository) RemoveFavorite(ctx context.Context, userID int64, fromCurrency, toCurrency string) error {
	query := `
		DELETE FROM user_favorites 
		WHERE user_id = $1 AND from_currency = $2 AND to_currency = $3
	`

	result, err := r.db.ExecContext(ctx, query, userID, fromCurrency, toCurrency)
	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
