package auth

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFA errors
var (
	ErrMFANotEnabled      = errors.New("MFA not enabled")
	ErrInvalidTOTPCode    = errors.New("invalid TOTP code")
	ErrInvalidBackupCode  = errors.New("invalid backup code")
	ErrBackupCodesExhausted = errors.New("no backup codes remaining")
	ErrMFAAlreadyEnabled  = errors.New("MFA already enabled")
	ErrMFASecretInvalid   = errors.New("invalid MFA secret")
)

// MFAConfig holds configuration for MFA
type MFAConfig struct {
	Issuer      string
	Digits      int
	Period      uint
	Algorithm   otp.Algorithm
	BackupCodes int // Number of backup codes to generate
}

// DefaultMFAConfig returns a default MFA configuration
func DefaultMFAConfig() MFAConfig {
	return MFAConfig{
		Issuer:      "CFC Bot Manager",
		Digits:      6,
		Period:      30,
		Algorithm:   otp.AlgorithmSHA256,
		BackupCodes: 10,
	}
}

// MFAManager handles TOTP-based MFA
type MFAManager struct {
	config MFAConfig
}

// NewMFAManager creates a new MFA manager
func NewMFAManager(config MFAConfig) *MFAManager {
	return &MFAManager{config: config}
}

// GenerateSecret generates a new TOTP secret for a user
func (m *MFAManager) GenerateSecret(accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      m.config.Issuer,
		AccountName: accountName,
		Digits:      otp.Digits(m.config.Digits),
		Period:      m.config.Period,
		Algorithm:   m.config.Algorithm,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP secret: %w", err)
	}
	return key, nil
}

// ValidateTOTP validates a TOTP code against the secret
func (m *MFAManager) ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates a set of backup codes
func (m *MFAManager) GenerateBackupCodes(count int) ([]string, error) {
	if count <= 0 {
		count = m.config.BackupCodes
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := make([]byte, 8) // 8 bytes = 16 hex chars
		if _, err := rand.Read(code); err != nil {
			return nil, fmt.Errorf("generate backup code: %w", err)
		}
		codes[i] = fmt.Sprintf("%x", code)
	}
	return codes, nil
}

// ValidateBackupCode checks if a backup code is valid and removes it from the list
func (m *MFAManager) ValidateBackupCode(backupCodes []string, code string) ([]string, error) {
	if len(backupCodes) == 0 {
		return backupCodes, ErrBackupCodesExhausted
	}
	code = strings.ToLower(strings.TrimSpace(code))
	for i, c := range backupCodes {
		if strings.EqualFold(c, code) {
			// Remove the used code
			newCodes := append(backupCodes[:i], backupCodes[i+1:]...)
			return newCodes, nil
		}
	}
	return backupCodes, ErrInvalidBackupCode
}

// GetTOTPURI returns the otpauth URI for QR code generation
func (m *MFAManager) GetTOTPURI(key *otp.Key, accountName string) string {
	return key.URL()
}

// EncodeSecret encodes a secret for storage (base32)
func EncodeSecret(secret []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
}

// DecodeSecret decodes a base32-encoded secret
func DecodeSecret(encoded string) ([]byte, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(encoded)
}

// MFAStatus represents the MFA status for a user
type MFAStatus struct {
	Enabled      bool
	Secret       string // base32 encoded
	BackupCodes  []string
	LastUsed     *time.Time
	RecoveryHint string // e.g., last 4 digits of phone
}

// EnableMFA enables MFA for a user
func (m *MFAManager) EnableMFA(user *domain.User, key *otp.Key) error {
	if user.MFAEnabled {
		return ErrMFAAlreadyEnabled
	}

	user.MFASecret = EncodeSecret([]byte(key.Secret()))
	user.MFAEnabled = true
	return nil
}

// DisableMFA disables MFA for a user
func (m *MFAManager) DisableMFA(user *domain.User) error {
	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}

	user.MFASecret = ""
	user.MFAEnabled = false
	return nil
}

// VerifyMFA verifies either a TOTP code or a backup code
func (m *MFAManager) VerifyMFA(user *domain.User, code string, backupCodes *[]string) error {
	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}

	code = strings.TrimSpace(code)
	if len(code) == 0 {
		return ErrInvalidTOTPCode
	}

	// First try TOTP
	secret, err := DecodeSecret(user.MFASecret)
	if err != nil {
		return ErrMFASecretInvalid
	}

	if m.ValidateTOTP(string(secret), code) {
		now := time.Now()
		user.LastLoginAt = &now
		return nil
	}

	// Then try backup codes
	if backupCodes != nil && len(*backupCodes) > 0 {
		newCodes, err := m.ValidateBackupCode(*backupCodes, code)
		if err == nil {
			*backupCodes = newCodes
			now := time.Now()
			user.LastLoginAt = &now
			return nil
		}
		// Return the actual backup code error
		return err
	}

	return ErrInvalidTOTPCode
}