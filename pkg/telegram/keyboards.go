package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CreateMainKeyboard создает основную клавиатуру бота
func CreateMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💱 Конвертировать"),
			tgbotapi.NewKeyboardButton("⭐ Избранное"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Курсы валют"),
			tgbotapi.NewKeyboardButton("ℹ️ Помощь"),
		),
	)
}

// CreateInlineKeyboard создает инлайн-клавиатуру для избранного
func CreateInlineKeyboard(favorites []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < len(favorites); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в ряду
		if i < len(favorites) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(favorites[i], "favorite_"+favorites[i]))
		}

		// Вторая кнопка в ряду (если есть)
		if i+1 < len(favorites) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(favorites[i+1], "favorite_"+favorites[i+1]))
		}

		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// CreateCurrencyKeyboard создает клавиатуру для выбора валют
func CreateCurrencyKeyboard() tgbotapi.InlineKeyboardMarkup {
	// Популярные валюты для быстрого выбора
	currencies := []string{"USD", "EUR", "RUB", "GBP", "JPY", "CNY", "CAD", "CHF"}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(currencies); i += 4 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 4 && i+j < len(currencies); j++ {
			currency := currencies[i+j]
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(currency, "currency_"+currency))
		}
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// RemoveKeyboard удаляет кастомную клавиатуру
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}
