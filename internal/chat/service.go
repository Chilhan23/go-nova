package chat

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tech-nova/internal/ai"
	"tech-nova/internal/config"
	"tech-nova/internal/telegram"
	"tech-nova/internal/ticket"
	"tech-nova/internal/websocket"
)

type Service interface {
	SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error)
	UploadAttachment(ctx context.Context, ticketID int, file *multipart.FileHeader, caption string) (*UploadAttachmentResponse, error)
	GetHistory(ctx context.Context, ticketID int) ([]MessageItemResponse, error)
	SaveProgrammerReply(ctx context.Context, ticketID int, senderName string, text string) error
}

type service struct {
	cfg        *config.Config
	chatRepo   Repository
	ticketRepo ticket.Repository
	aiService  ai.Service
	tgService  telegram.Service
	wsHub      *websocket.Hub
}

func NewService(
	cfg *config.Config,
	chatRepo Repository,
	ticketRepo ticket.Repository,
	aiService ai.Service,
	tgService telegram.Service,
	wsHub *websocket.Hub,
) Service {
	return &service{
		cfg:        cfg,
		chatRepo:   chatRepo,
		ticketRepo: ticketRepo,
		aiService:  aiService,
		tgService:  tgService,
		wsHub:      wsHub,
	}
}

func (s *service) SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	ticketObj, err := s.ticketRepo.GetByID(ctx, req.TicketID)
	if err != nil || ticketObj == nil {
		return nil, fmt.Errorf("ticket not found")
	}

	userMsg := &Message{
		TicketID:   ticketObj.ID,
		SenderType: "user",
		SenderName: &ticketObj.UserName,
		Message:    req.Message,
	}
	if err := s.chatRepo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	// 1. Create Forum Topic di Telegram jika belum ada
	if ticketObj.TelegramThreadID == nil || *ticketObj.TelegramThreadID == 0 {
		topicName := fmt.Sprintf("[%s] %s - %s", ticketObj.ModuleName, ticketObj.UserName, ticketObj.TicketCode)
		threadID, err := s.tgService.CreateForumTopic(ctx, topicName)
		if err == nil && threadID > 0 {
			_ = s.ticketRepo.UpdateThreadID(ctx, ticketObj.ID, threadID)
			ticketObj.TelegramThreadID = &threadID
		}
	}

	// 2. Forward ke Telegram Topic
	threadID := int64(0)
	if ticketObj.TelegramThreadID != nil {
		threadID = *ticketObj.TelegramThreadID
	}
	tgText := fmt.Sprintf("👤 <b>%s</b>:\n%s", ticketObj.UserName, req.Message)
	_ = s.tgService.SendMessage(ctx, tgText, threadID, nil)

	// 3. Jika ticket sudah dieskalasi ke programmer (human takeover), jangan panggil AI
	if ticketObj.Status == "escalated" || ticketObj.AssignedProgrammer != nil {
		return &SendMessageResponse{
			Status:    true,
			UserMsgID: userMsg.ID,
		}, nil
	}

	// 4. Panggil AI untuk First-Response
	historyMsgs, _ := s.chatRepo.GetByTicketID(ctx, ticketObj.ID)
	var aiHistory []ai.ContextMessage
	for _, m := range historyMsgs {
		role := "user"
		if m.SenderType == "ai" {
			role = "model"
		}
		aiHistory = append(aiHistory, ai.ContextMessage{
			Role:    role,
			Content: m.Message,
		})
	}

	aiReply, err := s.aiService.GenerateResponse(ctx, ai.AIRequest{
		Prompt:  req.Message,
		History: aiHistory,
	})
	if err != nil || aiReply == "" {
		aiReply = "Halo Kak! Pesan kendala Anda sudah diterima oleh sistem. Tim IT Support / Programmer kami akan segera membantu."
	}

	aiMsgName := "AI Support"
	aiMsg := &Message{
		TicketID:   ticketObj.ID,
		SenderType: "ai",
		SenderName: &aiMsgName,
		Message:    aiReply,
	}
	_ = s.chatRepo.Create(ctx, aiMsg)

	// Broadcast AI reply to WebSocket Room
	s.wsHub.BroadcastToTicket(ticketObj.ID, "helpdesk_new_message", map[string]interface{}{
		"id":          aiMsg.ID,
		"ticket_id":   ticketObj.ID,
		"sender_type": "ai",
		"sender_name": "AI Support",
		"message":     aiReply,
		"created_at":  aiMsg.CreatedAt,
	})

	return &SendMessageResponse{
		Status:    true,
		UserMsgID: userMsg.ID,
		AIMsgID:   &aiMsg.ID,
		Sender:    "ai",
		Reply:     aiReply,
	}, nil
}

