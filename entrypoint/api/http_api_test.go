package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/environments/noop"
)

func TestStartStopHandler(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantBody   string
		method     string
		payload    string
	}{
		{
			name:       "Start Network",
			target:     "/start/noop",
			payload:    "{}",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
			wantBody:   "Network Started\n",
		},
		{
			name:       "Stop service",
			target:     "/stop/noop",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "Network Stopped\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, bytes.NewReader([]byte(tc.payload)))
			recorder := httptest.NewRecorder()
			var handler http.HandlerFunc

			envManager := environments.NewEnvManager()
			envManager.RegisterEnv("noop", noop.NewNoopEnv())
			api := New(envManager)

			switch {
			case strings.Contains(tc.target, "/start/"):
				handler = api.startHandler
			case strings.Contains(tc.target, "/stop"):
				handler = api.stopHandler
			}

			handler.ServeHTTP(recorder, request)

			if status := recorder.Code; status != tc.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.wantStatus)
			}

			if body := recorder.Body.String(); !strings.Contains(body, tc.wantBody) {
				t.Errorf("handler returned unexpected body: got %v want %v", body, tc.wantBody)
			}
		})
	}
}
