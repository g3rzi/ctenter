package inject

import (
	"crypto/sha256"
	"fmt"
)

func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
