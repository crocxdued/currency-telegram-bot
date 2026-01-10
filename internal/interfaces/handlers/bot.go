package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/crocxdued/currency-telegram-bot/internal/domain/services"
	"github.com/crocxdued/currency-telegram-bot/pkg/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot             *tgbotapi.BotAPI
	exchangeService services.ExchangeService
	favoritesRepo   services.FavoritesRepository
	userStates      map[int64]string
}

func NewBotHandler(
	bot *tgbotapi.BotAPI,
	exchangeService services.ExchangeService,
	favoritesRepo services.FavoritesRepository,
) *BotHandler {
	return &BotHandler{
		bot:             bot,
		exchangeService: exchangeService,
		favoritesRepo:   favoritesRepo,
		userStates:      make(map[int64]string),
	}
}

// HandleUpdate обрабатывает входящие сообщения
func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		h.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
	}
}

// handleMessage обрабатывает текстовые сообщения
func (h *BotHandler) handleMessage(message *tgbotapi.Message) {
	text := message.Text

	if strings.HasPrefix(text, "/fav_") {
		h.handleAddFavorite(message)
		return
	}

	switch text {
	case "/start":
		h.handleStart(message)
	case "/help", "ℹ️ Помощь":
		h.handleHelp(message)
	case "💱 Конвертировать":
		h.handleConvert(message)
	case "⭐ Избранное":
		h.handleFavorites(message)
	case "📊 Курсы валют":
		h.handleRates(message)
	default:
		h.handleText(message)
	}
}

// handleStart приветственное сообщение
func (h *BotHandler) handleStart(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, `
🤖 *Currency Exchange Bot*

Я помогу вам:
💱 Конвертировать валюты
⭐ Сохранять избранные пары  
📊 Смотреть актуальные курсы

*Примеры использования:*
• 100 USD to RUB
• EUR/RUB
• 50.5 EUR USD

Используйте кнопки ниже или введите запрос вручную!`)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = telegram.CreateMainKeyboard()

	h.sendMessage(msg)
}

// handleConvert начинает процесс конвертации
func (h *BotHandler) handleConvert(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите запрос в формате:\n`100 USD to RUB`\nили\n`EUR/RUB`")
	msg.ParseMode = "Markdown"

	h.sendMessage(msg)
	h.userStates[message.Chat.ID] = "converting"
}

// handleText обрабатывает произвольный текст для конвертации
func (h *BotHandler) handleText(message *tgbotapi.Message) {
	text := strings.TrimSpace(message.Text)
	userID := message.Chat.ID

	result, err := h.parseAndConvert(userID, text)
	if err != nil {
		msg := tgbotapi.NewMessage(userID, "❌ "+err.Error())
		msg.ParseMode = "Markdown"
		h.sendMessage(msg)
		return
	}

	cleanText := strings.ToUpper(text)
	parts := strings.Fields(strings.ReplaceAll(cleanText, "/", " "))
	var currs []string
	for _, p := range parts {

		if _, err := strconv.ParseFloat(p, 64); err != nil && len(p) == 3 {
			currs = append(currs, p)
		}
	}

	msg := tgbotapi.NewMessage(userID, result)
	msg.ParseMode = "Markdown"

	if len(currs) >= 2 {
		msg.ReplyMarkup = h.createConversionKeyboard(currs[0], currs[1])
	}

	h.sendMessage(msg)
}

// parseAndConvert парсит и выполняет конвертацию
func (h *BotHandler) parseAndConvert(_ int64, text string) (string, error) {
	ctx := context.Background()
	text = strings.ToUpper(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "/", " ")
	text = strings.ReplaceAll(text, ",", ".")

	parts := strings.Fields(text)
	var amount float64 = 1
	var currencies []string

	for _, p := range parts {
		if val, err := strconv.ParseFloat(p, 64); err == nil {
			amount = val
		} else {

			cleanCurr := strings.TrimFunc(p, func(r rune) bool {
				return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
			})

			if len(cleanCurr) == 3 {
				currencies = append(currencies, cleanCurr)
			}
		}
	}

	if len(currencies) < 2 {
		return "", fmt.Errorf("нужно 2 валюты (напр. USD RUB)")
	}

	from, to := currencies[0], currencies[1]
	converted, err := h.exchangeService.ConvertAmount(ctx, amount, from, to)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("💎 *Результат обмена*\n\n") // Ошибка S1039 исправлена (убран fmt.Sprintf)
	sb.WriteString(fmt.Sprintf("📤 *Отдаете:* %.2f %s\n", amount, from))
	sb.WriteString(fmt.Sprintf("📥 *Получаете:* %.2f %s\n", converted, to))
	sb.WriteString("───\n")
	sb.WriteString(fmt.Sprintf("📊 *Курс:* 1 %s = %.4f %s", from, converted/amount, to))

	return sb.String(), nil
}