func (s *service) UploadAttachment(ctx context.Context, ticketID int, file *multipart.FileHeader, caption string) (*UploadAttachmentResponse, error) {
	ticketObj, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticketObj == nil {
		return nil, fmt.Errorf("ticket not found")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	isVideo := ext == ".mp4" || ext == ".webm" || ext == ".mov" || ext == ".mkv"
	attType := "image"
	if isVideo {
		attType = "video"
	}

	_ = os.MkdirAll(s.cfg.UploadDir, 0755)
	filename := fmt.Sprintf("%s_%s_%d%s", attType, ticketObj.TicketCode, time.Now().Unix(), ext)
	targetPath := filepath.Join(s.cfg.UploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()
	_, _ = dst.ReadFrom(src)

	fileURL := fmt.Sprintf("%s/uploads/%s", s.cfg.BaseURL, filename)
	msgText := caption
	if msgText == "" {
		if isVideo {
			msgText = "[🎬 Video Kendala]"
		} else {
			msgText = "[📷 Screenshot Kendala]"
		}
	}

	userMsg := &Message{
		TicketID:       ticketObj.ID,
		SenderType:     "user",
		SenderName:     &ticketObj.UserName,
		Message:        msgText,
		IsAttachment:   true,
		AttachmentType: &attType,
		AttachmentURL:  &fileURL,
	}
	_ = s.chatRepo.Create(ctx, userMsg)

	// Create / Ensure Telegram Topic
	threadID := int64(0)
	if ticketObj.TelegramThreadID != nil {
		threadID = *ticketObj.TelegramThreadID
	} else {
		topicName := fmt.Sprintf("[%s] %s - %s", ticketObj.ModuleName, ticketObj.UserName, ticketObj.TicketCode)
		newThreadID, err := s.tgService.CreateForumTopic(ctx, topicName)
		if err == nil && newThreadID > 0 {
			_ = s.ticketRepo.UpdateThreadID(ctx, ticketObj.ID, newThreadID)
			threadID = newThreadID
		}
	}

	// Forward Media ke Telegram
	mediaCaption := fmt.Sprintf("👤 <b>%s</b> (#%s)\n📝 %s", ticketObj.UserName, ticketObj.TicketCode, caption)
	if isVideo {
		_ = s.tgService.SendVideo(ctx, targetPath, mediaCaption, threadID)
	} else {
		_ = s.tgService.SendPhoto(ctx, targetPath, mediaCaption, threadID)
	}

	// Response Balasan Otomatis
	aiReply := "Siap Kak, lampiran kendala sudah saya terima dan tersimpan dengan aman 👍. Tim IT Support / Programmer kami akan segera memeriksa rekaman kendala ini ya Kak."
	if !isVideo {
		aiReply = "Terima kasih Kak, tangkapan layar sudah kami terima. Tim Programmer akan segera menganalisa tampilan error tersebut."
	}

	aiMsgName := "AI Support"
	aiMsg := &Message{
		TicketID:   ticketObj.ID,
		SenderType: "ai",
		SenderName: &aiMsgName,
		Message:    aiReply,
	}
	_ = s.chatRepo.Create(ctx, aiMsg)

	// Broadcast WS
	s.wsHub.BroadcastToTicket(ticketObj.ID, "helpdesk_new_message", map[string]interface{}{
		"id":          aiMsg.ID,
		"ticket_id":   ticketObj.ID,
		"sender_type": "ai",
		"sender_name": "AI Support",
		"message":     aiReply,
		"created_at":  aiMsg.CreatedAt,
	})

	return &UploadAttachmentResponse{
		Status:    true,
		UserMsgID: userMsg.ID,
		AIMsgID:   &aiMsg.ID,
		Sender:    "ai",
		FileURL:   fileURL,
		Reply:     aiReply,
	}, nil
}

func (s *service) GetHistory(ctx context.Context, ticketID int) ([]MessageItemResponse, error) {
	msgs, err := s.chatRepo.GetByTicketID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	var result []MessageItemResponse
	for _, m := range msgs {
		senderName := ""
		if m.SenderName != nil {
			senderName = *m.SenderName
		}
		attType := ""
		if m.AttachmentType != nil {
			attType = *m.AttachmentType
		}
		attURL := ""
		if m.AttachmentURL != nil {
			attURL = *m.AttachmentURL
		}
		result = append(result, MessageItemResponse{
			ID:             m.ID,
			TicketID:       m.TicketID,
			SenderType:     m.SenderType,
			SenderName:     senderName,
			Message:        m.Message,
			IsAttachment:   m.IsAttachment,
			AttachmentType: attType,
			AttachmentURL:  attURL,
			CreatedAt:      m.CreatedAt,
		})
	}
	return result, nil
}

func (s *service) SaveProgrammerReply(ctx context.Context, ticketID int, senderName string, text string) error {
	progMsg := &Message{
		TicketID:   ticketID,
		SenderType: "programmer",
		SenderName: &senderName,
		Message:    text,
	}
	if err := s.chatRepo.Create(ctx, progMsg); err != nil {
		return err
	}

	s.wsHub.BroadcastToTicket(ticketID, "helpdesk_new_message", map[string]interface{}{
		"id":                  progMsg.ID,
		"ticket_id":           ticketID,
		"sender_type":         "programmer",
		"assigned_programmer": senderName,
		"message":             text,
		"created_at":          progMsg.CreatedAt,
	})
	return nil
}
