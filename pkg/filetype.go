package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

type FileType struct {
	Extension   string
	MimeType    string
	Description string
}

func GetFileType(path string) (*FileType, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the first 512 bytes to check file signature
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Reset the buffer to the actual bytes read
	buffer = bytes.TrimRight(buffer, "\x00")

	switch {
	// Java KeyStore (JKS) file
	// JKS files start with a magic number 0xFEEDFEED (or in decimal: -17957139 as a 32-bit int)
	case len(buffer) >= 4 && binary.BigEndian.Uint32(buffer[:4]) == 0xFEEDFEED:
		return &FileType{
			Extension:   ".jks",
			MimeType:    "application/x-java-keystore",
			Description: "Java KeyStore (JKS)",
		}, nil

	// JCEKS (Java Cryptography Extension KeyStore)
	case len(buffer) >= 4 && binary.BigEndian.Uint32(buffer[:4]) == 0xCECECECE:
		return &FileType{
			Extension:   ".jceks",
			MimeType:    "application/x-java-keystore",
			Description: "Java Cryptography Extension KeyStore (JCEKS)",
		}, nil

	// PEM file - check for the standard header
	case bytes.HasPrefix(buffer, []byte("-----BEGIN ")):
		// Try to determine the specific PEM type
		headerStr := string(buffer[:100]) // Look at the first 100 bytes for the header
		pemType := "Certificate"

		if strings.Contains(headerStr, "CERTIFICATE") {
			pemType = "Certificate"
		} else if strings.Contains(headerStr, "PRIVATE KEY") {
			pemType = "Private Key"
		} else if strings.Contains(headerStr, "PUBLIC KEY") {
			pemType = "Public Key"
		} else if strings.Contains(headerStr, "CSR") || strings.Contains(headerStr, "CERTIFICATE REQUEST") {
			pemType = "Certificate Signing Request"
		}

		return &FileType{
			Extension:   ".pem",
			MimeType:    "application/x-pem-file",
			Description: fmt.Sprintf("PEM Encoded %s", pemType),
		}, nil

	// DER file - check for ASN.1 DER encoding signatures
	// Most DER files start with 0x30 (SEQUENCE) followed by a length byte
	case len(buffer) >= 2 && buffer[0] == 0x30:
		// Try to determine the type of DER object based on common OIDs and structures
		derType := "Unknown"

		// Certificate typically starts with 0x30 0x82 (sequence with 2-byte length)
		// followed by version, serial number, and signature algorithm
		if len(buffer) > 15 &&
			buffer[0] == 0x30 &&
			(buffer[1] == 0x82 || buffer[1] >= 0x80) &&
			bytes.Contains(buffer[:15], []byte{0x06, 0x03, 0x55, 0x04}) { // Contains an OID from X.500 directory
			derType = "Certificate"
		}

		// RSA private key typically starts with 0x30 0x82 followed by a version (usually 0x02 0x01 0x00)
		if len(buffer) > 10 &&
			buffer[0] == 0x30 &&
			buffer[1] >= 0x80 &&
			bytes.Contains(buffer[:10], []byte{0x02, 0x01, 0x00}) {
			derType = "Private Key"
		}

		// Public key typically has subjectPublicKeyInfo sequence
		if len(buffer) > 15 &&
			buffer[0] == 0x30 &&
			bytes.Contains(buffer[:15], []byte{0x06, 0x09, 0x2A, 0x86, 0x48}) { // Contains RSA OID
			derType = "Public Key"
		}

		return &FileType{
			Extension:   ".der",
			MimeType:    "application/x-x509-ca-cert",
			Description: fmt.Sprintf("DER Encoded %s", derType),
		}, nil

	// PKCS#12 / PFX files (often used for certificates with private keys)
	case len(buffer) >= 4 && buffer[0] == 0x30 && buffer[1] >= 0x80 &&
		bytes.Contains(buffer[:10], []byte{0x06, 0x09, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x0C}):
		return &FileType{
			Extension:   ".p12",
			MimeType:    "application/x-pkcs12",
			Description: "PKCS#12 / PFX Certificate Store",
		}, nil

	// Executable
	case bytes.HasPrefix(buffer, []byte{0x4D, 0x5A}): // Windows
		return &FileType{
			Extension:   ".exe",
			MimeType:    "application/x-msdownload",
			Description: "Windows Executable",
		}, nil
	case bytes.HasPrefix(buffer, []byte{0x7F, 0x45, 0x4C, 0x46}): // ELF (Linux)
		return &FileType{
			Extension:   "", // Could be any executable on Linux
			MimeType:    "application/x-executable",
			Description: "Linux Executable",
		}, nil

	// Text files - more difficult to detect reliably
	case isTextFile(buffer):
		return &FileType{
			Extension:   ".txt",
			MimeType:    "text/plain",
			Description: "Text File",
		}, nil

	default:
		return &FileType{
			Extension:   "",
			MimeType:    "application/octet-stream",
			Description: "Unknown File Type",
		}, nil
	}
}

func isTextFile(buffer []byte) bool {
	// Check if buffer contains mostly printable ASCII characters
	printableCount := 0
	for _, b := range buffer {
		if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
			printableCount++
		}
	}

	// If more than 90% of characters are printable, assume it's text
	return printableCount > len(buffer)*9/10
}
