package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type TrapEvent struct {
	Service   string `json:"service"`
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	Details   string `json:"details"`
	IP        string `json:"ip"`
}

var services = []string{"ai-chat-svc", "auth-svc", "admin-svc", "test-svc"}
var events = []string{
	"Prompt injection attempt",
	"Login brute force",
	"System override attempt",
	"Fake token replay",
	"Kernel-level enumeration",
	"Unauthorized config fetch",
}

var ips = []string{
	"8.8.8.8",       // USA
	"1.1.1.1",       // Australia
	"91.198.174.192",// Netherlands
	"202.108.22.5",  // China
	"77.88.55.77",   // Russia
	"200.89.75.197", // Argentina
	"41.77.224.2",   // Nigeria
	"213.186.33.5",  // France
	"185.60.216.35", // Germany
	"62.75.139.38",  // UAE
}

func main() {
	url := "http://localhost:8091/log"
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 20; i++ {
		ip := ips[rand.Intn(len(ips))]
		service := services[rand.Intn(len(services))]
		event := events[rand.Intn(len(events))]

		trap := TrapEvent{
			Service:   service,
			Event:     event,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Details:   "Simulated attack for testing",
			IP:        ip,
		}

		body, _ := json.Marshal(trap)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", ip)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Failed to send:", err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("Sent simulated trap from %s (%s)\n", ip, event)
		time.Sleep(1 * time.Second)
	}
}