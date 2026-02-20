package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func saveProcess(fm *FileManager, p *Process) error {
	file, err := fm.CreateFile(strconv.Itoa(p.Port) + ".json")
	if err != nil {
		return fmt.Errorf("Failed to create process file: %w", err)
	}
	defer file.Close()

	bytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("Failed marshalling json: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("Failed write json: %w", err)
	}

	return nil
}

func GetProcessByPort(port int) (*Process, error) {
	fm := NewGlobalFM()
	entries, err := os.ReadDir(fm.basePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if portStr, found := strings.CutSuffix(entry.Name(), ".json"); found {
			portNum, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			if portNum == port {
				p, err := generateProcess(entry.Name())
				if err != nil {
					return nil, fmt.Errorf("Couldn't generate process: %w", err)
				}
				return p, nil
			}
		}
	}
	return nil, fmt.Errorf("Couldn't find any processes")
}

func generateProcess(fileName string) (*Process, error) {
	fm := NewGlobalFM()
	file, err := os.Open(filepath.Join(fm.basePath, fileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed reading file: %w", err)
	}
	var p Process
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal file: %w", err)
	}

	return &p, nil
}

func ListProcesses() ([]int, error) {
	fm := NewGlobalFM()
	entries, err := os.ReadDir(fm.basePath)
	if err != nil {
		return []int{}, fmt.Errorf("Failed to read directory: %w", err)
	}

	var ports []int
	for _, entry := range entries {
		if portStr, found := strings.CutSuffix(entry.Name(), ".json"); found {
			portNum, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			ports = append(ports, portNum)
		}
	}

	return ports, nil
}

func RemoveProcess(port int) error {
	fm := NewGlobalFM()
	pathstr := strconv.Itoa(port)
	return os.Remove(filepath.Join(fm.basePath, pathstr+".json"))
}
