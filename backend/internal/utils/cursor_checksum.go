// ref: _ref/9router/open-sse/utils/cursorChecksum.js
package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

func GenerateHashed64Hex(input string, salt string) string {
	data := input + salt
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GenerateSessionID(authToken string) string {
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	return uuid.NewSHA1(namespace, []byte(authToken)).String()
}

func GenerateCursorChecksum(machineID string) string {
	timestamp := time.Now().UnixMilli() / 1000000

	byteArray := []byte{
		byte((timestamp >> 40) & 0xFF),
		byte((timestamp >> 32) & 0xFF),
		byte((timestamp >> 24) & 0xFF),
		byte((timestamp >> 16) & 0xFF),
		byte((timestamp >> 8) & 0xFF),
		byte(timestamp & 0xFF),
	}

	t := byte(165)
	for i := 0; i < len(byteArray); i++ {
		byteArray[i] = byte((int(byteArray[i])^int(t))+(i%256)) & 0xFF
		t = byteArray[i]
	}

	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var encoded []byte

	for i := 0; i < len(byteArray); i += 3 {
		a := byteArray[i]
		b := byte(0)
		c := byte(0)

		if i+1 < len(byteArray) {
			b = byteArray[i+1]
		}
		if i+2 < len(byteArray) {
			c = byteArray[i+2]
		}

		encoded = append(encoded, alphabet[a>>2])
		encoded = append(encoded, alphabet[((a&3)<<4)|(b>>4)])

		if i+1 < len(byteArray) {
			encoded = append(encoded, alphabet[((b&15)<<2)|(c>>6)])
		}
		if i+2 < len(byteArray) {
			encoded = append(encoded, alphabet[c&63])
		}
	}

	return string(encoded) + machineID
}

type CursorHeaders struct {
	Authorization string
	ContentType   string
	UserAgent     string
	Checksum      string
}

func BuildCursorHeaders(accessToken, machineID string) CursorHeaders {
	return CursorHeaders{
		Authorization: "Bearer " + accessToken,
		ContentType:   "application/connect+proto",
		UserAgent:     "cursor/0.48.7",
		Checksum:      GenerateCursorChecksum(machineID),
	}
}

func Base64URLEncode(data []byte) string {
	encoded := base64.URLEncoding.EncodeToString(data)
	for len(encoded) > 0 && encoded[len(encoded)-1] == '=' {
		encoded = encoded[:len(encoded)-1]
	}
	return encoded
}

func Base64URLDecode(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
