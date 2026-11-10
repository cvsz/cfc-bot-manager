package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog"
	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
)

type camfrogHandler struct {
	repo       repository.RepositoryProvider
	wsHub      *wsHub
	adapters   map[string]*camfrogAdapterState
	adaptersMu sync.RWMutex
}

type camfrogAdapterState struct {
	adapter  camfrog.CamfrogAdapter
	botID    domain.BotID
	roomID   domain.RoomID
	cancel   context.CancelFunc
	running  bool
	mu       sync.Mutex
}

func NewCamfrogHandler(repo repository.RepositoryProvider, wsHub *wsHub) *camfrogHandler {
	return &camfrogHandler{
		repo:     repo,
		wsHub:    wsHub,
		adapters: make(map[string]*camfrogAdapterState),
	}
}

func (h *camfrogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		writeAPIError(w, http.StatusNotFound, "not found")
		return
	}

	resource := parts[0]
	resourceID := parts[1]

	switch {
	case resource == "bots" && len(parts) >= 3:
		h.handleBotRoutes(w, r, resourceID, parts[2:])
	case resource == "rooms" && len(parts) >= 3:
		h.handleRoomRoutes(w, r, resourceID, parts[2:])
	case resource == "plugins" && len(parts) >= 3:
		h.handlePluginRoutes(w, r, resourceID, parts[2:])
	default:
		writeAPIError(w, http.StatusNotFound, "not found")
	}
}

func (h *camfrogHandler) handleBotRoutes(w http.ResponseWriter, r *http.Request, botIDStr string, parts []string) {
	if len(parts) == 0 {
		writeAPIError(w, http.StatusBadRequest, "missing action")
		return
	}

	botID, err := domain.ParseBotID(botIDStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid bot ID")
		return
	}

	action := parts[0]

	switch {
	case action == "status" && r.Method == http.MethodGet:
		h.getBotStatus(w, r, botID)
	case action == "language" && r.Method == http.MethodPut:
		h.setBotLanguage(w, r, botID)
	case action == "plugins" && len(parts) >= 2 && r.Method == http.MethodPut:
		h.toggleBotPlugin(w, r, botID, parts[1])
	case action == "start" && r.Method == http.MethodPost:
		h.startBot(w, r, botID)
	case action == "stop" && r.Method == http.MethodPost:
		h.stopBot(w, r, botID)
	case action == "restart" && r.Method == http.MethodPost:
		h.restartBot(w, r, botID)
	default:
		writeAPIError(w, http.StatusNotFound, "not found")
	}
}

func (h *camfrogHandler) handleRoomRoutes(w http.ResponseWriter, r *http.Request, roomIDStr string, parts []string) {
	if len(parts) == 0 {
		writeAPIError(w, http.StatusBadRequest, "missing action")
		return
	}

	roomID, err := domain.ParseRoomID(roomIDStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid room ID")
		return
	}

	action := parts[0]

	switch {
	case action == "pause" && r.Method == http.MethodPost:
		h.pauseRoom(w, r, roomID)
	case action == "language" && r.Method == http.MethodPut:
		h.setRoomLanguage(w, r, roomID)
	default:
		writeAPIError(w, http.StatusNotFound, "not found")
	}
}

func (h *camfrogHandler) handlePluginRoutes(w http.ResponseWriter, r *http.Request, pluginIDStr string, parts []string) {
	if len(parts) == 0 {
		writeAPIError(w, http.StatusBadRequest, "missing action")
		return
	}

	pluginID, err := strconv.Atoi(pluginIDStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid plugin ID")
		return
	}

	action := parts[0]

	switch {
	case action == "settings" && len(parts) >= 2:
		roomIDStr := parts[1]
		roomID, err := domain.ParseRoomID(roomIDStr)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid room ID")
			return
		}

		if r.Method == http.MethodGet {
			h.getPluginSettings(w, r, pluginID, roomID)
		} else if r.Method == http.MethodPut {
			h.savePluginSettings(w, r, pluginID, roomID)
		} else {
			writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	default:
		writeAPIError(w, http.StatusNotFound, "not found")
	}
}

func (h *camfrogHandler) getBotStatus(w http.ResponseWriter, r *http.Request, botID domain.BotID) {
	adapter := h.getOrCreateAdapter(botID)
	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "bot not found or not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	status, err := adapter.adapter.Status(ctx)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get status: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, status)
}

func (h *camfrogHandler) setBotLanguage(w http.ResponseWriter, r *http.Request, botID domain.BotID) {
	var req struct {
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Locale == "" {
		writeAPIError(w, http.StatusBadRequest, "locale is required")
		return
	}

	adapter := h.getOrCreateAdapter(botID)
	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "bot not found or not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := adapter.adapter.Language(ctx, req.Locale); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to set language: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *camfrogHandler) toggleBotPlugin(w http.ResponseWriter, r *http.Request, botID domain.BotID, pluginIDStr string) {
	pluginID, err := strconv.Atoi(pluginIDStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid plugin ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	adapter := h.getOrCreateAdapter(botID)
	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "bot not found or not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := adapter.adapter.SwitchPlugin(ctx, pluginID, req.Enabled); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to toggle plugin: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *camfrogHandler) startBot(w http.ResponseWriter, r *http.Request, botID domain.BotID) {
	adapter := h.getOrCreateAdapter(botID)
	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "bot not found or not configured")
		return
	}

	adapter.mu.Lock()
	if adapter.running {
		adapter.mu.Unlock()
		writeAPIError(w, http.StatusConflict, "bot already running")
		return
	}
	adapter.running = true
	adapter.mu.Unlock()

	ctx, cancel := context.WithCancel(r.Context())
	adapter.cancel = cancel

	go h.runBot(ctx, adapter)

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "starting"})
}

