package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetUserID(v interface{}) uint {
	switch id := v.(type) {
	case uint:
		return id
	case float64:
		return uint(id)
	case int:
		return uint(id)
	default:
		return 0
	}
}

// DirEntry represents a directory entry for listing
type DirEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ListDirectories lists subdirectories in the given path
func ListDirectories(path string) ([]DirEntry, error) {
	// Clean the path
	path = filepath.Clean(path)

	// Open the directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []DirEntry
	for _, entry := range entries {
		// Only include directories
		if entry.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			dirs = append(dirs, DirEntry{
				Name: entry.Name(),
				Path: filepath.Join(path, entry.Name()),
			})
		}
	}

	return dirs, nil
}
