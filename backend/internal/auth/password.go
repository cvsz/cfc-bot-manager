package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP recommended as of 2023)
const (
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLength    = 16
)

// Password hashing errors
var (
	ErrInvalidHashFormat   = errors.New("invalid hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
	ErrHashVerification    = errors.New("password verification failed")
)

// HashPassword creates an Argon2id hash of the password
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("password cannot be empty")
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword checks if the password matches the hash
func VerifyPassword(password, encodedHash string) error {
	if len(password) == 0 {
		return ErrHashVerification
	}

	// Parse the encoded hash
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return ErrInvalidHashFormat
	}

	// Verify argon2 version
	if params.version != argon2.Version {
		return ErrIncompatibleVersion
	}

	// Compute hash with provided password
	computedHash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(hash)))

	// Constant-time comparison
	if subtle.ConstantTimeCompare(computedHash, hash) == 1 {
		return nil
	}

	return ErrHashVerification
}

// HashParams holds the parsed Argon2 parameters
type HashParams struct {
	version int
	memory  uint32
	time    uint32
	threads uint8
}

// decodeHash parses an encoded Argon2 hash string
func decodeHash(encodedHash string) (*HashParams, []byte, []byte, error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	// Parse parameters: m=65536,t=3,p=4
	params := &HashParams{version: version}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.threads); err != nil {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	return params, salt, hash, nil
}

// NeedsRehash checks if the hash should be updated with current parameters
func NeedsRehash(encodedHash string) (bool, error) {
	params, _, _, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Check if parameters match current defaults
	return params.memory != argon2Memory || params.time != argon2Time || params.threads != argon2Threads, nil
}