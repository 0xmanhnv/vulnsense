package adapter

import (
	"context"
	"log/slog"
	"time"

	"vulnsense/internal/config"
	"vulnsense/internal/usecase"

	"gopkg.in/telebot.v3"
)

// TelegramAlerter sends notifications to a Telegram chat.
type TelegramAlerter struct {
	bot    *telebot.Bot
	chatID int64
	logger *slog.Logger
}

// NewTelegramAlerter creates a new TelegramAlerter.
// It returns a "noop" alerter if the bot token or chat ID is not provided.
func NewTelegramAlerter(cfg config.TelegramConfig, logger *slog.Logger) usecase.Alerter {
	if cfg.BotToken == "" || cfg.ChatID == 0 {
		logger.Warn("Telegram bot token or chat ID is not configured. Alerter will be disabled.")
		return &NoOpAlerter{logger: logger}
	}

	pref := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		// Log the error but return a NoOpAlerter so the app can start without alerting.
		logger.Error("Failed to create Telegram bot, alerter will be disabled", "error", err)
		return &NoOpAlerter{logger: logger}
	}

	return &TelegramAlerter{
		bot:    bot,
		chatID: cfg.ChatID,
		logger: logger.With("component", "telegram_alerter"),
	}
}

// Notify sends a message to the configured Telegram chat.
func (a *TelegramAlerter) Notify(ctx context.Context, message string) error {
	chat := &telebot.Chat{ID: a.chatID}
	_, err := a.bot.Send(chat, message, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	if err != nil {
		a.logger.Error("Failed to send telegram message", "error", err)
	}
	return err
}

// NoOpAlerter is an alerter that does nothing.
// It's used when the real alerter is not configured.
type NoOpAlerter struct {
	logger *slog.Logger
}

// Notify logs the message instead of sending it.
func (a *NoOpAlerter) Notify(ctx context.Context, message string) error {
	a.logger.Info("Alerter is not configured. Dropping message.", "message", message)
	return nil
}
