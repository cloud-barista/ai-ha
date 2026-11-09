package token

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestNewIsUnpredictable(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		tok, err := New()
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if seen[tok] {
			t.Fatalf("New returned a repeat: %q", tok)
		}
		seen[tok] = true

		raw, err := base64.RawURLEncoding.DecodeString(tok)
		if err != nil {
			t.Fatalf("token is not base64url: %v", err)
		}
		if len(raw) != tokenBytes {
			t.Fatalf("token carries %d bytes, want %d", len(raw), tokenBytes)
		}
	}
}

// A token is URL-safe and unpadded so that it survives being pasted into a
// shell assignment, an env file or a header without quoting.
func TestNewIsSafeToPaste(t *testing.T) {
	for range 100 {
		tok, err := New()
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		for _, c := range tok {
			isSafe := c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' ||
				c >= '0' && c <= '9' || c == '-' || c == '_'
			if !isSafe {
				t.Fatalf("token %q contains %q, which needs quoting", tok, c)
			}
		}
	}
}

func TestMatches(t *testing.T) {
	tok, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	hash := Hash(tok)

	if !Matches(hash, tok) {
		t.Error("a token did not match its own hash")
	}
	if Matches(hash, tok+"x") {
		t.Error("a different token matched")
	}
	if Matches(hash, "") {
		t.Error("the empty string matched")
	}
	if Matches(nil, tok) {
		t.Error("a nil hash matched")
	}
}

// The digest, not the token, is what the database holds.
func TestHashDoesNotRevealTheToken(t *testing.T) {
	const tok = "a-token-value"
	hash := Hash(tok)

	if len(hash) != 32 {
		t.Fatalf("hash is %d bytes, want 32 (SHA-256)", len(hash))
	}
	if string(hash) == tok {
		t.Fatal("the hash is the token")
	}
	// Deterministic: the same token must find the same row on every lookup.
	if string(Hash(tok)) != string(hash) {
		t.Fatal("Hash is not deterministic")
	}
}

func writeTokenFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "admin.token")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// The trailing newline "openssl rand" leaves must not become part of the token,
// or the token an operator copies from the file would never match.
func TestLoadFileIgnoresSurroundingWhitespace(t *testing.T) {
	const plain = "0123456789abcdef0123456789abcdef"

	hash, err := LoadFile(writeTokenFile(t, plain+"\n"))
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if !Matches(hash, plain) {
		t.Fatal("the token in the file does not match its digest")
	}
}

func TestLoadFileRefusesWhatCannotBeATrustedToken(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"not configured", ""},
		{"missing", filepath.Join(t.TempDir(), "absent")},
		{"empty", writeTokenFile(t, "\n")},
		{"typed by hand", writeTokenFile(t, "changeme\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadFile(tt.path); err == nil {
				t.Fatal("LoadFile accepted it")
			}
		})
	}
}
