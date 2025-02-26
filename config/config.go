package config

import "time"

var CertExtensions = []string{
	".pem",
	".der",
	".crt",
	".cer",
	".pkcs12",
	".jks",
	".bcfks",
}

// Certificate file information
type FileInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
}

// The files found for each extension
type ExtensionResult struct {
	Type  string     `json:"type"`
	Files []FileInfo `json:"files"`
}

// Tthe complete search results
type SearchResult struct {
	SearchPath string            `json:"search_path"`
	TotalFiles int               `json:"total_files"`
	Results    []ExtensionResult `json:"results"`
	SearchTime time.Time         `json:"search_time"`
}
