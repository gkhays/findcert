package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"org.gkh/certfinder/report"
	"org.gkh/certfinder/ui"
)

func main() {
	fmt.Printf("%sCertificate File Finder%s\n", ui.ColorYellow, ui.ColorReset)

	searchPath := flag.String("path", ".", "Directory path to search")
	outputFile := flag.String("output", "results.json", "Output JSON file path")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s (%s on %s/%s; %s)\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		os.Exit(0)
	}

	absPath, err := filepath.Abs(*searchPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n'", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Error: Pa;th does not exist: %s\n", absPath)
		os.Exit(1)
	}

	report.Execute(absPath, *outputFile)
}
