package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type ChatRequest struct {
	UserID string `json:"user_id"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type TrapEvent struct {
	Service string `json:"service"`
	Event string `json:"event"`
	Timestamp string `json:"timestamp"`
	Details string `json:"details"`
}

func sendTrapLog(event TrapEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("Failed to marshal trap event:", err)
		return
	}

	resp, err := http.Post("http://localhost:8091/log", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Failed to send trap log:", err)
		return
	}
	defer resp.Body.Close()
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Chat request from user %s: %s", req.UserID, req.Message)

	if isMalicious(req.Message){
		event := TrapEvent{
			Service: "ai-chat-svc",
			Event: "Prompt injection attempt", 
			Timestamp: time.Now().Format(time.RFC3339),
			Details: fmt.Sprintf("UserID: %s | Message: %s", req.UserID, req.Message),
		}
		sendTrapLog(event)
	}

	response := ChatResponse{
		Response: "Thank you for your message. Your issue is being reviewed by an AI assistant.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isMalicious(msg string) bool {
	injectionIndicators := []string{
		"ignore previous", "system prompt", "return all logs", "admin", "bypass", "print", "secrets", "env",
		}
		msg = strings.ToLower(msg)
		for _, keyword := range injectionIndicators {
			if strings.Contains(msg, keyword) {
				return true
			}
		}
		return false
}

func main() {
	http.HandleFunc("/message", chatHandler)
	log.Println("ai-chat-svc listening on :8070")
	log.Fatal(http.ListenAndServe(":8070", nil))
}