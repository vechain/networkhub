package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
)

type Server struct {
	envMgr *environments.EnvManager
}

func New(envMgr *environments.EnvManager) *Server {
	return &Server{
		envMgr: envMgr,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/start", s.startHandler)
	http.HandleFunc("/stop", s.stopHandler)

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	return nil
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the part of the path after "/start/"
	envName := strings.TrimPrefix(r.URL.Path, "/start/")
	if envName == "" {
		http.Error(w, "Environment type must be specified", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var networkCfg network.Network
	if err := json.NewDecoder(r.Body).Decode(&networkCfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	env := s.envMgr.Env(envName)
	if env == nil {
		http.Error(w, fmt.Sprintf("Environment type %s does not exist", envName), http.StatusBadRequest)
		return
	}

	err := env.StartNetwork(&networkCfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to start environment - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Network Started\n")
}

func (s *Server) stopHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the part of the path after "/stop/"
	envName := strings.TrimPrefix(r.URL.Path, "/stop/")
	if envName == "" {
		http.Error(w, "Environment type must be specified", http.StatusBadRequest)
		return
	}

	env := s.envMgr.Env(envName)
	if env == nil {
		http.Error(w, fmt.Sprintf("Environment type %s does not exist", envName), http.StatusBadRequest)
		return
	}

	err := env.StopNetwork()
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to stop environment - %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Network Stopped\n")
}
