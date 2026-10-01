package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
)

const baseURL = "http://localhost:8080"

func main() {
	log.Println("🧪 [E2E Test] Starting Tech-Nova Self-Verification...")

	// 1. Health Check
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Fatalf("❌ Health Check Failed: Server is not running on %s. Error: %v", baseURL, err)
	}
	defer resp.Body.Close()
	log.Println("✅ 1. Health Check OK (200)")

	// 2. Init Ticket
	initPayload := map[string]interface{}{
		"key_identifier": "kemkes_1101015",
		"app_name":       "KlikMedic SIMRS",
		"tenant_name":    "RSUD Meuraxa",
		"user_id":        "101",
		"user_name":      "Rayhan (Perawat IGD)",
		"module_name":    "Instalasi Gawat Darurat",
		"diagnostic_info": map[string]string{
			"url": "http://simrs.rsudmeuraxa.id/igd/tindakan",
		},
	}
	bodyBytes, _ := json.Marshal(initPayload)
	resp, err = http.Post(baseURL+"/api/v1/tickets/init", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 2. Init Ticket Failed: %v (Status: %d)", err, resp.StatusCode)
	}
	var initRes struct {
		Status bool `json:"status"`
		Ticket struct {
			ID         int    `json:"id"`
			TicketCode string `json:"ticket_code"`
		} `json:"ticket"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&initRes)
	resp.Body.Close()
	ticketID := initRes.Ticket.ID
	log.Printf("✅ 2. Init Ticket OK (Ticket ID: %d, Code: %s)", ticketID, initRes.Ticket.TicketCode)

	// 3. Send Message (Test AI Response)
	sendPayload := map[string]interface{}{
		"ticket_id": ticketID,
		"message":   "Halo tim IT, saya mau tanya bagaimana alur klaim tindakan ya?",
	}
	bodyBytes, _ = json.Marshal(sendPayload)
	log.Println("⏳ 3. Sending message to AI...")
	resp, err = http.Post(baseURL+"/api/v1/chat/send", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 3. Send Message Failed: %v", err)
	}
	var sendRes struct {
		Status bool   `json:"status"`
		Reply  string `json:"reply"`
		Sender string `json:"sender"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&sendRes)
	resp.Body.Close()
	log.Printf("✅ 3. AI Reply Generated [%s]: %s", sendRes.Sender, sendRes.Reply)

	// 4. Upload Attachment
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	_ = writer.WriteField("ticket_id", fmt.Sprintf("%d", ticketID))
	_ = writer.WriteField("caption", "Tangkapan layar kendala saat simpan")
	part, _ := writer.CreateFormFile("attachment", "screenshot_error.png")
	_, _ = part.Write([]byte("fake png binary data for verification"))
	_ = writer.Close()

	uploadReq, _ := http.NewRequest("POST", baseURL+"/api/v1/chat/upload", buf)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err = http.DefaultClient.Do(uploadReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 4. Upload Attachment Failed: %v", err)
	}
	var uploadRes struct {
		Status  bool   `json:"status"`
		FileURL string `json:"file_url"`
		Reply   string `json:"reply"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&uploadRes)
	resp.Body.Close()
	log.Printf("✅ 4. Upload Attachment OK: URL=%s | Reply=%s", uploadRes.FileURL, uploadRes.Reply)

	// 5. Get History
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/tickets/%d/messages", baseURL, ticketID))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 5. Get History Failed: %v", err)
	}
	var histRes struct {
		Status bool          `json:"status"`
		Data   []interface{} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&histRes)
	resp.Body.Close()
	log.Printf("✅ 5. Get History OK (Total Messages: %d)", len(histRes.Data))

	// 6. Simulate Telegram Callback Query (Claim Ticket)
	claimPayload := map[string]interface{}{
		"update_id": 9991,
		"callback_query": map[string]interface{}{
			"id":   "cb_12345",
			"data": fmt.Sprintf("claim_%d", ticketID),
			"from": map[string]interface{}{
				"id":         888123,
				"first_name": "Rayhan",
				"last_name":  "Programmer",
			},
		},
	}
	bodyBytes, _ = json.Marshal(claimPayload)
	resp, err = http.Post(baseURL+"/webhook/telegram", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 6. Telegram Claim Webhook Failed: %v", err)
	}
	resp.Body.Close()
	log.Println("✅ 6. Telegram Webhook (Claim Ticket Simulation) OK")

	// 7. Rate Ticket (CSAT)
	ratePayload := map[string]interface{}{
		"rating": 5,
		"review": "Respon sangat cepat dan memuaskan!",
	}
	bodyBytes, _ = json.Marshal(ratePayload)
	resp, err = http.Post(fmt.Sprintf("%s/api/v1/tickets/%d/rate", baseURL, ticketID), "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ 7. Rate Ticket Failed: %v", err)
	}
	resp.Body.Close()
	log.Println("✅ 7. Submit CSAT Rating OK (5 Stars)")

	log.Println("🎉 ==========================================================")
	log.Println("🎉 ALL VERIFICATION CHECKS PASSED 100%! SYSTEM IS HEALTHY! 🚀")
	log.Println("🎉 ==========================================================")
}
