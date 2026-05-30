package tui

import (
	"fmt"
	"os"

	osc52 "github.com/aymanbagabas/go-osc52/v2"
)

func copyURL(url string) error {
	if _, err := fmt.Fprint(os.Stderr, osc52.New(url)); err != nil {
		return fmt.Errorf("failed to copy URL: %w", err)
	}
	return nil
}
