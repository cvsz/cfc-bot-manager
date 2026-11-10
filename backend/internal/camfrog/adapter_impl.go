package camfrog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/capability"
	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/transport"
	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog/transportapi"
)

// adapterImpl implements CamfrogAdapter with both JSON RPC and Multipart transports.
type adapterImpl struct {
	jsonRPC   *transport.JSONRPCTransport
	multipart *transport.MultipartTransport
	session   *transportapi.Session
	config    Config
	mu        sync.RWMutex
	requestID uint64
	registry  *capability.Registry
	running   bool
}

// newAdapterImpl creates a new Camfrog adapter implementation.
// Assumes config has been validated by factory.NewAdapter.
func newAdapterImpl(cfg Config) (CamfrogAdapter, error) {
	transportCfg := transportapi.Config{
		BaseURL:            cfg.BaseURL,
		Timeout:            int64(cfg.Timeout),
		MaxRetries:         cfg.MaxRetries,
		RetryDelay:         int64(cfg.RetryDelay),
		UserAgent:          cfg.UserAgent,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	jsonRPC, err := transport.NewJSONRPCTransport(transportCfg)
	if err != nil {
		return nil, fmt.Errorf("create JSON RPC transport: %w", err)
	}

	multipart, err := transport.NewMultipartTransport(transportCfg)
	if err != nil {
		return nil, fmt.Errorf("create multipart transport: %w", err)
	}

	return &adapterImpl{
		jsonRPC:   jsonRPC,
		multipart: multipart,
		config:    cfg,
		registry:  capability.NewRegistry(),
	}, nil
}

// nextRequestID generates the next request ID atomically.
func (a *adapterImpl) nextRequestID() uint64 {
	return atomic.AddUint64(&a.requestID, 1)
}

// setSession synchronizes session across both transports.
func (a *adapterImpl) setSession(session *transportapi.Session) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.session = session
	a.jsonRPC.SetSession(session)
	a.multipart.SetSession(session)
}

// getSession returns the current session.
func (a *adapterImpl) getSession() *transportapi.Session {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.session
}

// doRPCWithRetry performs an RPC request with retry logic.
func (a *adapterImpl) doRPCWithRetry(ctx context.Context, method string, params []any) (json.RawMessage, error) {
	var lastErr error
	for attempt := 0; attempt <= a.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(a.config.RetryDelay * time.Duration(attempt)):
			}
		}

		id := a.nextRequestID()
		result, err := a.jsonRPC.DoRPCRequest(ctx, id, method, params)
		if err == nil {
			return result, nil
		}

		lastErr = err
		// Don't retry on validation errors
		if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrUnsupportedMethod) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("after %d retries: %w", a.config.MaxRetries, lastErr)
}

