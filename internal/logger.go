package internal

import (
	"log"
	"os"
)

func InitLogger() error {
	f, err := os.OpenFile("/tmp/devserve.daemon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime)
	return nil
}
