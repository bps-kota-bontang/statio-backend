package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// XLSXConverterService handles XLSX to XLS conversion via Node.js subprocess
type XLSXConverterService struct {
	scriptPath string
}

// NewXLSXConverterService creates a new XLSX converter service
func NewXLSXConverterService() (*XLSXConverterService, error) {
	// Determine the script path relative to the binary
	scriptPath, err := findConverterScript()
	if err != nil {
		return nil, fmt.Errorf("converter script not found: %w", err)
	}

	// Verify Node.js is installed
	if _, err := exec.LookPath("node"); err != nil {
		return nil, fmt.Errorf("node.js not found in PATH: %w", err)
	}

	return &XLSXConverterService{
		scriptPath: scriptPath,
	}, nil
}

// ConvertXLSXToXLS converts XLSX bytes to XLS bytes using Node.js subprocess
func (s *XLSXConverterService) ConvertXLSXToXLS(xlsxData []byte) ([]byte, error) {
	if len(xlsxData) == 0 {
		return nil, fmt.Errorf("empty XLSX data")
	}

	// Create command to run Node.js script
	cmd := exec.Command("node", s.scriptPath)

	// Prepare pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start converter process: %w", err)
	}

	// Write XLSX data to stdin in a goroutine
	go func() {
		defer stdin.Close()
		if _, err := io.Copy(stdin, bytes.NewReader(xlsxData)); err != nil {
			// Log error but don't return (process already started)
			fmt.Fprintf(os.Stderr, "failed to write to stdin: %v\n", err)
		}
	}()

	// Read XLS data from stdout
	var xlsBuffer bytes.Buffer
	if _, err := io.Copy(&xlsBuffer, stdout); err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	// Read any errors from stderr
	var errBuffer bytes.Buffer
	if _, err := io.Copy(&errBuffer, stderr); err != nil {
		return nil, fmt.Errorf("failed to read stderr: %w", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		errMsg := errBuffer.String()
		if errMsg != "" {
			return nil, fmt.Errorf("converter error: %s", errMsg)
		}
		return nil, fmt.Errorf("converter process failed: %w", err)
	}

	xlsData := xlsBuffer.Bytes()
	if len(xlsData) == 0 {
		return nil, fmt.Errorf("converter returned empty data")
	}

	return xlsData, nil
}

// findConverterScript locates the Node.js converter script
func findConverterScript() (string, error) {
	// Try multiple possible locations
	possiblePaths := []string{
		"./scripts/xlsx-to-xls.js",                                     // From binary location
		"../scripts/xlsx-to-xls.js",                                    // One level up
		"../../scripts/xlsx-to-xls.js",                                 // Two levels up
		filepath.Join(getExecutableDir(), "scripts", "xlsx-to-xls.js"), // Relative to executable
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("xlsx-to-xls.js not found in any expected location")
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() string {
	if runtime.GOOS == "windows" {
		// For Windows, use executable path
		exe, err := os.Executable()
		if err == nil {
			return filepath.Dir(exe)
		}
	}

	// For development, use working directory
	wd, err := os.Getwd()
	if err == nil {
		return wd
	}

	return "."
}
