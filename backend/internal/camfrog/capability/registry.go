package capability

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Registry manages plugin capabilities and their verification status.
type Registry struct {
	mu           sync.RWMutex
	capabilities map[int]*PluginCapability
}

// NewRegistry creates a new capability registry.
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[int]*PluginCapability),
	}
}

// Register adds or updates a plugin capability.
func (r *Registry) Register(cap *PluginCapability) error {
	if cap == nil {
		return fmt.Errorf("capability cannot be nil")
	}
	if cap.PluginID <= 0 {
		return fmt.Errorf("invalid plugin ID: %d", cap.PluginID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.capabilities[cap.PluginID] = cap
	return nil
}

// Get retrieves a plugin capability by ID.
func (r *Registry) Get(pluginID int) (*PluginCapability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cap, ok := r.capabilities[pluginID]
	return cap, ok
}

// GetAll returns all registered capabilities.
func (r *Registry) GetAll() []*PluginCapability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	caps := make([]*PluginCapability, 0, len(r.capabilities))
	for _, cap := range r.capabilities {
		caps = append(caps, cap)
	}
	return caps
}

// GetByStatus returns capabilities with a specific status.
func (r *Registry) GetByStatus(status CapabilityStatus) []*PluginCapability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var caps []*PluginCapability
	for _, cap := range r.capabilities {
		if cap.Status == status {
			caps = append(caps, cap)
		}
	}
	return caps
}

// IsVerified returns true if the plugin capability is VERIFIED.
func (r *Registry) IsVerified(pluginID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cap, ok := r.capabilities[pluginID]
	return ok && cap.Status == CapabilityVerified
}

// IsEnabled returns true if the plugin is not DISABLED or UNSUPPORTED.
func (r *Registry) IsEnabled(pluginID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cap, ok := r.capabilities[pluginID]
	if !ok {
		return false
	}
	return cap.Status != CapabilityDisabled && cap.Status != CapabilityUnsupported
}

// UpdateStatus updates the capability status.
func (r *Registry) UpdateStatus(pluginID int, status CapabilityStatus, verifiedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cap, ok := r.capabilities[pluginID]
	if !ok {
		return fmt.Errorf("plugin %d not found", pluginID)
	}

	cap.Status = status
	cap.LastVerified = time.Now()
	cap.VerifiedBy = verifiedBy
	return nil
}

// LoadFromJSON loads capabilities from JSON data.
func (r *Registry) LoadFromJSON(data []byte) error {
	var caps []*PluginCapability
	if err := json.Unmarshal(data, &caps); err != nil {
		return fmt.Errorf("unmarshal capabilities: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cap := range caps {
		r.capabilities[cap.PluginID] = cap
	}
	return nil
}

// ToJSON serializes capabilities to JSON.
func (r *Registry) ToJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	caps := make([]*PluginCapability, 0, len(r.capabilities))
	for _, cap := range r.capabilities {
		caps = append(caps, cap)
	}
	return json.MarshalIndent(caps, "", "  ")
}

// DefaultRegistry returns a registry pre-populated with observed plugins.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	// Pre-populate with observed plugins from evidence
	observed := map[int]string{
		1:  "8 Ball",
		2:  "IM Control",
		3:  "Messenger",
		4:  "Greeter",
		5:  "Math Trivia",
		6:  "Bad Words",
		7:  "Trivia",
		8:  "Word Trivia",
		9:  "Roulette",
		10: "Unknown",
		11: "Ghost",
		12: "Who Banned",
		13: "Talkback",
		14: "Dice",
		15: "DJ Manager",
		16: "Auto Moderator",
	}

	for id, name := range observed {
		status := CapabilityUnverified
		if id == 9 || id == 10 {
			status = CapabilityDisabled // Roulette and Unknown are disabled
		}

		r.Register(&PluginCapability{
			PluginID:       id,
			Name:           name,
			Version:        "1.0.0",
			Status:         status,
			ProductVersion: "Legacy/Cloud Bot (unverified)",
			Capabilities:   []string{},
			ConfigSchema:   make(map[string]ConfigField),
			LastVerified:   time.Time{},
			VerifiedBy:     "observed capture",
		})
	}

	return r
}