func (h *BotHandler) handleHelp(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, `
*📖 Справка по использованию бота*

*Основные команды:*
/start - начать работу
/help - эта справка

*Форматы запросов:*
• 100 USD to RUB
• EUR/RUB  
• 50.5 EUR USD

*Избранное:*
Добавляйте часто используемые пары в избранное для быстрого доступа!`)
	msg.ParseMode = "Markdown"

	h.sendMessage(msg)
}

// handleFavorites показывает избранное пользователя
func (h *BotHandler) handleFavorites(message *tgbotapi.Message) {
	userID := message.Chat.ID
	ctx := context.Background()

	favorites, err := h.favoritesRepo.GetUserFavorites(ctx, userID)
	if err != nil {
		log.Printf("Error getting favorites: %v", err)
		h.sendMessage(tgbotapi.NewMessage(userID, "❌ Не удалось загрузить список избранного."))
		return
	}

	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(userID, "🌟 У вас пока нет избранных пар.\n\nЧтобы добавить, отправьте команду: `/fav_USD_RUB` или воспользуйтесь кнопкой «В избранное» после конвертации.")
		msg.ParseMode = "Markdown"
		h.sendMessage(msg)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, fav := range favorites {
		pairText := fmt.Sprintf("%s/%s", fav.FromCurrency, fav.ToCurrency)

		// Создаем строку из двух кнопок:
		// Первая кнопка — сама пара (при нажатии сработает как текст "USD/RUB")
		// Вторая кнопка — корзина (триггерит удаление через callback)
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(pairText, pairText),
			tgbotapi.NewInlineKeyboardButtonData("🗑️", fmt.Sprintf("remfav_%s_%s", fav.FromCurrency, fav.ToCurrency)),
		)
		rows = append(rows, row)
	}

	msg := tgbotapi.NewMessage(userID, "⭐ *Ваши избранные пары:*\nНажмите на пару для быстрого расчета или на корзину для удаления.")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	h.sendMessage(msg)
}

func (h *BotHandler) handleRates(message *tgbotapi.Message) {
	ctx := context.Background()
	pairs := [][2]string{
		{"USD", "RUB"},
		{"EUR", "RUB"},
		{"CNY", "RUB"}, // Юань
		{"TRY", "RUB"}, // Лира
		{"KZT", "RUB"}, // Тенге
		{"USD", "EUR"}, // Евро/Доллар
		{"AED", "RUB"}, // Дирхам
	}

	var ratesText strings.Builder
	ratesText.WriteString("📊 *Текущие курсы:*\n\n")

	found := false
	for _, pair := range pairs {
		rate, err := h.exchangeService.GetRate(ctx, pair[0], pair[1])
		if err != nil {
			log.Printf("LOG: Ошибка для %s/%s: %v", pair[0], pair[1], err)
			continue
		}
		found = true
		ratesText.WriteString(fmt.Sprintf("💱 *%s/%s:* %.4f\n", pair[0], pair[1], rate))
	}

	if !found {
		ratesText.WriteString("❌ Сервисы временно недоступны.")
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, ratesText.String())
	msg.ParseMode = "Markdown"
	h.sendMessage(msg)
}

