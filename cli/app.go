package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"org.gkh/findcert/cmd"
	"org.gkh/findcert/config"
	"org.gkh/findcert/ui"
)

func Execute(path string, outputFile string) {
	spinner := ui.NewSpinner()
	spinner.Start("Searching for certificate files...")

	results, err := cmd.ListCertificates(path)

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

	searchResult := config.SearchResult{
		SearchPath: path,
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
