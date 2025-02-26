package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"org.gkh/findcert/config"
)

func ListCertificates(path string) ([]config.ExtensionResult, error) {
	results := make([]config.ExtensionResult, len(config.CertExtensions))
	for i, ext := range config.CertExtensions {
		results[i] = config.ExtensionResult{Type: ext}
	}

	// TODO(gkh) - provide a "skip" list
	skipList := ""

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
		for i, ext := range config.CertExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				fileInfo := config.FileInfo{
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

	return results, err
}
