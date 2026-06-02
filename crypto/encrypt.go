package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var encryptionKey []byte
var fallbackMode bool

func init() {
	Init()
}

// Init (re-)initializes the crypto package from the AI_ENCRYPTION_KEY env var.
// Exposed so tests can call it after changing the environment.
func Init() {
	keyHex := os.Getenv("AI_ENCRYPTION_KEY")
	if keyHex == "" {
		log.Println("[crypto] AI_ENCRYPTION_KEY not set — API keys will be stored in plaintext")
		fallbackMode = true
		return
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		log.Printf("[crypto] AI_ENCRYPTION_KEY is not valid hex — falling back to plaintext: %v", err)
		fallbackMode = true
		return
	}

	if len(key) != 32 {
		log.Printf("[crypto] AI_ENCRYPTION_KEY must be 32 bytes (64 hex chars), got %d bytes — falling back to plaintext", len(key))
		fallbackMode = true
		return
	}

	fallbackMode = false
	encryptionKey = key
}

// Encrypt encrypts plaintext using AES-256-GCM. Returns hex-encoded ciphertext.
func Encrypt(plaintext string) (string, error) {
	if fallbackMode {
		return plaintext, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a hex-encoded AES-256-GCM ciphertext back to plaintext.
func Decrypt(cipherHex string) (string, error) {
	if fallbackMode {
		return cipherHex, nil
	}

	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", fmt.Errorf("decode hex: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}
