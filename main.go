package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"org.gkh/findcert/cli"
	"org.gkh/findcert/cmd"
	"org.gkh/findcert/pkg"
	"org.gkh/findcert/ui"
)

func main() {
	fmt.Printf("%sCertificate File Finder%s\n", ui.ColorYellow, ui.ColorReset)

	searchPath := flag.String("path", ".", "Directory path to search")
	outputFile := flag.String("output", "results.json", "Output JSON file path")
	showVersion := flag.Bool("version", false, "Show version information")
	listNoExt := flag.Bool("noext", false, "List files with no extension")

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s (%s on %s/%s; %s)\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		os.Exit(0)
	}

	if *listNoExt {
		if len(*searchPath) > 0 {
			fmt.Printf("Listing files with no extension...\n")
			results, err := cmd.ListNoExt(*searchPath)
			if err != nil {
				fmt.Printf("%v", err)
				os.Exit(1)
			}
			fmt.Printf("Found %d\n", len(results))
			for _, file := range results {
				fmt.Println(file.Path)
				filetype, err := pkg.GetFileType(file.Path)
				if err != nil {
					fmt.Printf("%v\n", err)
				}
				fmt.Printf("  - %s, %s\n", filetype.Extension, filetype.Description)
			}
			os.Exit(0)
		}
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

	cli.Execute(absPath, *outputFile)
}
