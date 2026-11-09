// Package token issues and verifies the bearer secrets the framework hands to
// agents: restore claim tokens, agent registration tokens and the long-lived
// per-agent tokens.
//
// A token is random enough that the database need only keep its SHA-256
// digest. That is not how a password is stored — a password is low-entropy and
// needs a slow hash — but for a 32-byte random value a digest is sufficient and
// keeps the framework from holding anything it could leak.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

// tokenBytes is the entropy of a token.
const tokenBytes = 32

// minFileTokenLength is the shortest token accepted from a file. The digest
// below is only sufficient for a random value: "openssl rand -hex 32" prints
// 64 characters, and something short was typed by hand.
const minFileTokenLength = 32

// New returns a random token. Unlike model.NewID it has no time-based
// substitute: a token that cannot be made unpredictable is an error.
func New() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Hash returns the SHA-256 digest stored in place of the token.
func Hash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// Matches reports whether token hashes to hash, in constant time.
func Matches(hash []byte, token string) bool {
	return subtle.ConstantTimeCompare(hash, Hash(token)) == 1
}

// LoadFile reads a token an operator keeps in a file and returns its digest,
// so that the plain token does not stay in the process's memory.
func LoadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("no token file is configured")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	plain := strings.TrimSpace(string(data))
	if len(plain) < minFileTokenLength {
		return nil, fmt.Errorf("token file %s holds fewer than %d characters: generate one with \"openssl rand -hex 32\"",
			path, minFileTokenLength)
	}
	return Hash(plain), nil
}
