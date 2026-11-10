package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDGeneration(t *testing.T) {
	t.Parallel()

	// Test all ID types generate valid UUIDs
	orgID := NewOrganizationID()
	userID := NewUserID()
	roleID := NewRoleID()
	permID := NewPermissionID()
	sessionID := NewSessionID()
	roomID := NewRoomID()
	botID := NewBotID()
	pluginID := NewPluginID()
	configID := NewPluginConfigID()
	auditID := NewAuditLogID()
	apiKeyID := NewAPIKeyID()

	ids := []interface{}{orgID, userID, roleID, permID, sessionID, roomID, botID, pluginID, configID, auditID, apiKeyID}

	for _, id := range ids {
		uuidStr := id.(interface{ String() string }).String()
		_, err := uuid.Parse(uuidStr)
		require.NoError(t, err, "ID should be valid UUID: %v", id)
	}
}

func TestIDParsing(t *testing.T) {
	t.Parallel()

	// Test parsing valid UUIDs
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	orgID, err := ParseOrganizationID(validUUID)
	require.NoError(t, err)
	assert.Equal(t, validUUID, orgID.String())

	userID, err := ParseUserID(validUUID)
	require.NoError(t, err)
	assert.Equal(t, validUUID, userID.String())

	// Test parsing invalid UUID
	_, err = ParseOrganizationID("not-a-uuid")
	require.Error(t, err)
}

func TestEntityCreation(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Microsecond)

	// Test Organization
	org := Organization{
		ID:          NewOrganizationID(),
		Name:        "Test Org",
		Slug:        "test-org",
		Description: "Test organization",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, "test-org", org.Slug)

	// Test User
	user := User{
		ID:           NewUserID(),
		Organization: org.ID,
		Email:        "test@example.com",
		DisplayName:  "Test User",
		PasswordHash: "hashed-password",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	assert.Equal(t, "test@example.com", user.Email)
	assert.True(t, user.IsActive)

	// Test Room
	room := Room{
		ID:             NewRoomID(),
		OrganizationID: org.ID,
		Name:           "Test Room",
		AccessKeyHash:  "hashed-access-key",
		Language:       "en_US.UTF-8",
		IsPaused:       false,
		Status:         "disconnected",
		Config: RoomConfig{
			AutoJoin: true,
			MaxUsers: 50,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	assert.Equal(t, "Test Room", room.Name)
	assert.Equal(t, "en_US.UTF-8", room.Language)
	assert.True(t, room.Config.AutoJoin)

	// Test Bot
	bot := Bot{
		ID:             NewBotID(),
		OrganizationID: org.ID,
		Name:           "Test Bot",
		Username:       "testbot",
		PasswordHash:   "hashed-bot-password",
		Status:         "stopped",
		Config: BotConfig{
			AutoReconnect: true,
			LogLevel:      "info",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	assert.Equal(t, "testbot", bot.Username)
	assert.Equal(t, "stopped", bot.Status)

	// Test Plugin
	plugin := Plugin{
		ID:            NewPluginID(),
		PluginNumber:  1,
		Name:          "8 Ball",
		Version:       "1.0.0",
		Compatibility: "UNVERIFIED",
		IsEnabled:     true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	assert.Equal(t, 1, plugin.PluginNumber)
	assert.Equal(t, "UNVERIFIED", plugin.Compatibility)

	// Test PluginConfiguration
	pluginConfig := PluginConfiguration{
		ID:        NewPluginConfigID(),
		PluginID:  plugin.ID,
		RoomID:    &room.ID,
		Settings:  `{"messages":["Hello"]}`,
		IsEnabled: true,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	assert.Equal(t, plugin.ID, pluginConfig.PluginID)
	assert.Equal(t, room.ID, *pluginConfig.RoomID)

	// Test AuditLog
	auditLog := AuditLog{
		ID:             NewAuditLogID(),
		OrganizationID: org.ID,
		ActorUserID:    &user.ID,
		ActorType:      "user",
		Action:         "room.create",
		ResourceType:   "room",
		ResourceID:     room.ID.String(),
		NewValue:       `{"name":"Test Room"}`,
		CorrelationID:  "corr-123",
		CreatedAt:      now,
	}
	assert.Equal(t, "room.create", auditLog.Action)
	assert.Equal(t, "corr-123", auditLog.CorrelationID)

	// Test APIKey
	apiKey := APIKey{
		ID:             NewAPIKeyID(),
		OrganizationID: org.ID,
		Name:           "Test API Key",
		KeyHash:        "hashed-key",
		Prefix:         "cfc_abcd12",
		Permissions:    []string{"bots:read", "rooms:read"},
		CreatedAt:      now,
	}
	assert.Equal(t, "Test API Key", apiKey.Name)
	assert.Len(t, apiKey.Permissions, 2)

	// Test BotConnection
	conn := BotConnection{
		ID:        uuid.New(),
		BotID:     bot.ID,
		RoomID:    &room.ID,
		EventType: "connected",
		Message:   "Successfully connected to room",
		Metadata:  `{"ip":"192.168.1.1"}`,
		CreatedAt: now,
	}
	assert.Equal(t, "connected", conn.EventType)
	assert.Equal(t, bot.ID, conn.BotID)
}

func TestRoomConfigJSON(t *testing.T) {
	t.Parallel()

	config := RoomConfig{
		AutoJoin:   true,
		MaxUsers:   100,
		WelcomeMsg: "Welcome to the room!",
	}

	// Just verify it can be created and accessed
	assert.True(t, config.AutoJoin)
	assert.Equal(t, 100, config.MaxUsers)
	assert.Equal(t, "Welcome to the room!", config.WelcomeMsg)
}

func TestBotConfigJSON(t *testing.T) {
	t.Parallel()

	config := BotConfig{
		AutoReconnect: true,
		LogLevel:      "debug",
		ProxyURL:      "http://proxy:8080",
	}

	assert.True(t, config.AutoReconnect)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "http://proxy:8080", config.ProxyURL)
}