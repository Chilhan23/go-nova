package websocket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all CORS for multi-RS integration
	},
}

func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketIDStr := c.DefaultQuery("ticket_id", "0")
		key := c.DefaultQuery("key", "")
		ticketID, _ := strconv.Atoi(ticketIDStr)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WebSocket Upgrade Error]: %v", err)
			return
		}

		client := &Client{
			Hub:      hub,
			Conn:     conn,
			Send:     make(chan []byte, 256),
			TicketID: ticketID,
			Key:      key,
		}

		client.Hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
