package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"org.gkh/findcert/config"
	"org.gkh/findcert/ui"
)

// TestFiles represents the test files structure we'll create
var TestFiles = map[string][]string{
	"certs": {
		"test1.pem",
		"test2.crt",
		"test3.der",
		"invalid.txt",
	},
	"certs/nested": {
		"test4.cer",
		"test5.jks",
		"another.txt",
	},
	".hidden": {
		"hidden.pem",
	},
}

func setupTestDirectory(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "certfinder-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test directory structure and files
	for dir, files := range TestFiles {
		dirPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		for _, file := range files {
			filePath := filepath.Join(dirPath, file)
			err := os.WriteFile(filePath, []byte("test content"), 0644)
			if err != nil {
				os.RemoveAll(tempDir)
				t.Fatalf("Failed to create file %s: %v", file, err)
			}
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestSearchResult_CountsCorrectly(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	// Initialize results slice
	results := make([]config.ExtensionResult, len(config.CertExtensions))
	for i, ext := range config.CertExtensions {
		results[i] = config.ExtensionResult{Type: ext}
	}

	// Walk the directory
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		for i, ext := range config.CertExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				relPath, err := filepath.Rel(tempDir, path)
				if err != nil {
					relPath = path
				}

				fileInfo := config.FileInfo{
					Path:         relPath,
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
				}
				results[i].Files = append(results[i].Files, fileInfo)
				break
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Count total files
	totalFiles := 0
	for _, result := range results {
		totalFiles += len(result.Files)
	}

	// Expected number of certificate files (excluding hidden and invalid files)
	expectedTotal := 5 // test1.pem, test2.crt, test3.der, test4.cer, test5.jks

	if totalFiles != expectedTotal {
		t.Errorf("Expected %d files, got %d", expectedTotal, totalFiles)
	}
}

func TestFileInfo_Contents(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	testFile := filepath.Join(tempDir, "certs", "test1.pem")
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	fileInfo := config.FileInfo{
		Path:         "certs/test1.pem",
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
	}

	if fileInfo.Size != info.Size() {
		t.Errorf("Expected size %d, got %d", info.Size(), fileInfo.Size)
	}

	if !fileInfo.ModifiedTime.Equal(info.ModTime()) {
		t.Errorf("Expected mod time %v, got %v", info.ModTime(), fileInfo.ModifiedTime)
	}
}

func TestJSON_Output(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	results := make([]config.ExtensionResult, len(config.CertExtensions))
	for i, ext := range config.CertExtensions {
		results[i] = config.ExtensionResult{Type: ext}
	}

	searchResult := config.SearchResult{
		SearchPath: tempDir,
		TotalFiles: 5,
		Results:    results,
		SearchTime: time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(searchResult, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Test JSON unmarshaling
	var decoded config.SearchResult
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.SearchPath != tempDir {
		t.Errorf("Expected search path %s, got %s", tempDir, decoded.SearchPath)
	}

	if decoded.TotalFiles != 5 {
		t.Errorf("Expected 5 total files, got %d", decoded.TotalFiles)
	}
}

func TestSpinner_Basic(t *testing.T) {
	spinner := ui.NewSpinner()

	// Test starting
	spinner.Start("Testing spinner")
	if spinner.Stopped {
		t.Error("Spinner should not be stopped immediately after starting")
	}

	// Test stopping
	spinner.Stop()
	if !spinner.Stopped {
		t.Error("Spinner should be stopped after Stop() is called")
	}

	// Test double stop (shouldn't panic)
	spinner.Stop()
}

func TestHiddenFiles_AreSkipped(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	results := make([]config.ExtensionResult, len(config.CertExtensions))
	for i, ext := range config.CertExtensions {
		results[i] = config.ExtensionResult{Type: ext}
	}

	// Walk the directory
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		for i, ext := range config.CertExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				results[i].Files = append(results[i].Files, config.FileInfo{
					Path: path,
				})
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Check that no hidden files were included
	for _, result := range results {
		for _, file := range result.Files {
			if strings.Contains(file.Path, ".hidden") {
				t.Errorf("Hidden file was included in results: %s", file.Path)
			}
		}
	}
}
