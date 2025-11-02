package app

import (
	"fmt"
	"os"
	"strings"

)



func derefInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func derefString(ptr *string) string {
	if ptr == nil {
		return "Unknown"
	}
	return *ptr
}

func GetServerIdFromPath(path string) (string, error) {
	// Example path: C:\Games\Steam\steamapps\common\Sandstorm Dedicated Server\Server1\Logs
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", path)
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, "backup") {
			continue
		}
		return strings.Trim(name, ".log"), nil
	}
	return "", fmt.Errorf("could not determine server ID from path: %s", path)
}
