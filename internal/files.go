package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type FileManager struct {
	basePath string
}

func NewLocalFM() *FileManager {
	return &FileManager{basePath: ".devserve"}
}

func NewGlobalFM() *FileManager {
	// hardcoded for now
	return &FileManager{basePath: os.ExpandEnv("$HOME/.config/devserve")}
}

func (fm *FileManager) InitDir() error {
	err := os.MkdirAll(fm.basePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (fm *FileManager) createFile(fileName string) (io.WriteCloser, error) {
	file, err := os.Create(filepath.Join(fm.basePath, fileName))
	if err != nil {
		return nil, fmt.Errorf("couldn't create %s: %w", fileName, err)
	}
	return file, nil
}

func (fm *FileManager) CreateLogFiles() (stdOut, stdErr io.Writer, err error) {
	outFile, err := fm.createFile("out.log")
	if err != nil {
		return nil, nil, err
	}

	errFile, err := fm.createFile("err.log")
	if err != nil {
		return nil, nil, err
	}
	return outFile, errFile, nil
}

func (fm *FileManager) SavePID(pid int) error {
	file, err := fm.createFile("pid.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	pidStr := strconv.Itoa(pid)
	_, err = file.Write([]byte(pidStr))
	if err != nil {
		return err
	}
	return nil
}
