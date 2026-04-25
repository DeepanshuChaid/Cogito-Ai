package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

type Cache struct {
	Files map[string]string `json:"files"` // Path -> Hash
}

func LoadCache() Cache {
	var c Cache
	c.Files = make(map[string]string)
	data, err := os.ReadFile(".cogito/cache.json")
	if err == nil {
		json.Unmarshal(data, &c)
	}
	return c
}

func SaveCache(c Cache) {
	os.MkdirAll(".cogito", os.ModePerm)
	data, _ := json.MarshalIndent(c, "", "  ")
	os.WriteFile(".cogito/cache.json", data, 0644)
}

func GetFileHash(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
