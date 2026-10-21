package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GeneratePublicID(prefix string) string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
	}
	if prefix == "" {
		return hex.EncodeToString(bytes)
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
}
