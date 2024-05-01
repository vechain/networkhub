package api

import (
	"encoding/json"
	"fmt"
	"github.com/vechain/networkhub/hub"
	"github.com/vechain/networkhub/preset"
	"net/http"
	"strings"

	"github.com/vechain/networkhub/network"
)

type Server struct {
	entrypoint *hub.NetworkHub
	presets    *preset.Networks
}

func New(envMgr *hub.NetworkHub, presets *preset.Networks) *Server {
	return &Server{
		entrypoint: envMgr,
		presets:    presets,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/preset", s.presetHandler)
	http.HandleFunc("/config", s.configHandler)
	http.HandleFunc("/start", s.startHandler)
	http.HandleFunc("/stop", s.stopHandler)

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	return nil
}

func (s *Server) presetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the part of the path after "/preset/"
	networkPresetID := strings.TrimPrefix(r.URL.Path, "/preset/")
	if networkPresetID == "" {
		http.Error(w, "Network preset ID must be specified", http.StatusBadRequest)
		return
	}

	networkCfg, err := s.presets.Load(networkPresetID)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to load network preset - %s", err), http.StatusBadRequest)
		return
	}

	networkID, err := s.entrypoint.LoadNetworkConfig(networkCfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to load network config - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("{\"networkId\": \"%s\"}", networkID))

}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var networkCfg network.Network
	if err := json.NewDecoder(r.Body).Decode(&networkCfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	networkID, err := s.entrypoint.LoadNetworkConfig(&networkCfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to load network config - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("{\"networkId\": \"%s\"}", networkID))

}

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	// GET /start/NETWORKID
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the part of the path after "/start/"
	networkID := strings.TrimPrefix(r.URL.Path, "/start/")
	if networkID == "" {
		http.Error(w, "Network ID must be specified", http.StatusBadRequest)
		return
	}

	err := s.entrypoint.StartNetwork(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to start network - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Network Started\n")
}

func (s *Server) stopHandler(w http.ResponseWriter, r *http.Request) {
	// GET /stop/NETWORKID
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the part of the path after "/start/"
	networkID := strings.TrimPrefix(r.URL.Path, "/stop/")
	if networkID == "" {
		http.Error(w, "Network ID must be specified", http.StatusBadRequest)
		return
	}

	err := s.entrypoint.StopNetwork(networkID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to stop network - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Network Stopped\n")
}
