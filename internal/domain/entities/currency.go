package entities

import (
	"time"
)

// Currency представляет валюту с кодом и названием
type Currency struct {
	Code string
	Name string
}

// ExchangeRate представляет курс обмена между двумя валютами
type ExchangeRate struct {
	From        Currency
	To          Currency
	Rate        float64
	LastUpdated time.Time
}

// UserFavorite представляет избранную пару валют пользователя
type UserFavorite struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	FromCurrency string    `db:"from_currency"`
	ToCurrency   string    `db:"to_currency"`
	CreatedAt    time.Time `db:"created_at"`
}
