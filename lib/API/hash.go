package API

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"
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

func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	var result string
	for i := 0; i < 8; i++ {
		// Generate a random number between 0 and 9
		num := rand.Intn(10)
		result += strconv.Itoa(num)
	}
	return result
}
