package cmd

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"
)

type FIPSResult struct {
	IsCompliant bool
	Reasons     []string
}

// Is the provided X.509 certificate FIPS 140-3 compliant?
func IsFIPSCompliant(certPath string) (*FIPSResult, error) {
	// Read certificate file
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	result := &FIPSResult{
		IsCompliant: true,
		Reasons:     []string{},
	}

	// Check signature algorithm
	if !isFIPSSignatureAlgorithm(cert.SignatureAlgorithm) {
		result.IsCompliant = false
		result.Reasons = append(result.Reasons,
			fmt.Sprintf("Signature algorithm %v is not FIPS 140-3 compliant", cert.SignatureAlgorithm))
	}

	// Check public key algorithm and key size
	if !isFIPSCompliantPublicKey(cert.PublicKey) {
		result.IsCompliant = false
		result.Reasons = append(result.Reasons, "Public key type or size is not FIPS 140-3 compliant")
	}

	// Check certificate expiration
	if !hasExpired(cert) {
		result.IsCompliant = false
		result.Reasons = append(result.Reasons, "Certificate is expired or not yet valid")
	}

	return result, nil
}

// Is the signature algorithm is FIPS 140-3 compliant?
func isFIPSSignatureAlgorithm(sigAlg x509.SignatureAlgorithm) bool {
	// FIPS 140-3 compliant signature algorithms
	compliantAlgorithms := map[x509.SignatureAlgorithm]bool{
		x509.SHA256WithRSA:   true,
		x509.SHA384WithRSA:   true,
		x509.SHA512WithRSA:   true,
		x509.ECDSAWithSHA256: true,
		x509.ECDSAWithSHA384: true,
		x509.ECDSAWithSHA512: true,
	}

	return compliantAlgorithms[sigAlg]
}

// I the public key type and size FIPS 140-3 compliant?
func isFIPSCompliantPublicKey(pubKey interface{}) bool {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		// RSA keys must be at least 2048 bits
		return key.N.BitLen() >= 2048
	case *ecdsa.PublicKey:
		// Check for approved curves (P-256, P-384, P-521)
		curve := key.Curve.Params().Name
		return curve == "P-256" || curve == "P-384" || curve == "P-521"
	case *dsa.PublicKey:
		// DSA is not approved for FIPS 140-3
		return false
	default:
		// Unknown key type
		return false
	}
}

// Has the certificate is expired?
func hasExpired(cert *x509.Certificate) bool {
	now := time.Now()
	return now.After(cert.NotBefore) && now.Before(cert.NotAfter)
}

func GetCertificateExpirationInfo(cert *x509.Certificate) string {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		daysUntilValid := int(cert.NotBefore.Sub(now).Hours() / 24)
		return fmt.Sprintf("Certificate is not yet valid. Will become valid in %d days (on %s)",
			daysUntilValid, cert.NotBefore.Format("Jan 2, 2006"))
	} else if now.After(cert.NotAfter) {
		daysExpired := int(now.Sub(cert.NotAfter).Hours() / 24)
		return fmt.Sprintf("Certificate has expired %d days ago (on %s)",
			daysExpired, cert.NotAfter.Format("Jan 2, 2006"))
	} else {
		daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)
		return fmt.Sprintf("Certificate is currently valid. Expires in %d days (on %s)",
			daysRemaining, cert.NotAfter.Format("Jan 2, 2006"))
	}
}

// Prints the FIPS compliance check result
func PrintFIPSResult(result *FIPSResult, cert *x509.Certificate) {
	if result.IsCompliant {
		fmt.Println("Certificate is FIPS 140-3 compliant.")
	} else {
		fmt.Println("Certificate is NOT FIPS 140-3 compliant for the following reasons:")
		for _, reason := range result.Reasons {
			fmt.Printf("- %s\n", reason)
		}
	}

	// Print expiration information
	fmt.Println("\nExpiration Information:")
	fmt.Println(GetCertificateExpirationInfo(cert))
}
