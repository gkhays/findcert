package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"org.gkh/findcert/config"
)

func ListNoExt(dirPath string) ([]config.FileInfo, error) {
	var file config.FileInfo

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	results := make([]config.FileInfo, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		if ext := filepath.Ext(filename); ext != "" {
			continue
		}

		path := filepath.Join(dirPath, filename)
		fileinfo, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get file information for %s: %w", filename, err)
		}

		mode := fileinfo.Mode()
		isExecutable := mode&0111 != 0

		if !isExecutable {
			file.Path = path
			results = append(results, file)
		}
	}

	return results, nil
}
