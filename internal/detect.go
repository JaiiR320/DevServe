package internal

import (
	"errors"
	"os"
)

func DetectPackageManager() (string, error) {
	for f, c := range LockToPM {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		return c, nil
	}
	err := "failed to detect package\nsupported package managers and lockfiles:"
	for l, p := range LockToPM {
		err = err + "\n" + p + " : " + l
	}

	return "", errors.New(err)
}
