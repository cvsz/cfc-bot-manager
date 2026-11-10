package capability

import (
	"time"
)

// CapabilityStatus represents the verification status of a plugin capability.
type CapabilityStatus string

const (
	CapabilityVerified         CapabilityStatus = "VERIFIED"
	CapabilityPartiallyVerified CapabilityStatus = "PARTIALLY_VERIFIED"
	CapabilityUnverified       CapabilityStatus = "UNVERIFIED"
	CapabilityUnsupported      CapabilityStatus = "UNSUPPORTED"
	CapabilityDisabled         CapabilityStatus = "DISABLED"
)

// PluginCapability describes a plugin's capability and verification status.
type PluginCapability struct {
	PluginID       int
	Name           string
	Version        string
	Status         CapabilityStatus
	ProductVersion string // e.g., "Cloud Bot 1.2.3" or "Legacy Standalone 2.0"
	Capabilities   []string
	ConfigSchema   map[string]ConfigField // Field name -> schema
	LastVerified   time.Time
	VerifiedBy     string // Test identity or evidence reference
}

// ConfigField describes a plugin configuration field.
type ConfigField struct {
	Name        string
	Type        string // "string", "int", "bool", "select", "textarea", "json"
	Required    bool
	Default     any
	Options     []string // For select fields
	Validation  string   // Regex or validation rule
	Description string
}