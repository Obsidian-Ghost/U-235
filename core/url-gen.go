package core

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

func GenerateShortID(originalURL string) string {
	// Add timestamp for extra entropy
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	data := originalURL + timestamp

	// Generate MD5 hash
	hash := md5.Sum([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	// Return substring of hash with specified length
	if len(hashStr) > 5 {
		return hashStr[:5]
	}
	return hashStr

}
