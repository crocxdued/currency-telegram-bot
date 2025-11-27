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
	userStates      map[int64]string // простой state management
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

	// Пытаемся распарсить запрос на конвертацию
	result, err := h.parseAndConvert(message.Chat.ID, text)
	if err != nil {
		msg := tgbotapi.NewMessage(userID, fmt.Sprintf("❌ Ошибка: %s\n\nПопробуйте в формате:\n`100 USD to RUB`", err.Error()))
		msg.ParseMode = "Markdown"
		h.sendMessage(msg)
		return
	}

	msg := tgbotapi.NewMessage(userID, result)
	h.sendMessage(msg)
}

// parseAndConvert парсит и выполняет конвертацию
func (h *BotHandler) parseAndConvert(userID int64, text string) (string, error) {
	ctx := context.Background()

	// Парсим разные форматы: "100 USD to RUB", "EUR/RUB", "100.5 EUR USD"
	var amount float64 = 1
	var from, to string

	// Формат: "100 USD to RUB"
	if parts := strings.Split(text, " "); len(parts) >= 4 {
		if parsedAmount, err := strconv.ParseFloat(parts[0], 64); err == nil {
			amount = parsedAmount
			from = parts[1]
			to = parts[3]
		}
	}

	// Формат: "EUR/RUB" или "100 EUR/RUB"
	if from == "" {
		if strings.Contains(text, "/") {
			parts := strings.Split(text, " ")
			if len(parts) == 1 {
				// "EUR/RUB"
				currencyParts := strings.Split(text, "/")
				if len(currencyParts) == 2 {
					from = currencyParts[0]
					to = currencyParts[1]
				}
			} else if len(parts) == 2 {
				// "100 EUR/RUB"
				if parsedAmount, err := strconv.ParseFloat(parts[0], 64); err == nil {
					amount = parsedAmount
					currencyParts := strings.Split(parts[1], "/")
					if len(currencyParts) == 2 {
						from = currencyParts[0]
						to = currencyParts[1]
					}
				}
			}
		}
	}

	if from == "" || to == "" {
		return "", fmt.Errorf("не удалось распознать запрос")
	}

	// Выполняем конвертацию
	converted, err := h.exchangeService.ConvertAmount(ctx, amount, from, to)
	if err != nil {
		return "", fmt.Errorf("ошибка конвертации: %s", err.Error())
	}

	// Форматируем результат
	result := fmt.Sprintf("💱 *%.2f %s* = *%.2f %s*", amount, from, converted, to)

	// Предлагаем добавить в избранное
	result += fmt.Sprintf("\n\n⭐ Добавить в избранное: /fav_%s_%s", from, to)

	return result, nil
}

// handleHelp показывает справку
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
	ctx := context.Background()
	favorites, err := h.favoritesRepo.GetUserFavorites(ctx, message.Chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при загрузке избранного")
		h.sendMessage(msg)
		return
	}

	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⭐ У вас пока нет избранных пар валют.\n\nДобавьте их с помощью команды:\n/fav_USD_EUR")
		h.sendMessage(msg)
		return
	}

	var favoritePairs []string
	for _, fav := range favorites {
		favoritePairs = append(favoritePairs, fmt.Sprintf("%s/%s", fav.FromCurrency, fav.ToCurrency))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "⭐ Ваши избранные пары:")
	h.sendMessage(msg)

	// Отправляем инлайн-клавиатуру с избранными парами
	keyboardMsg := tgbotapi.NewMessage(message.Chat.ID, "Выберите пару для конвертации:")
	keyboardMsg.ReplyMarkup = telegram.CreateInlineKeyboard(favoritePairs)
	h.sendMessage(keyboardMsg)
}

// handleRates показывает текущие курсы
func (h *BotHandler) handleRates(message *tgbotapi.Message) {
	ctx := context.Background()

	// Получаем курсы для популярных пар
	pairs := [][2]string{
		{"USD", "RUB"},
		{"EUR", "RUB"},
		{"USD", "EUR"},
		{"GBP", "USD"},
	}

	var ratesText strings.Builder
	ratesText.WriteString("📊 *Текущие курсы:*\n\n")

	for _, pair := range pairs {
		rate, err := h.exchangeService.GetRate(ctx, pair[0], pair[1])
		if err != nil {
			continue
		}
		ratesText.WriteString(fmt.Sprintf("💱 *%s/%s:* %.4f\n", pair[0], pair[1], rate))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, ratesText.String())
	msg.ParseMode = "Markdown"
	h.sendMessage(msg)
}

// handleCallback обрабатывает нажатия на инлайн-кнопки
func (h *BotHandler) handleCallback(callback *tgbotapi.CallbackQuery) {
	// TODO: Реализовать обработку инлайн-кнопок
	callbackConfig := tgbotapi.NewCallback(callback.ID, "Функция в разработке")
	if _, err := h.bot.Request(callbackConfig); err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

// sendMessage отправляет сообщение с обработкой ошибок
func (h *BotHandler) sendMessage(msg tgbotapi.MessageConfig) {
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
