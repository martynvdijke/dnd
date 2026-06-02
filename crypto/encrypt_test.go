package crypto

import (
	"os"
	"testing"
)

const testKey1 = "0000000000000000000000000000000000000000000000000000000000000001"
const testKey2 = "0000000000000000000000000000000000000000000000000000000000000002"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	os.Setenv("AI_ENCRYPTION_KEY", testKey1)
	Init()

	plaintext := "sk-test-api-key-12345"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("round trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	os.Setenv("AI_ENCRYPTION_KEY", testKey1)
	Init()

	plaintext := "test-key-value"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	os.Setenv("AI_ENCRYPTION_KEY", testKey2)
	Init()

	_, err = Decrypt(encrypted)
	if err == nil {
		t.Fatal("expected decrypt to fail with wrong key")
	}
}

func TestFallbackMode(t *testing.T) {
	os.Unsetenv("AI_ENCRYPTION_KEY")
	Init()

	plaintext := "sk-plaintext-key"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt in fallback mode failed: %v", err)
	}
	if encrypted != plaintext {
		t.Fatalf("fallback encrypt should return plaintext unchanged")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt in fallback mode failed: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("fallback decrypt mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEmptyPlaintext(t *testing.T) {
	os.Setenv("AI_ENCRYPTION_KEY", testKey1)
	Init()

	encrypted, err := Encrypt("")
	if err != nil {
		t.Fatalf("encrypt empty string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt empty string failed: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}

func TestDecryptTooShort(t *testing.T) {
	os.Setenv("AI_ENCRYPTION_KEY", testKey1)
	Init()

	_, err := Decrypt("aabb")
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}
