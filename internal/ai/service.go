package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"tech-nova/internal/config"
)

type Service interface {
	GenerateResponse(ctx context.Context, req AIRequest) (string, error)
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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *service) GenerateResponse(ctx context.Context, req AIRequest) (string, error) {
	if s.cfg.AIAPIKey == "" {
		return "Halo! Pesan kendala Anda telah kami terima. Tim Programmer / IT Support kami akan segera merespons kendala Anda ya 👍.", nil
	}

	endpoint := s.cfg.AIEndpoint
	if endpoint == "" {
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	systemPrompt := "Kamu adalah AI IT Support dan Customer Care yang ramah, profesional, dan membantu pengguna menyelesaikan kendala teknis aplikasi. Selalu gunakan sapaan 'Kak' dan bahasa Indonesia yang santun."

	var messages []chatMessage
	messages = append(messages, chatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, h := range req.History {
		role := "user"
		if h.Role == "ai" || h.Role == "assistant" || h.Role == "model" {
			role = "assistant"
		}
		messages = append(messages, chatMessage{
			Role:    role,
			Content: h.Content,
		})
	}

	messages = append(messages, chatMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	model := s.cfg.AIModel
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	reqBody := openAIRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AIAPIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 && result.Choices[0].Message.Content != "" {
		return result.Choices[0].Message.Content, nil
	}

	return "Siap Kak, kendala sudah kami catat dan tim kami akan segera menindaklanjuti.", nil
}