// handleCallback обрабатывает нажатия на инлайн-кнопки
func (h *BotHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	userID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID

	if strings.Contains(data, "/") {

		cleanData := data
		if idx := strings.Index(data, "_"); idx != -1 && !strings.HasPrefix(data, "conv_") {

			cleanData = data[idx+1:]
		}

		result, err := h.parseAndConvert(userID, cleanData)
		if err != nil {
			_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, "❌ "+err.Error()))
			return
		}

		msg := tgbotapi.NewMessage(userID, result)
		msg.ParseMode = "Markdown"

		parts := strings.Split(cleanData, "/")
		if len(parts) == 2 {
			msg.ReplyMarkup = h.createConversionKeyboard(parts[0], parts[1])
		}

		h.sendMessage(msg)
		_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
		return
	}

	if strings.HasPrefix(data, "conv_") {
		parts := strings.Split(data, "_")
		if len(parts) == 4 {
			amountStr := parts[1]
			from := parts[2]
			to := parts[3]

			result, err := h.parseAndConvert(userID, fmt.Sprintf("%s %s %s", amountStr, from, to))
			if err != nil {
				_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, "Ошибка"))
				return
			}

			editMsg := tgbotapi.NewEditMessageText(userID, messageID, result)
			editMsg.ParseMode = "Markdown"
			kb := h.createConversionKeyboard(from, to)
			editMsg.ReplyMarkup = &kb

			_, _ = h.bot.Send(editMsg)
			_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
			return
		}
	}

	if strings.HasPrefix(data, "addfav_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			from := parts[1]
			to := parts[2]
			ctx := context.Background()

			err := h.favoritesRepo.AddFavorite(ctx, userID, from, to)

			var callbackText string
			if err != nil {
				callbackText = "❌ Ошибка при сохранении"
			} else {
				callbackText = fmt.Sprintf("✅ Пара %s/%s добавлена в избранное!", from, to)
			}

			callbackCfg := tgbotapi.NewCallback(callback.ID, callbackText)
			_, _ = h.bot.Request(callbackCfg)
			return
		}
	}

	if strings.HasPrefix(data, "remfav_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			from, to := parts[1], parts[2]

			err := h.favoritesRepo.RemoveFavorite(context.Background(), userID, from, to)

			var text string
			if err != nil {
				text = "❌ Ошибка при удалении"
			} else {
				text = fmt.Sprintf("🗑️ %s/%s удалено из избранного", from, to)
			}

			callbackCfg := tgbotapi.NewCallback(callback.ID, text)
			_, _ = h.bot.Request(callbackCfg)
			return
		}
	}

	_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

// sendMessage отправляет сообщение с обработкой ошибок
func (h *BotHandler) sendMessage(msg tgbotapi.MessageConfig) {
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *BotHandler) handleAddFavorite(message *tgbotapi.Message) {

	parts := strings.Split(message.Text, "_")

	if len(parts) < 3 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат. Используйте: /fav_USD_RUB")
		h.sendMessage(msg)
		return
	}

	fromCurrency := strings.ToUpper(strings.TrimSpace(parts[1]))
	toCurrency := strings.ToUpper(strings.TrimSpace(parts[2]))

	ctx := context.Background()
	err := h.favoritesRepo.AddFavorite(ctx, message.Chat.ID, fromCurrency, toCurrency)
	if err != nil {

		log.Printf("Error adding favorite: %v", err)

		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось сохранить пару в избранное.")
		h.sendMessage(msg)
		return
	}

	successText := fmt.Sprintf("✅ Пара *%s/%s* добавлена в ваше избранное!", fromCurrency, toCurrency)
	msg := tgbotapi.NewMessage(message.Chat.ID, successText)
	msg.ParseMode = "Markdown"
	h.sendMessage(msg)
}

func (h *BotHandler) createConversionKeyboard(from, to string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("10 "+from, fmt.Sprintf("conv_10_%s_%s", from, to)),
			tgbotapi.NewInlineKeyboardButtonData("100 "+from, fmt.Sprintf("conv_100_%s_%s", from, to)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("500 "+from, fmt.Sprintf("conv_500_%s_%s", from, to)),
			tgbotapi.NewInlineKeyboardButtonData("1000 "+from, fmt.Sprintf("conv_1000_%s_%s", from, to)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Обратный курс ("+to+"/"+from+")", fmt.Sprintf("conv_1_%s_%s", to, from)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Добавить "+from+"/"+to+" в избранное", fmt.Sprintf("addfav_%s_%s", from, to)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("remfav_%s_%s", from, to)),
		),
	)
}
