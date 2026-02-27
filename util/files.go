package util

import (
	"os"
	"strings"
)

// LastNLines reads a file and returns the last n lines.
// Returns an empty slice if the file doesn't exist or is empty.
func LastNLines(path string, n int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	content := strings.TrimRight(string(data), "\n")
	if content == "" {
		return nil, nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}