func (h *camfrogHandler) stopBot(w http.ResponseWriter, r *http.Request, botID domain.BotID) {
	h.adaptersMu.Lock()
	adapter, ok := h.adapters[botID.String()]
	h.adaptersMu.Unlock()

	if !ok {
		writeAPIError(w, http.StatusNotFound, "bot not running")
		return
	}

	adapter.mu.Lock()
	if adapter.cancel != nil {
		adapter.cancel()
	}
	adapter.running = false
	adapter.mu.Unlock()

	_ = adapter.adapter.Close()

	h.adaptersMu.Lock()
	delete(h.adapters, botID.String())
	h.adaptersMu.Unlock()

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *camfrogHandler) restartBot(w http.ResponseWriter, r *http.Request, botID domain.BotID) {
	h.stopBot(w, r, botID)
	time.Sleep(500 * time.Millisecond)
	h.startBot(w, r, botID)
}

func (h *camfrogHandler) runBot(ctx context.Context, adapter *camfrogAdapterState) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			adapter.mu.Lock()
			if !adapter.running {
				adapter.mu.Unlock()
				return
			}
			adapter.mu.Unlock()

			status, err := adapter.adapter.Status(ctx)
			if err != nil {
				// Log error but continue
				continue
			}

			// Broadcast status change via WebSocket
			if h.wsHub != nil {
				h.wsHub.BroadcastBotStatus(adapter.botID, status)
			}
		}
	}
}

func (h *camfrogHandler) getOrCreateAdapter(botID domain.BotID) *camfrogAdapterState {
	h.adaptersMu.RLock()
	adapter, ok := h.adapters[botID.String()]
	h.adaptersMu.RUnlock()

	if ok {
		return adapter
	}

	// Fetch bot from repository
	bot, err := h.repo.Bots().GetByID(context.Background(), botID)
	if err != nil {
		return nil
	}

	// Fetch room if assigned
	var room *domain.Room
	if bot.RoomID != nil {
		room, _ = h.repo.Rooms().GetByID(context.Background(), *bot.RoomID)
	}

	// Create adapter config
	cfg := camfrog.Config{
		BaseURL:    "http://localhost:8080", // TODO: Get from bot config
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		UserAgent:  "CFC Bot Manager/1.0",
	}

	adapterObj, err := camfrog.NewAdapter(cfg)
	if err != nil {
		return nil
	}

	adapterState := &camfrogAdapterState{
		adapter: adapterObj,
		botID:   botID,
		running: false,
	}

	if room != nil {
		adapterState.roomID = room.ID
	}

	h.adaptersMu.Lock()
	h.adapters[botID.String()] = adapterState
	h.adaptersMu.Unlock()

	return adapterState
}

func (h *camfrogHandler) pauseRoom(w http.ResponseWriter, r *http.Request, roomID domain.RoomID) {
	var req struct {
		Pause bool `json:"pause"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Find adapter for this room
	h.adaptersMu.RLock()
	var adapter *camfrogAdapterState
	for _, a := range h.adapters {
		if a.roomID == roomID {
			adapter = a
			break
		}
	}
	h.adaptersMu.RUnlock()

	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "no active bot for this room")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := adapter.adapter.RoomPause(ctx, roomID.String(), req.Pause); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to pause room: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *camfrogHandler) setRoomLanguage(w http.ResponseWriter, r *http.Request, roomID domain.RoomID) {
	var req struct {
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Locale == "" {
		writeAPIError(w, http.StatusBadRequest, "locale is required")
		return
	}

	h.adaptersMu.RLock()
	var adapter *camfrogAdapterState
	for _, a := range h.adapters {
		if a.roomID == roomID {
			adapter = a
			break
		}
	}
	h.adaptersMu.RUnlock()

	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "no active bot for this room")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := adapter.adapter.RoomSetLanguage(ctx, roomID.String(), req.Locale); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to set room language: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *camfrogHandler) getPluginSettings(w http.ResponseWriter, r *http.Request, pluginID int, roomID domain.RoomID) {
	h.adaptersMu.RLock()
	var adapter *camfrogAdapterState
	for _, a := range h.adapters {
		if a.roomID == roomID {
			adapter = a
			break
		}
	}
	h.adaptersMu.RUnlock()

	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "no active bot for this room")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	settings, err := adapter.adapter.GetPluginSettings(ctx, pluginID, roomID.String())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get plugin settings: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, settings)
}

func (h *camfrogHandler) savePluginSettings(w http.ResponseWriter, r *http.Request, pluginID int, roomID domain.RoomID) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.adaptersMu.RLock()
	var adapter *camfrogAdapterState
	for _, a := range h.adapters {
		if a.roomID == roomID {
			adapter = a
			break
		}
	}
	h.adaptersMu.RUnlock()

	if adapter == nil {
		writeAPIError(w, http.StatusNotFound, "no active bot for this room")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := adapter.adapter.SavePluginSettings(ctx, pluginID, roomID.String(), settings); err != nil {
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save plugin settings: %v", err))
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}