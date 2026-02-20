package internal

import (
	"errors"
	"os"
)

var LockToPM = map[string]string{
	"package-lock.json": "npm",
	"pnpm-lock.yaml":    "pnpm",
	"yarn.lock":         "yarn",
	"bun.lock":          "bun",
	"bun.lockb":         "bun",
}

var PMToCommand = map[string]string{
	"npm":  "npm run dev",
	"pnpm": "pnpm run dev",
	"yarn": "yarn dev",
	"bun":  "bun run dev",
}

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
