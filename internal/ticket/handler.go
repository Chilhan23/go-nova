package ticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"tech-nova/internal/tenant"
)

type Handler struct {
	ticketService Service
	tenantService tenant.Service
}

func NewHandler(ticketService Service, tenantService tenant.Service) *Handler {
	return &Handler{
		ticketService: ticketService,
		tenantService: tenantService,
	}
}

func (h *Handler) InitTicket(c *gin.Context) {
	var req InitTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	t, err := h.tenantService.GetOrCreate(c.Request.Context(), tenant.RegisterTenantRequest{
		KeyIdentifier: req.KeyIdentifier,
		AppName:       req.AppName,
		TenantName:    req.TenantName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	ticketObj, err := h.ticketService.GetOrCreateActive(c.Request.Context(), req, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"ticket": ticketObj,
	})
}

func (h *Handler) RateTicket(c *gin.Context) {
	ticketIDStr := c.Param("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Invalid ticket ID"})
		return
	}

	var req RateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	if err := h.ticketService.Rate(c.Request.Context(), ticketID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Terima kasih atas penilaian Anda!",
	})
}
