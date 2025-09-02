package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jame-developer/aeontrac/internal/api/handlers"
	"github.com/jame-developer/aeontrac/internal/api/router"
	"github.com/jame-developer/aeontrac/internal/appcore"
	"github.com/jame-developer/aeontrac/pkg/reporting"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	os.Setenv("AEON_ROOT", "/home/jan/dev/aeontrac")
	appcore.SetTestDir(dir)

}

func teardownTest(t *testing.T) {
	appcore.SetTestDir("")
}

func TestStartHandler(t *testing.T) {
	t.Run("with time parameter", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)
		config, data, dataFolder, err := appcore.LoadApp()

		if err != nil {
			t.Fatalf("Failed to load app: %v", err)
		}
		r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))
		timeStr := "2023-01-01T10:00:00"
		reqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: timeStr})
		req, err := http.NewRequest("POST", "/api/tracking/start", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := "Time tracking started successfully."
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("without time parameter", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)
		config, data, dataFolder, err := appcore.LoadApp()

		if err != nil {
			t.Fatalf("Failed to load app: %v", err)
		}
		r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))
		reqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: ""})
		req, err := http.NewRequest("POST", "/api/tracking/start", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := "Time tracking started successfully."
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)

		r := router.SetupRouter(zap.NewNop(), nil, nil, "", os.Getenv("AEON_ROOT"))
		req, err := http.NewRequest("POST", "/api/tracking/start", bytes.NewBuffer([]byte("invalid json")))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})
}

func TestStopHandler(t *testing.T) {
	t.Run("with time parameter", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)
		config, data, dataFolder, err := appcore.LoadApp()

		if err != nil {
			t.Fatalf("Failed to load app: %v", err)
		}
		r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))

		// Start a timer to be stopped
		startTimeStr := "2023-01-01T10:00:00"
		startReqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: startTimeStr})
		startReq, err := http.NewRequest("POST", "/api/tracking/start", bytes.NewBuffer(startReqBody))
		if err != nil {
			t.Fatal(err)
		}
		startRR := httptest.NewRecorder()
		r.ServeHTTP(startRR, startReq)
		if startRR.Code != http.StatusOK {
			t.Fatalf("pre-test start handler failed with status: %v", startRR.Code)
		}

		timeStr := "2023-01-01T18:00:00"
		reqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: timeStr})
		req, err := http.NewRequest("POST", "/api/tracking/stop", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := "Time tracking stopped successfully."
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("without time parameter", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)
		config, data, dataFolder, err := appcore.LoadApp()

		if err != nil {
			t.Fatalf("Failed to load app: %v", err)
		}
		r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))

		// Start a timer to be stopped
		startTimeStr := "2023-01-01T10:00:00"
		startReqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: startTimeStr})
		startReq, err := http.NewRequest("POST", "/api/tracking/start", bytes.NewBuffer(startReqBody))
		if err != nil {
			t.Fatal(err)
		}
		startRR := httptest.NewRecorder()
		r.ServeHTTP(startRR, startReq)
		if startRR.Code != http.StatusOK {
			t.Fatalf("pre-test start handler failed with status: %v", startRR.Code)
		}

		reqBody, _ := json.Marshal(handlers.StartTrackingRequest{Comment: ""})
		req, err := http.NewRequest("POST", "/api/tracking/stop", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := "Time tracking stopped successfully."
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		setupTest(t)
		defer teardownTest(t)
		config, data, dataFolder, err := appcore.LoadApp()

		if err != nil {
			t.Fatalf("Failed to load app: %v", err)
		}
		r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))
		req, err := http.NewRequest("POST", "/api/tracking/stop", bytes.NewBuffer([]byte("invalid json")))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})
}

func TestReportHandler(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)
	config, data, dataFolder, err := appcore.LoadApp()

	if err != nil {
		t.Fatalf("Failed to load app: %v", err)
	}
	r := router.SetupRouter(zap.NewNop(), config, data, dataFolder, os.Getenv("AEON_ROOT"))
	req, err := http.NewRequest("GET", "/api/report/today", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}

	var report reporting.TodayReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Errorf("handler returned invalid json: %v", err)
	}
}
