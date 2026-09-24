package recon

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rny/scanner/internal/domain"
)

// PortProbe represents the scan status of a single TCP port.
type PortProbe struct {
	Port      int    `json:"port"`
	Service   string `json:"service"`
	Protocol  string `json:"protocol"`
	Open      bool   `json:"open"`
	LatencyMs int64  `json:"latencyMs"`
	Category  string `json:"category"` // "Web" | "Remote Access" | "Mail/DNS" | "Database"
	RiskNote  string `json:"riskNote,omitempty"`
}

// TLSInfo holds live certificate & cipher handshake telemetry.
type TLSInfo struct {
	Supported     bool     `json:"supported"`
	Version       string   `json:"version,omitempty"`
	CipherSuite   string   `json:"cipherSuite,omitempty"`
	Issuer        string   `json:"issuer,omitempty"`
	Subject       string   `json:"subject,omitempty"`
	ValidFrom     string   `json:"validFrom,omitempty"`
	ValidUntil    string   `json:"validUntil,omitempty"`
	DaysRemaining int      `json:"daysRemaining"`
	SANs          []string `json:"sans,omitempty"`
}

// SubdomainHit represents an active subdomain resolved via parallel DNS probing.
type SubdomainHit struct {
	Subdomain string   `json:"subdomain"`
	IPs       []string `json:"ips"`
}

