package email

import (
	"fmt"
	"net"
	"strings"
)

// DomainInfo holds the verification results
type DomainInfo struct {
	Domain      string
	HasMX       bool
	HasSPF      bool
	SpfRecord   string
	HasDMARC    bool
	DmarcRecord string
}

// ValidateDomain attempts to verify that the domain has MX, SPF, and DMARC
// This is a basic approach with no caching and might be slow in real usage.
func ValidateDomain(domain string) (*DomainInfo, error) {
	info := &DomainInfo{Domain: domain}

	// 1. Check MX
	mxRecords, err := net.LookupMX(domain)
	if err == nil && len(mxRecords) > 0 {
		info.HasMX = true
	}

	// 2. Check SPF by scanning TXT for "v=spf1"
	txtRecords, _ := net.LookupTXT(domain)
	for _, txt := range txtRecords {
		if strings.HasPrefix(txt, "v=spf1") {
			info.HasSPF = true
			info.SpfRecord = txt
			break
		}
	}

	// 3. Check DMARC by scanning _dmarc.domain
	dmarcRecords, _ := net.LookupTXT(fmt.Sprintf("_dmarc.%s", domain))
	for _, txt := range dmarcRecords {
		if strings.HasPrefix(txt, "v=DMARC1") {
			info.HasDMARC = true
			info.DmarcRecord = txt
			break
		}
	}

	return info, nil
}

// IsDomainValid returns an error if the domain is missing MX/SPF/DMARC
func IsDomainValid(domain string) error {
	info, err := ValidateDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to validate domain: %w", err)
	}

	// For demonstration, we require ALL (MX, SPF, DMARC).
	// Adjust if you want different logic.
	if !info.HasMX || !info.HasSPF || !info.HasDMARC {
		return fmt.Errorf("domain %s missing required records (MX=%t, SPF=%t, DMARC=%t)",
			domain, info.HasMX, info.HasSPF, info.HasDMARC)
	}

	return nil
}

// ExtractDomain from an email. e.g., "someone@example.com" -> "example.com"
func ExtractDomain(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format")
	}
	return parts[1], nil
}
