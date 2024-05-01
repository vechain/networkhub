package api

import (
	"bytes"
	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/preset"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			name:       "Config Network",
			target:     "/config",
			payload:    "{\"environment\":\"noop\"}",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
			wantBody:   "{\"networkId\": \"noop\"}",
		},
		{
			name:       "Start Network",
			target:     "/start/noop",
			payload:    "{}",
			method:     http.MethodGet,
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
		{
			name:       "Start non-existent network",
			target:     "/start/no-exist",
			payload:    "{}",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Unable to start network - network no-exist is not configured\n",
		},
		{
			name:       "Stop non-existent network",
			target:     "/stop/no-exist",
			payload:    "{}",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Unable to stop network - network no-exist is not configured\n",
		},
		{
			name:       "Load existing preset network",
			target:     "/preset/noop-network",
			payload:    "{}",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "{\"networkId\": \"noop\"}",
		},
		{
			name:       "Load non-preset network",
			target:     "/preset/noop-network-no-exist",
			payload:    "{}",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
			wantBody:   "unable to load network preset - unable to find preset with id noop-network-no-exist",
		},
	}

	envManager := hub.NewNetworkHub()
	envManager.RegisterEnvironment("noop", noop.NewNoopEnv)

	presets := preset.NewPresetNetworks()
	presets.Register("noop-network", &network.Network{Environment: "noop"})

	api := New(envManager, presets)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, bytes.NewReader([]byte(tc.payload)))
			recorder := httptest.NewRecorder()
			var handler http.HandlerFunc

			switch {
			case strings.Contains(tc.target, "/preset"):
				handler = api.presetHandler
			case strings.Contains(tc.target, "/config"):
				handler = api.configHandler
			case strings.Contains(tc.target, "/start"):
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
