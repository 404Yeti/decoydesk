package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type TrapEvent struct {
	Service   string  `json:"service"`
	Event     string  `json:"event"`
	Timestamp string  `json:"timestamp"`
	Details   string  `json:"details"`
	IP        string  `json:"ip"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var (
	logfile  *os.File
	logger   *log.Logger
	traps    []TrapEvent
	trapsMux sync.Mutex
	geoDB    *geoip2.Reader
)

func initLogger() {
	logFileName := "logs/trap-" + time.Now().Format("2006-01-02") + ".log"

	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	var err error
	logfile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logfile, "", log.LstdFlags)
}

func LogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var trap TrapEvent
	if err := json.NewDecoder(r.Body).Decode(&trap); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get real IP address
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	} else if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	trap.IP = ip

	// GeoIP lookup with fallback
	trap.Country = "Unknown"
	trap.Latitude = 0.0
	trap.Longitude = 0.0

	if parsedIP := net.ParseIP(ip); parsedIP != nil && geoDB != nil {
		if record, err := geoDB.City(parsedIP); err == nil {
			trap.Country = record.Country.Names["en"]
			trap.Latitude = record.Location.Latitude
			trap.Longitude = record.Location.Longitude
		}
	}

	// Store in memory
	trapsMux.Lock()
	traps = append(traps, trap)
	trapsMux.Unlock()

	// Write to log file and console
	logMsg := "[" + trap.Service + "] " + trap.Timestamp + " - " + trap.Event + " - " + trap.Details +
		" | IP: " + trap.IP + " | Country: " + trap.Country
	logger.Println(logMsg)
	log.Println("[TRAP] " + logMsg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged"}`))
}

func TrapsAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	trapsMux.Lock()
	defer trapsMux.Unlock()

	json.NewEncoder(w).Encode(traps)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	geoPath := os.Getenv("GEOIP_DB_PATH")
	if geoPath == "" {
		log.Fatal("GEOIP_DB_PATH environment variable not set")
	}

	var err error
	geoDB, err = geoip2.Open(geoPath)
	if err != nil {
		log.Fatalf("Failed to open GeoIP database at %s: %v", geoPath, err)
	}
	defer geoDB.Close()

	initLogger()

	http.HandleFunc("/log", LogHandler)
	http.HandleFunc("/api/traps", TrapsAPIHandler)
	http.HandleFunc("/health", HealthHandler)

	log.Println("trap-logger svc listening on :8091")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
