package telegram

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"tech-nova/internal/ticket"
	"tech-nova/internal/websocket"
)

type MessageRelayer interface {
	SaveProgrammerReply(ctx context.Context, ticketID int, senderName string, text string) error
}

type Handler struct {
	tgService      Service
	ticketRepo     ticket.Repository
	messageRelayer MessageRelayer
	wsHub          *websocket.Hub
}

func NewHandler(
	tgService Service,
	ticketRepo ticket.Repository,
	messageRelayer MessageRelayer,
	wsHub *websocket.Hub,
) *Handler {
	return &Handler{
		tgService:      tgService,
		ticketRepo:     ticketRepo,
		messageRelayer: messageRelayer,
		wsHub:          wsHub,
	}
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	var update TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false})
		return
	}

	ctx := c.Request.Context()

	// 1. Handle Callback Query (Klik tombol Klaim / Resolve)
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		cbData := cb.Data
		senderName := cb.From.FirstName
		if cb.From.LastName != "" {
			senderName += " " + cb.From.LastName
		}

		if strings.HasPrefix(cbData, "claim_") {
			ticketIDStr := strings.TrimPrefix(cbData, "claim_")
			ticketID, _ := strconv.Atoi(ticketIDStr)
			claimed, _ := h.ticketRepo.UpdateProgrammer(ctx, ticketID, senderName)
			if claimed {
				_ = h.tgService.AnswerCallbackQuery(ctx, cb.ID, "Berhasil! Tiket resmi diklaim oleh Anda.", true)
				tObj, _ := h.ticketRepo.GetByID(ctx, ticketID)
				if tObj != nil && tObj.TelegramThreadID != nil {
					notice := fmt.Sprintf("👨‍💻 <b>TIKET DIKLAIM</b>\nTiket <code>#%s</code> diambil alih oleh <b>%s</b>.", tObj.TicketCode, senderName)
					_ = h.tgService.SendMessage(ctx, notice, *tObj.TelegramThreadID, nil)

					// Broadcast status changed to WebSocket
					h.wsHub.BroadcastToTicket(tObj.ID, "helpdesk_status_changed", map[string]interface{}{
						"ticket_id":           tObj.ID,
						"ticket_status":       "escalated",
						"assigned_programmer": senderName,
					})
				}
			} else {
				_ = h.tgService.AnswerCallbackQuery(ctx, cb.ID, "Tiket sudah diambil oleh programmer lain.", true)
			}
		} else if strings.HasPrefix(cbData, "resolve_") {
			ticketIDStr := strings.TrimPrefix(cbData, "resolve_")
			ticketID, _ := strconv.Atoi(ticketIDStr)
			_ = h.ticketRepo.UpdateStatus(ctx, ticketID, "waiting_user")
			_ = h.tgService.AnswerCallbackQuery(ctx, cb.ID, "Permintaan penyelesaian terkirim ke pengguna!", true)

			// Broadcast status changed to WebSocket
			h.wsHub.BroadcastToTicket(ticketID, "helpdesk_status_changed", map[string]interface{}{
				"ticket_id":     ticketID,
				"ticket_status": "waiting_user",
			})
		}
		c.JSON(http.StatusOK, gin.H{"status": true})
		return
	}

	// 2. Handle Pesan Teks / Balasan dari Programmer di Forum Topic
	if update.Message != nil && update.Message.MessageThreadID > 0 {
		msg := update.Message
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			text = strings.TrimSpace(msg.Caption)
		}

		if text != "" {
			tObj, err := h.ticketRepo.GetByThreadID(ctx, msg.MessageThreadID)
			if err == nil && tObj != nil {
				senderName := "Tim IT Support"
				if msg.From != nil {
					senderName = msg.From.FirstName
					if msg.From.LastName != "" {
						senderName += " " + msg.From.LastName
					}
				}

				if h.messageRelayer != nil {
					_ = h.messageRelayer.SaveProgrammerReply(ctx, tObj.ID, senderName, text)
				}
				log.Printf("📨 [Relay] Programmer reply dispatched to Ticket %d: %s", tObj.ID, text)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}
