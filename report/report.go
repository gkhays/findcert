package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"org.gkh/certfinder/ui"
)

// Certificate file extensions to search for
var certExtensions = []string{
	".pem",
	".der",
	".crt",
	".cer",
	".pkcs12",
	".jks",
	".bcfks",
}

// FileInfo represents information about a found certificate file
type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
}

// ExtensionResult represents files found for each extension
type ExtensionResult struct {
	Type  string     `json:"type"`
	Files []FileInfo `json:"files"`
}

// SearchResult represents the complete search results
type SearchResult struct {
	TotalFiles int               `json:"total_files"`
	Results    []ExtensionResult `json:"results"`
	SearchTime time.Time         `json:"search_time"`
}

func Execute(path string, outputFile string) {
	// Create slice to store results for each extension
	results := make([]ExtensionResult, len(certExtensions))
	for i, ext := range certExtensions {
		results[i] = ExtensionResult{Type: ext}
	}

	skipList := ""

	spinner := ui.NewSpinner()
	spinner.Start("Searching for certificate files...")

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == skipList {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		//fmt.Printf("Searching %q: %s\n", path, info.Name())

		// Skip hidden directories and files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches any certificate extension
		for i, ext := range certExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				fileInfo := FileInfo{
					Path:         path,
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
				}
				results[i].Files = append(results[i].Files, fileInfo)
				break
			}
		}

		return nil
	})

	spinner.Stop()

	if err != nil {
		fmt.Printf("Error walking directory tree: %v\n", err)
		os.Exit(1)
	}

	// Count total files and create final result
	totalFiles := 0
	for _, result := range results {
		totalFiles += len(result.Files)
	}

	searchResult := SearchResult{
		TotalFiles: totalFiles,
		Results:    results,
		SearchTime: time.Now(),
	}

	// Print results to console
	for _, result := range results {
		fmt.Printf("%sFiles with extension %s:%s\n", ui.ColorGreen, result.Type, ui.ColorReset)
		if len(result.Files) == 0 {
			fmt.Println("No files found")
		} else {
			for _, file := range result.Files {
				displayPath := file.Path
				if len(file.Path) > 50 {
					displayPath = file.Path[:23] + "..." + file.Path[len(file.Path)-24:]
				}
				fmt.Printf("%s (Size: %d bytes, Modified: %s)\n",
					displayPath, file.Size, file.ModifiedTime.Format(time.RFC3339))
			}
		}
		fmt.Println()
	}

	// Write results to JSON file
	jsonData, err := json.MarshalIndent(searchResult, "", "  ")
	if err != nil {
		fmt.Printf("Error creating JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON file: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Printf("%sSummary:%s\n", ui.ColorYellow, ui.ColorReset)
	fmt.Printf("Total certificate files found: %d\n", totalFiles)
	fmt.Println("Results have been saved to results.json")
}
