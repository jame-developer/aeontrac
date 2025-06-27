package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jame-developer/aeontrac/internal/appcore"
	"github.com/jame-developer/aeontrac/pkg/commands"
	"github.com/jame-developer/aeontrac/pkg/reporting"
)

// ReportHandler handles GET /report and returns today's report as JSON.
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	config, data, _, err := appcore.LoadApp()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load app: %v", err), http.StatusInternalServerError)
		return
	}

	report := reporting.GetTodayReport(config.WorkingHours, data)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode report: %v", err), http.StatusInternalServerError)
		return
	}
}

type TimeRequest struct {
	Time *string `json:"time"`
}

// StartHandler handles POST /start to start time tracking.
func StartHandler(w http.ResponseWriter, r *http.Request) {
	var req TimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, data, dataFolder, err := appcore.LoadApp()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load app: %v", err), http.StatusInternalServerError)
		return
	}

	args := []string{}
	if req.Time != nil && *req.Time != "" {
		args = append(args, *req.Time)
	}

	commands.StartCommand(args, data)
	if err := appcore.SaveApp(config, data, dataFolder); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save app: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Time tracking started successfully.")
}

// StopHandler handles POST /stop to stop time tracking.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	var req TimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, data, dataFolder, err := appcore.LoadApp()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load app: %v", err), http.StatusInternalServerError)
		return
	}

	args := []string{}
	if req.Time != nil && *req.Time != "" {
		args = append(args, *req.Time)
	}

	commands.StopCommand(args, config.WorkingHours, data)
	if err := appcore.SaveApp(config, data, dataFolder); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save app: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Time tracking stopped successfully.")
}


func main() {
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/stop", StopHandler)
	http.HandleFunc("/report", ReportHandler)

	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}