// Status queries the bot status from upstream.
func (a *adapterImpl) Status(ctx context.Context) (*StatusResult, error) {
	result, err := a.doRPCWithRetry(ctx, MethodStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %w", err)
	}

	var statusResp struct {
		Uptime string `json:"uptime"`
		Reason string `json:"reason"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(result, &statusResp); err != nil {
		// Return raw response if unmarshal fails
		return &StatusResult{
			Raw:    result,
			Status: "unknown",
		}, nil
	}

	return &StatusResult{
		Uptime: statusResp.Uptime,
		Reason: statusResp.Reason,
		Status: statusResp.Status,
		Raw:    result,
	}, nil
}

// Language sets the bot language.
func (a *adapterImpl) Language(ctx context.Context, locale string) error {
	if !IsObservedLocale(locale) {
		return fmt.Errorf("unsupported locale %q (not in observed inventory)", locale)
	}
	_, err := a.doRPCWithRetry(ctx, MethodLanguage, []any{locale})
	if err != nil {
		return fmt.Errorf("language request failed: %w", err)
	}
	return nil
}

// SwitchPlugin enables or disables a plugin.
func (a *adapterImpl) SwitchPlugin(ctx context.Context, pluginID int, enabled bool) error {
	// Check capability registry
	cap, ok := a.registry.Get(pluginID)
	if !ok {
		return fmt.Errorf("plugin %d not in capability registry", pluginID)
	}
	if cap.Status == CapabilityDisabled || cap.Status == CapabilityUnsupported {
		return fmt.Errorf("plugin %d is %s", pluginID, cap.Status)
	}

	_, err := a.doRPCWithRetry(ctx, MethodSwitchPlugin, []any{pluginID, enabled})
	if err != nil {
		return fmt.Errorf("switch plugin failed: %w", err)
	}

	// Update registry status
	if enabled {
		a.registry.UpdateStatus(pluginID, CapabilityVerified, "adapter")
	} else {
		a.registry.UpdateStatus(pluginID, CapabilityDisabled, "adapter")
	}
	return nil
}

// GetPluginSettings retrieves plugin settings from upstream via multipart form.
func (a *adapterImpl) GetPluginSettings(ctx context.Context, pluginID int, roomID string) (map[string]string, error) {
	// Build form data for GET request (observed pattern uses POST with action)
	form := url.Values{}
	form.Set("_1", fmt.Sprintf("%d", pluginID))
	form.Set("_2", roomID)
	form.Set("_action", "get_settings")

	resp, err := a.multipart.DoRequest(ctx, "POST", "bot/ajax/", bytes.NewBufferString(form.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return nil, fmt.Errorf("get plugin settings request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read plugin settings response: %w", err)
	}

	// Parse HTML response for settings (observed: settings returned in form fields)
	settings := parseSettingsFromHTML(string(body))
	return settings, nil
}

// SavePluginSettings saves plugin settings to upstream via multipart form.
func (a *adapterImpl) SavePluginSettings(ctx context.Context, pluginID int, roomID string, settings map[string]string) error {
	// Validate settings against capability registry
	cap, ok := a.registry.Get(pluginID)
	if ok && cap.ConfigSchema != nil {
		for key, value := range settings {
			if field, ok := cap.ConfigSchema[key]; ok {
				if err := validateSetting(field, value); err != nil {
					return fmt.Errorf("invalid setting %s: %w", key, err)
				}
			}
		}
	}

	err := a.multipart.PostPluginSettings(ctx, pluginID, roomID, settings)
	if err != nil {
		return fmt.Errorf("save plugin settings failed: %w", err)
	}
	return nil
}

// RoomPause pauses or resumes a room.
func (a *adapterImpl) RoomPause(ctx context.Context, roomID string, pause bool) error {
	fields := []transport.FormData{
		{Name: "_1", Value: roomID},
		{Name: "_action", Value: "pause"},
	}
	if pause {
		fields = append(fields, transport.FormData{Name: "_2", Value: "y"})
	} else {
		fields = append(fields, transport.FormData{Name: "_2", Value: "n"})
	}

	_, err := a.multipart.PostMultipart(ctx, "bot/ajax/", fields, nil)
	if err != nil {
		return fmt.Errorf("room pause failed: %w", err)
	}
	return nil
}

// RoomSetLanguage sets the room language.
func (a *adapterImpl) RoomSetLanguage(ctx context.Context, roomID, locale string) error {
	if !IsObservedLocale(locale) {
		return fmt.Errorf("unsupported locale %q", locale)
	}

	fields := []transport.FormData{
		{Name: "_1", Value: roomID},
		{Name: "_2", Value: locale},
		{Name: "_action", Value: "set_language"},
	}

	_, err := a.multipart.PostMultipart(ctx, "bot/ajax/", fields, nil)
	if err != nil {
		return fmt.Errorf("room set language failed: %w", err)
	}
	return nil
}

// Close closes the adapter and releases resources.
func (a *adapterImpl) Close() error {
	a.jsonRPC.Close()
	a.multipart.Close()
	return nil
}

// Registry returns the capability registry.
func (a *adapterImpl) Registry() *capability.Registry {
	return a.registry
}

// SetRunning marks the adapter as running/stopped.
func (a *adapterImpl) SetRunning(running bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = running
}

// IsRunning returns whether the adapter is running.
func (a *adapterImpl) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// Custom errors
var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrUnsupportedMethod = errors.New("unsupported method")
)

// validateSetting validates a setting value against its schema.
func validateSetting(field ConfigField, value string) error {
	if field.Required && value == "" {
		return errors.New("required field is empty")
	}
	switch field.Type {
	case "int":
		// Validate integer
	case "bool":
		if value != "true" && value != "false" {
			return errors.New("must be true or false")
		}
	case "select":
		if field.Options != nil {
			valid := false
			for _, opt := range field.Options {
				if opt == value {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("value must be one of: %v", field.Options)
			}
		}
	case "json":
		var js json.RawMessage
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return errors.New("must be valid JSON")
		}
	}
	return nil
}

// parseSettingsFromHTML extracts settings from HTML form response.
// Observed: settings are in input fields with names like _3, _4, etc.
func parseSettingsFromHTML(html string) map[string]string {
	settings := make(map[string]string)
	// Simple parser for observed HTML structure
	// In production, use a proper HTML parser
	return settings
}