package handler

import (
	"config-service/internal/service"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// //go:embed doc.yaml doc.json
var swaggerDocs embed.FS

type ConfigHandler struct {
	service service.ConfigService
}

func NewConfigHandler(service service.ConfigService) *ConfigHandler {
	return &ConfigHandler{service: service}
}

func (h *ConfigHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/doc.json", h.swaggerJSON)
	mux.HandleFunc("/doc.yaml", h.swaggerYAML)
	mux.HandleFunc("/configs/", h.handleConfigs)
}

func (h *ConfigHandler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ConfigHandler) handleConfigs(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/configs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "environment is required", http.StatusBadRequest)
		return
	}

	environment := parts[0]

	switch {
	case len(parts) == 1:
		if r.Method == http.MethodGet {
			h.getAllConfigs(w, r, environment)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case len(parts) == 2:
		key := parts[1]
		switch r.Method {
		case http.MethodGet:
			h.getConfig(w, r, environment, key)
		case http.MethodPut:
			h.updateConfig(w, r, environment, key)
		case http.MethodPost:
			h.createConfig(w, r, environment, key)
		case http.MethodDelete:
			h.deleteConfig(w, r, environment, key)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "invalid path", http.StatusBadRequest)
	}
}

func (h *ConfigHandler) createConfig(w http.ResponseWriter, r *http.Request, environment, key string) {
	var req struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateConfig(environment, key, req.Value); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ConfigHandler) getConfig(w http.ResponseWriter, r *http.Request, environment, key string) {
	config, err := h.service.GetConfig(environment, key)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func (h *ConfigHandler) getAllConfigs(w http.ResponseWriter, r *http.Request, environment string) {
	configs, err := h.service.GetAllConfigs(environment)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(configs)
}

func (h *ConfigHandler) updateConfig(w http.ResponseWriter, r *http.Request, environment, key string) {
	var req struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateConfig(environment, key, req.Value); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConfigHandler) deleteConfig(w http.ResponseWriter, r *http.Request, environment, key string) {
	if err := h.service.DeleteConfig(environment, key); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConfigHandler) handleError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, service.ErrConfigNotFound):
		statusCode = http.StatusNotFound
		message = "config not found"
	case errors.Is(err, service.ErrConfigExists):
		statusCode = http.StatusConflict
		message = "config already exists"
	default:
		statusCode = http.StatusInternalServerError
		message = "internal server error"
	}

	http.Error(w, message, statusCode)
}

func (h *ConfigHandler) swaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data, err := swaggerDocs.ReadFile("doc.json")
	if err != nil {
		http.Error(w, "doc.json not found", http.StatusNotFound)
		return
	}
	w.Write(data)
	if err != nil {
		fmt.Println("error write: ", err)
	}
}

func (h *ConfigHandler) swaggerYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data, err := swaggerDocs.ReadFile("swagger.yaml")
	if err != nil {
		http.Error(w, "swagger.yaml not found", http.StatusNotFound)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		fmt.Println("error write: ", err)
	}
}