// SecurityCheck represents a single HTTP or DNS security control audit item.
type SecurityCheck struct {
	Control string `json:"control"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
}

// ReconReport aggregates parallel port scanning, TLS handshake, subdomains, and security posture.
type ReconReport struct {
	Domain         string          `json:"domain"`
	TargetIP       string          `json:"targetIp"`
	OpenPortsCount int             `json:"openPortsCount"`
	PortsScanned   int             `json:"portsScanned"`
	Ports          []PortProbe     `json:"ports"`
	TLS            TLSInfo         `json:"tls"`
	Subdomains     []SubdomainHit  `json:"subdomains"`
	SecurityGrade  string          `json:"securityGrade"`
	SecurityScore  int             `json:"securityScore"`
	SecurityChecks []SecurityCheck `json:"securityChecks"`
	DurationMs     int64           `json:"durationMs"`
}

type portTarget struct {
	port     int
	service  string
	category string
	risk     string
}

var standardPorts = []portTarget{
	{21, "FTP", "Remote Access", "Cleartext file transfer port"},
	{22, "SSH", "Remote Access", "Encrypted shell administration"},
	{25, "SMTP", "Mail/DNS", "Mail transfer relay"},
	{53, "DNS", "Mail/DNS", "Domain Name System"},
	{80, "HTTP", "Web", "Standard web traffic"},
	{443, "HTTPS", "Web", "Encrypted TLS web traffic"},
	{3000, "Node / Dev HTTP", "Web", "Development web server"},
	{3306, "MySQL", "Database", "Database port exposed to public internet"},
	{5432, "PostgreSQL", "Database", "Database port exposed to public internet"},
	{6379, "Redis", "Database", "In-memory cache exposed to public internet"},
	{8000, "HTTP API", "Web", "Alternative application server"},
	{8080, "HTTP Proxy", "Web", "Alternative HTTP proxy / service"},
	{8443, "HTTPS Alt", "Web", "Alternative TLS web service"},
	{27017, "MongoDB", "Database", "NoSQL port exposed to public internet"},
}

var commonSubPrefixes = []string{
	"www", "api", "app", "dev", "staging", "docs", "mail", "status", "cdn", "auth", "admin", "blog",
}

// RunRecon executes a full parallel port scan, TLS inspection, subdomain enumeration, and security audit.
func RunRecon(ctx context.Context, rawTarget string) ReconReport {
	start := time.Now()
	clean := domain.CleanDomainName(rawTarget)

	report := ReconReport{
		Domain:       clean,
		PortsScanned: len(standardPorts),
	}

	// Resolve primary IP
	ips, _ := net.DefaultResolver.LookupHost(ctx, clean)
	if len(ips) > 0 {
		report.TargetIP = ips[0]
	}

	var (
		wg         sync.WaitGroup
		ports      []PortProbe
		tlsInfo    TLSInfo
		subdomains []SubdomainHit
		secChecks  []SecurityCheck
		secScore   int
		secGrade   string
	)

	wg.Add(4)

	// 1. Parallel Port Scanner (14 concurrent workers)
	go func() {
		defer wg.Done()
		ports = scanPortsParallel(ctx, clean)
	}()

	// 2. TLS Handshake & Certificate Inspector
	go func() {
		defer wg.Done()
		tlsInfo = inspectTLS(ctx, clean)
	}()

	// 3. Parallel Subdomain Enumerator
	go func() {
		defer wg.Done()
		subdomains = discoverSubdomainsParallel(ctx, clean)
	}()

	// 4. HTTP & DNS Security Posture Audit
	go func() {
		defer wg.Done()
		secChecks, secScore, secGrade = auditSecurityPosture(ctx, clean)
	}()

	wg.Wait()

	openCount := 0
	for _, p := range ports {
		if p.Open {
			openCount++
		}
	}

	report.OpenPortsCount = openCount
	report.Ports = ports
	report.TLS = tlsInfo
	report.Subdomains = subdomains
	report.SecurityChecks = secChecks
	report.SecurityScore = secScore
	report.SecurityGrade = secGrade
	report.DurationMs = time.Since(start).Milliseconds()

	return report
}

func scanPortsParallel(ctx context.Context, host string) []PortProbe {
	results := make([]PortProbe, len(standardPorts))
	var wg sync.WaitGroup

	for i, pt := range standardPorts {
		wg.Add(1)
		go func(idx int, target portTarget) {
			defer wg.Done()
			t0 := time.Now()
			dialer := net.Dialer{Timeout: 1250 * time.Millisecond}
			conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, target.port))
			latency := time.Since(t0).Milliseconds()
			open := err == nil
			if conn != nil {
				_ = conn.Close()
			}
			results[idx] = PortProbe{
				Port:      target.port,
				Service:   target.service,
				Protocol:  "TCP",
				Open:      open,
				LatencyMs: latency,
				Category:  target.category,
				RiskNote:  target.risk,
			}
		}(i, pt)
	}

	wg.Wait()

	// Sort open ports first, then by port number ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Open != results[j].Open {
			return results[i].Open
		}
		return results[i].Port < results[j].Port
	})

	return results
}

func inspectTLS(ctx context.Context, host string) TLSInfo {
	dialer := &net.Dialer{Timeout: 2500 * time.Millisecond}
	conn, err := tls.DialWithDialer(dialer, "tcp", host+":443", &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return TLSInfo{Supported: false}
	}
	defer conn.Close()

	state := conn.ConnectionState()
	info := TLSInfo{
		Supported:   true,
		Version:     tlsVersionName(state.Version),
		CipherSuite: tls.CipherSuiteName(state.CipherSuite),
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		if len(cert.Issuer.Organization) > 0 {
			info.Issuer = cert.Issuer.Organization[0]
		} else {
			info.Issuer = cert.Issuer.CommonName
		}
		info.Subject = cert.Subject.CommonName
		info.ValidFrom = cert.NotBefore.Format("2006-01-02")
		info.ValidUntil = cert.NotAfter.Format("2006-01-02")
		info.DaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)

		for i, san := range cert.DNSNames {
			if i >= 8 {
				break
			}
			info.SANs = append(info.SANs, san)
		}
	}

	return info
}

func discoverSubdomainsParallel(ctx context.Context, baseDomain string) []SubdomainHit {
	var (
		mu   sync.Mutex
		hits []SubdomainHit
		wg   sync.WaitGroup
	)

	for _, sub := range commonSubPrefixes {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()
			fqdn := prefix + "." + baseDomain
			dnsCtx, cancel := context.WithTimeout(ctx, 1400*time.Millisecond)
			defer cancel()

			ips, err := net.DefaultResolver.LookupHost(dnsCtx, fqdn)
			if err == nil && len(ips) > 0 {
				var cleanIPs []string
				for i, ip := range ips {
					if i >= 2 {
						break
					}
					cleanIPs = append(cleanIPs, ip)
				}
				mu.Lock()
				hits = append(hits, SubdomainHit{
					Subdomain: fqdn,
					IPs:       cleanIPs,
				})
				mu.Unlock()
			}
		}(sub)
	}

	wg.Wait()
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Subdomain < hits[j].Subdomain
	})
	return hits
}

func auditSecurityPosture(ctx context.Context, host string) ([]SecurityCheck, int, string) {
	var checks []SecurityCheck
	score := 0

	reqCtx, cancel := context.WithTimeout(ctx, 3200*time.Millisecond)
	defer cancel()

	client := &http.Client{Timeout: 3200 * time.Millisecond}
	req, _ := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://"+host, nil)
	if req != nil {
		req.Header.Set("User-Agent", "Scanner-Security-Auditor/1.0")
	}

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()

		hsts := resp.Header.Get("Strict-Transport-Security") != ""
		if hsts {
			score += 25
		}
		checks = append(checks, SecurityCheck{
			Control: "Strict-Transport-Security (HSTS)",
			Passed:  hsts,
			Detail:  ternaryStr(hsts, "Enforces encrypted HTTPS connections", "Missing HSTS header"),
		})

		csp := resp.Header.Get("Content-Security-Policy") != ""
		if csp {
			score += 20
		}
		checks = append(checks, SecurityCheck{
			Control: "Content-Security-Policy (CSP)",
			Passed:  csp,
			Detail:  ternaryStr(csp, "Active XSS & injection policy header", "No CSP header configured"),
		})

		xfo := resp.Header.Get("X-Frame-Options") != "" || strings.Contains(resp.Header.Get("Content-Security-Policy"), "frame-ancestors")
		if xfo {
			score += 15
		}
		checks = append(checks, SecurityCheck{
			Control: "Clickjacking Defense (X-Frame / Ancestors)",
			Passed:  xfo,
			Detail:  ternaryStr(xfo, "Frame embedding restricted", "Permits arbitrary iframe embedding"),
		})
	} else {
		checks = append(checks, SecurityCheck{
			Control: "HTTPS Web Listener",
			Passed:  false,
			Detail:  "No HTTPS web listener responded on port 443",
		})
	}

	// DNS SPF & DMARC checks
	txts, _ := net.DefaultResolver.LookupTXT(reqCtx, host)
	hasSPF := false
	for _, t := range txts {
		if strings.Contains(strings.ToLower(t), "v=spf1") {
			hasSPF = true
			break
		}
	}
	if hasSPF {
		score += 20
	}
	checks = append(checks, SecurityCheck{
		Control: "DNS Sender Policy Framework (SPF)",
		Passed:  hasSPF,
		Detail:  ternaryStr(hasSPF, "Authorized outbound mail servers declared", "No SPF TXT record found"),
	})

	dmarcTxts, _ := net.DefaultResolver.LookupTXT(reqCtx, "_dmarc."+host)
	hasDMARC := false
	for _, t := range dmarcTxts {
		if strings.Contains(strings.ToLower(t), "v=dmarc1") {
			hasDMARC = true
			break
		}
	}
	if hasDMARC {
		score += 20
	}
	checks = append(checks, SecurityCheck{
		Control: "DNS DMARC Anti-Spoofing Policy",
		Passed:  hasDMARC,
		Detail:  ternaryStr(hasDMARC, "Active _dmarc policy protecting domain spoofing", "No _dmarc record published"),
	})

	grade := "C"
	switch {
	case score >= 90:
		grade = "A+"
	case score >= 75:
		grade = "A"
	case score >= 60:
		grade = "B"
	case score >= 40:
		grade = "C"
	default:
		grade = "D"
	}

	return checks, score, grade
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	default:
		return "TLS 1.0"
	}
}

func ternaryStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
