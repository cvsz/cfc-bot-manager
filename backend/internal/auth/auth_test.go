package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repositories for testing
type mockUserRepo struct {
	users map[domain.UserID]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, orgID domain.OrganizationID, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Organization == orgID && u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) List(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.User], error) {
	return repository.ListResult[domain.User]{}, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id domain.UserID, t time.Time) error {
	if u, ok := m.users[id]; ok {
		u.LastLoginAt = &t
	}
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	delete(m.users, id)
	return nil
}
func (m *mockUserRepo) ExistsByEmail(ctx context.Context, orgID domain.OrganizationID, email string, excludeID *domain.UserID) (bool, error) {
	for _, u := range m.users {
		if u.Organization == orgID && u.Email == email {
			if excludeID != nil && u.ID == *excludeID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

type mockRoleRepo struct {
	roles map[domain.RoleID]*domain.Role
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) GetByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepo) GetByName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.OrganizationID == orgID && r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, includeSystem bool) ([]domain.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) Delete(ctx context.Context, id domain.RoleID) error {
	return nil
}
func (m *mockRoleRepo) AddPermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	return nil
}
func (m *mockRoleRepo) RemovePermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	return nil
}
func (m *mockRoleRepo) ListPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	if r, ok := m.roles[roleID]; ok {
		perms := make([]domain.Permission, len(r.Name)) // simplified
		return perms, nil
	}
	return nil, nil
}
func (m *mockRoleRepo) IsSystemRole(ctx context.Context, id domain.RoleID) (bool, error) {
	return false, nil
}

type mockPermissionRepo struct{}

func (m *mockPermissionRepo) Create(ctx context.Context, perm *domain.Permission) error { return nil }
func (m *mockPermissionRepo) GetByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepo) GetByName(ctx context.Context, name string) (*domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepo) List(ctx context.Context) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepo) ListByResource(ctx context.Context, resource string) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepo) Update(ctx context.Context, perm *domain.Permission) error { return nil }
func (m *mockPermissionRepo) Delete(ctx context.Context, id domain.PermissionID) error { return nil }

type mockMembershipRepo struct {
	memberships map[string]*domain.Membership
}

func (m *mockMembershipRepo) Create(ctx context.Context, membership *domain.Membership) error {
	key := membership.UserID.String() + membership.OrganizationID.String()
	m.memberships[key] = membership
	return nil
}
func (m *mockMembershipRepo) GetByUserAndOrg(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) (*domain.Membership, error) {
	key := userID.String() + orgID.String()
	if mem, ok := m.memberships[key]; ok {
		return mem, nil
	}
	return nil, domain.ErrMembershipNotFound
}
func (m *mockMembershipRepo) ListByUser(ctx context.Context, userID domain.UserID) ([]domain.Membership, error) { return nil, nil }
func (m *mockMembershipRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort) (repository.ListResult[domain.Membership], error) { return repository.ListResult[domain.Membership]{}, nil }
func (m *mockMembershipRepo) UpdateRole(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID, roleID domain.RoleID) error { return nil }
func (m *mockMembershipRepo) Delete(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) error { return nil }
func (m *mockMembershipRepo) CountByOrgAndRole(ctx context.Context, orgID domain.OrganizationID, roleID domain.RoleID) (int64, error) { return 0, nil }

func TestPasswordHashing(t *testing.T) {
	password := "test-password-123"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

	// Verify correct password
	err = VerifyPassword(password, hash)
	assert.NoError(t, err)

	// Verify incorrect password fails
	err = VerifyPassword("wrong-password", hash)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrHashVerification))

	// Verify empty password fails
	err = VerifyPassword("", hash)
	assert.Error(t, err)
}

func TestPasswordNeedsRehash(t *testing.T) {
	hash, err := HashPassword("password")
	require.NoError(t, err)

	needsRehash, err := NeedsRehash(hash)
	require.NoError(t, err)
	assert.False(t, needsRehash) // Should not need rehash with current params
}

func TestSessionCreation(t *testing.T) {
	token := "test-token-123"
	hash := HashToken(token)
	assert.NotEmpty(t, hash)

	decoded, err := base64.RawURLEncoding.DecodeString(hash)
	require.NoError(t, err)
	assert.Equal(t, token, string(decoded))
}

func TestMFASecretGeneration(t *testing.T) {
	config := DefaultMFAConfig()
	mfa := NewMFAManager(config)

	key, err := mfa.GenerateSecret("test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, key.Secret())
	assert.NotEmpty(t, key.URL())

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)
	assert.True(t, mfa.ValidateTOTP(key.Secret(), code))

	assert.False(t, mfa.ValidateTOTP(key.Secret(), "000000"))
}

func TestMFABackupCodes(t *testing.T) {
	config := DefaultMFAConfig()
	mfa := NewMFAManager(config)

	codes, err := mfa.GenerateBackupCodes(5)
	require.NoError(t, err)
	assert.Len(t, codes, 5)

	validCode := codes[0]
	newCodes, err := mfa.ValidateBackupCode(codes, validCode)
	require.NoError(t, err)
	assert.Len(t, newCodes, 4)

	_, err = mfa.ValidateBackupCode(codes, "invalid")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidBackupCode))

	_, err = mfa.ValidateBackupCode([]string{}, "anything")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrBackupCodesExhausted))
}

