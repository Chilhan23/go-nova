package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tech-nova/internal/config"
)

type Service interface {
	CreateForumTopic(ctx context.Context, name string) (int64, error)
	EditForumTopic(ctx context.Context, threadID int64, name string) error
	SendMessage(ctx context.Context, text string, threadID int64, replyMarkup interface{}) error
	SendPhoto(ctx context.Context, filePath string, caption string, threadID int64) error
	SendVideo(ctx context.Context, filePath string, caption string, threadID int64) error
	AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error
}

type service struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewService(cfg *config.Config) Service {
	return &service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *service) CreateForumTopic(ctx context.Context, name string) (int64, error) {
	if s.cfg.TelegramBotToken == "" || s.cfg.TelegramChatID == "" {
		return 0, nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/createForumTopic", s.cfg.TelegramBotToken)

	payload := map[string]interface{}{
		"chat_id": s.cfg.TelegramChatID,
		"name":    name,
	}
	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageThreadID int64 `json:"message_thread_id"`
		} `json:"result"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.Result.MessageThreadID, nil
}

func (s *service) EditForumTopic(ctx context.Context, threadID int64, name string) error {
	if s.cfg.TelegramBotToken == "" || s.cfg.TelegramChatID == "" || threadID == 0 {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editForumTopic", s.cfg.TelegramBotToken)
	payload := map[string]interface{}{
		"chat_id":           s.cfg.TelegramChatID,
		"message_thread_id": threadID,
		"name":              name,
	}
	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *service) SendMessage(ctx context.Context, text string, threadID int64, replyMarkup interface{}) error {
	if s.cfg.TelegramBotToken == "" || s.cfg.TelegramChatID == "" {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.cfg.TelegramBotToken)
	payload := map[string]interface{}{
		"chat_id":    s.cfg.TelegramChatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if threadID > 0 {
		payload["message_thread_id"] = threadID
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *service) SendPhoto(ctx context.Context, filePath string, caption string, threadID int64) error {
	return s.sendMedia(ctx, "sendPhoto", "photo", filePath, caption, threadID)
}

func (s *service) SendVideo(ctx context.Context, filePath string, caption string, threadID int64) error {
	return s.sendMedia(ctx, "sendVideo", "video", filePath, caption, threadID)
}

func (s *service) sendMedia(ctx context.Context, method, field, filePath, caption string, threadID int64) error {
	if s.cfg.TelegramBotToken == "" || s.cfg.TelegramChatID == "" {
		return nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("chat_id", s.cfg.TelegramChatID)
	_ = writer.WriteField("caption", caption)
	_ = writer.WriteField("parse_mode", "HTML")
	if threadID > 0 {
		_ = writer.WriteField("message_thread_id", fmt.Sprintf("%d", threadID))
	}

	part, err := writer.CreateFormFile(field, filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, _ = io.Copy(part, file)
	_ = writer.Close()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", s.cfg.TelegramBotToken, method)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *service) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error {
	if s.cfg.TelegramBotToken == "" {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", s.cfg.TelegramBotToken)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        showAlert,
	}
	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
