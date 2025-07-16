package main 
import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"bytes"
)

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	event := TrapEvent {
		Service: "auth-svc",
		Event: "Login attempt",
		Timestamp: time.Now().Format(time.RFC3339),
		Details: "Email: " + creds.Email + "| Password: " + creds.Password,
	}

	sendTrapLog(event)

	log.Printf("Login attempt: %s | %s | Time: %s\n", creds.Email, creds.Password, time.Now())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": "ey.fake.jwt.token",
	})
}

func main() {
	http.HandleFunc("/login", loginHandler)
	log.Println("auth-svc listening on :8060")
	log.Fatal(http.ListenAndServe(":8060", nil))
}