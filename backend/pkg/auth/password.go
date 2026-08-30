package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func HashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	saltHex := hex.EncodeToString(salt)
	derived := pbkdf2.Key([]byte(password), []byte(saltHex), 120000, 32, sha256.New)
	return fmt.Sprintf("pbkdf2_sha256$120000$%s$%s", saltHex, hex.EncodeToString(derived))
}

func VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt := parts[2]
	expected := parts[3]
	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, 32, sha256.New)
	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(derived)), []byte(expected)) == 1
}
