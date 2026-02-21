package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileManager struct {
	basePath string
}

func NewLocalFM() *FileManager {
	return &FileManager{basePath: ".devserve"}
}

func NewGlobalFM() *FileManager {
	// hardcoded for now
	// return &FileManager{basePath: os.ExpandEnv("$HOME/.config/devserve")}
	return NewLocalFM()
}

func (fm *FileManager) InitDir() error {
	err := os.MkdirAll(fm.basePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (fm *FileManager) CreateFile(fileName string) (io.WriteCloser, error) {
	err := fm.InitDir()
	if err != nil {
		return nil, fmt.Errorf("Failed to init dir: %w", err)
	}

	file, err := os.Create(filepath.Join(fm.basePath, fileName))
	if err != nil {
		return nil, fmt.Errorf("couldn't create %s: %w", fileName, err)
	}
	return file, nil
}

func (fm *FileManager) CreateLogFiles() (stdOut, stdErr io.Writer, err error) {
	err = fm.InitDir()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to init dir: %w", err)
	}

	outFile, err := fm.CreateFile("out.log")
	if err != nil {
		return nil, nil, err
	}

	errFile, err := fm.CreateFile("err.log")
	if err != nil {
		return nil, nil, err
	}
	return outFile, errFile, nil
}
