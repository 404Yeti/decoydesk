package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChatRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type TrapEvent struct {
	Service   string `json:"service"`
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details"`
}

var trapLoggerURL string

func sendTrapLog(event TrapEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("Failed to marshal trap event:", err)
		return
	}

	resp, err := http.Post(trapLoggerURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Failed to send trap log:", err)
		return
	}
	defer resp.Body.Close()
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Chat request from user: %s", req.UserID)

	if isMalicious(req.Message) {
		event := TrapEvent{
			Service:   "ai-chat-svc",
			Event:     "Prompt injection attempt",
			Timestamp: time.Now().Format(time.RFC3339),
			Details:   fmt.Sprintf("UserID: %s | Message: %s", req.UserID, req.Message),
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
		"ignore previous", "system prompt", "return all logs", "admin",
		"bypass", "print", "secrets", "env", "token", "curl", "wget",
	}
	msg = strings.ToLower(msg)
	for _, keyword := range injectionIndicators {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	trapLoggerURL = os.Getenv("TRAP_LOGGER_URL")
	if trapLoggerURL == "" {
		log.Fatal("TRAP_LOGGER_URL environment variable not set")
	}

	http.HandleFunc("/message", chatHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("ai-chat-svc listening on :8070")
	log.Fatal(http.ListenAndServe(":8070", nil))
}