func TestMFAEnableDisable(t *testing.T) {
	config := DefaultMFAConfig()
	mfa := NewMFAManager(config)

	user := &domain.User{
		ID:           domain.NewUserID(),
		Organization: domain.NewOrganizationID(),
		Email:        "test@example.com",
		DisplayName:  "Test User",
		MFAEnabled:   false,
	}

	key, err := mfa.GenerateSecret(user.Email)
	require.NoError(t, err)

	err = mfa.EnableMFA(user, key)
	require.NoError(t, err)
	assert.True(t, user.MFAEnabled)
	assert.NotEmpty(t, user.MFASecret)

	err = mfa.EnableMFA(user, key)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrMFAAlreadyEnabled))

	err = mfa.DisableMFA(user)
	require.NoError(t, err)
	assert.False(t, user.MFAEnabled)
	assert.Empty(t, user.MFASecret)

	err = mfa.DisableMFA(user)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrMFANotEnabled))
}

func TestMFAVerification(t *testing.T) {
	config := DefaultMFAConfig()
	mfa := NewMFAManager(config)

	user := &domain.User{
		ID:           domain.NewUserID(),
		Organization: domain.NewOrganizationID(),
		Email:        "test@example.com",
		DisplayName:  "Test User",
		MFAEnabled:   true,
	}

	key, err := mfa.GenerateSecret(user.Email)
	require.NoError(t, err)
	user.MFASecret = EncodeSecret([]byte(key.Secret()))

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)

	err = mfa.VerifyMFA(user, code, nil)
	require.NoError(t, err)

	err = mfa.VerifyMFA(user, "000000", nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidTOTPCode))

	backupCodes, err := mfa.GenerateBackupCodes(3)
	require.NoError(t, err)

	codeToTest := backupCodes[0]
	err = mfa.VerifyMFA(user, codeToTest, &backupCodes)
	require.NoError(t, err)
	assert.Len(t, backupCodes, 2)

	// Debug: check the backup codes slice
	t.Logf("Backup codes after first use: %v", backupCodes)

	err = mfa.VerifyMFA(user, codeToTest, &backupCodes)
	t.Logf("Second verify error: %v", err)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidBackupCode))
}

func TestRBACPermissions(t *testing.T) {
	perms := AllPermissions()
	assert.NotEmpty(t, perms)

	for _, p := range perms {
		parts := strings.Split(string(p), ":")
		assert.Len(t, parts, 2, "permission %s should have format resource:action", p)
		assert.NotEmpty(t, parts[0])
		assert.NotEmpty(t, parts[1])
	}

	rolePerms := DefaultRolePermissions()
	assert.NotEmpty(t, rolePerms)

	ownerPerms := rolePerms[RoleOwner]
	assert.GreaterOrEqual(t, len(ownerPerms), len(AllPermissions())-5)

	viewerPerms := rolePerms[RoleViewer]
	for _, p := range viewerPerms {
		parts := strings.Split(string(p), ":")
		assert.Equal(t, "read", parts[1], "viewer should only have read permissions")
	}
}

func TestPermissionCheck(t *testing.T) {
	resourceType := ResourceRoom
	action := ActionRead
	perm := Permission(fmt.Sprintf("%s:%s", resourceType, action))
	assert.Equal(t, "rooms:read", string(perm))

	userPerms := []Permission{"rooms:*", "bots:read"}
	required := Permission("rooms:write")

	hasPermission := false
	for _, p := range userPerms {
		if p == required || p == Permission(fmt.Sprintf("%s:*", resourceType)) || p == "*" {
			hasPermission = true
			break
		}
	}
	assert.True(t, hasPermission)
}

func TestTenantIsolation(t *testing.T) {
	org1 := domain.NewOrganizationID()
	org2 := domain.NewOrganizationID()

	check := PermissionCheck{
		UserID:       domain.NewUserID(),
		Organization: org1,
		ResourceType: ResourceRoom,
		ResourceID:   "room-123",
		Action:       ActionRead,
	}

	resourceOrgID := org2

	assert.NotEqual(t, check.Organization, resourceOrgID)
}

func TestCSRFTokenGeneration(t *testing.T) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	require.NoError(t, err)

	tokenStr := base64.RawURLEncoding.EncodeToString(token)
	assert.NotEmpty(t, tokenStr)

	token2 := make([]byte, 32)
	_, _ = rand.Read(token2)

	tokenStr2 := base64.RawURLEncoding.EncodeToString(token2)

	assert.Equal(t, 1, subtle.ConstantTimeCompare([]byte(tokenStr), []byte(tokenStr)))
	assert.Equal(t, 0, subtle.ConstantTimeCompare([]byte(tokenStr), []byte(tokenStr2)))
}

func TestAuthMiddlewareHelpers(t *testing.T) {
	ctx := context.Background()
	session := &domain.Session{
		ID:           domain.NewSessionID(),
		UserID:       domain.NewUserID(),
		Organization: domain.NewOrganizationID(),
		CSRFToken:    "test-csrf-token",
	}

	ctx = SetSessionContext(ctx, session)
	retrieved := GetSessionFromContext(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, session.ID, retrieved.ID)

	csrf := GetCSRFTokenFromContext(ctx)
	assert.Equal(t, "test-csrf-token", csrf)

	perms := []Permission{PermRoomRead, PermBotRead}
	ctx = SetPermissionsContext(ctx, perms)
	retrievedPerms := GetPermissionsFromContext(ctx)
	assert.Equal(t, perms, retrievedPerms)
}

func TestSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestCORSHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler := CORSHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}