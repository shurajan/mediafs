package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"
)

var SharedSecret = []byte("change-me")

func GenerateSignature(path string, expires int64) string {
	message := fmt.Sprintf("%s|%d", path, expires)
	h := hmac.New(sha256.New, SharedSecret)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySignature(path string, expires int64, sig string) bool {
	expected := GenerateSignature(path, expires)
	return hmac.Equal([]byte(expected), []byte(sig))
}

func Generate24hPublicLink(baseURL, path string) string {
	expires := time.Now().Add(24 * time.Hour).Unix()
	sig := GenerateSignature(path, expires)
	return fmt.Sprintf("%s/api/media/public/stream?path=%s&expires=%d&sig=%s",
		baseURL, url.QueryEscape(path), expires, sig)
}
