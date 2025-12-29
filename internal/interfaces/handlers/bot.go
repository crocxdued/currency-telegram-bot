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

	// Вытаскиваем валюты для создания кнопок
	cleanText := strings.ToUpper(text)
	parts := strings.Fields(strings.ReplaceAll(cleanText, "/", " "))
	var currs []string
	for _, p := range parts {
		if len(p) == 3 {
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
func (h *BotHandler) parseAndConvert(userID int64, text string) (string, error) {
	ctx := context.Background()

	// Очистка: в верхний регистр, запятые в точки
	text = strings.ToUpper(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, ",", ".")

	// Разбиваем строку на части по пробелам и слэшам
	parts := strings.Fields(strings.ReplaceAll(text, "/", " "))

	var amount float64 = 1
	var currencies []string

	for _, p := range parts {
		if val, err := strconv.ParseFloat(p, 64); err == nil {
			amount = val
		} else if len(p) == 3 {
			currencies = append(currencies, p)
		}
	}

	if len(currencies) < 2 {
		return "", fmt.Errorf("укажите две валюты, например: `100 USD RUB`")
	}

	from, to := currencies[0], currencies[1]

	converted, err := h.exchangeService.ConvertAmount(ctx, amount, from, to)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💎 *Результат обмена*\n\n"))
	sb.WriteString(fmt.Sprintf("📤 *Отдаете:* %.2f %s\n", amount, from))
	sb.WriteString(fmt.Sprintf("📥 *Получаете:* %.2f %s\n", converted, to))
	sb.WriteString("───\n")
	sb.WriteString(fmt.Sprintf("📊 *Курс:* 1 %s = %.4f %s", from, converted/amount, to))

	return sb.String(), nil
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

	// Обработка кнопок конвертации (префикс conv_)
	if strings.HasPrefix(data, "conv_") {
		parts := strings.Split(data, "_") // conv, amount, from, to
		if len(parts) == 4 {
			amountStr := parts[1]
			from := parts[2]
			to := parts[3]

			// Делаем новый расчет
			result, err := h.parseAndConvert(userID, fmt.Sprintf("%s %s %s", amountStr, from, to))
			if err != nil {
				h.bot.Request(tgbotapi.NewCallback(callback.ID, "Ошибка"))
				return
			}

			// Редактируем сообщение
			editMsg := tgbotapi.NewEditMessageText(userID, messageID, result)
			editMsg.ParseMode = "Markdown"
			kb := h.createConversionKeyboard(from, to)
			editMsg.ReplyMarkup = &kb

			h.bot.Send(editMsg)
			h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
			return
		}
	}

	if strings.HasPrefix(data, "favorite_") {

	}

	h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

// sendMessage отправляет сообщение с обработкой ошибок
func (h *BotHandler) sendMessage(msg tgbotapi.MessageConfig) {
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *BotHandler) handleAddFavorite(message *tgbotapi.Message) {
	// 1. Разбираем текст сообщения формата "/fav_USD_RUB"
	// strings.Split разделяет строку по символу "_"
	parts := strings.Split(message.Text, "_")

	// Проверяем, что в команде достаточно частей (должно быть 3: "/fav", "USD", "RUB")
	if len(parts) < 3 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат. Используйте: /fav_USD_RUB")
		h.sendMessage(msg)
		return
	}

	// 2. ОЧИСТКА ДАННЫХ (Критически важно!)
	// strings.TrimSpace убирает лишние пробелы и символы переноса строки,
	// из-за которых возникала ошибка "currency not found".
	fromCurrency := strings.ToUpper(strings.TrimSpace(parts[1]))
	toCurrency := strings.ToUpper(strings.TrimSpace(parts[2]))

	// 3. Сохранение в базу данных
	ctx := context.Background()
	err := h.favoritesRepo.AddFavorite(ctx, message.Chat.ID, fromCurrency, toCurrency)
	if err != nil {
		// Логируем ошибку для отладки в консоль
		log.Printf("Error adding favorite: %v", err)

		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Не удалось сохранить пару в избранное.")
		h.sendMessage(msg)
		return
	}

	// 4. Уведомление пользователя об успехе
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
	)
}
