package telegram

import (
	"crypto/rand"
	"time"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0,O,1,I
const codeLen = 8

func generateCode() string {
	b := make([]byte, codeLen)
	_, _ = rand.Read(b)
	out := make([]byte, codeLen)
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out)
}

func CreateLinkCode(userID int64) (string, string, error) {
	code := generateCode()
	h := hashCode(code)
	exp := time.Now().UTC().Add(15 * time.Minute).Format("2006-01-02 15:04:05")
	if err := InsertLinkCode(userID, h, exp); err != nil {
		return "", "", err
	}
	return code, exp, nil
}
