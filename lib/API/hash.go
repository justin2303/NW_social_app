package API

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashGUID(guid string) string {
	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Write the GUID to the hasher
	hasher.Write([]byte(guid))

	// Get the final hash sum
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
