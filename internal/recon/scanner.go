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

	"io"

	"github.com/rny/scanner/internal/domain"
	"github.com/rny/scanner/internal/webintel"
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
	Subdomain    string   `json:"subdomain"`
	IPs          []string `json:"ips"`
	CNAME        string   `json:"cname,omitempty"`
	TakeoverRisk string   `json:"takeoverRisk,omitempty"`
}

// SecurityCheck represents a single HTTP or DNS security control audit item.
type SecurityCheck struct {
	Control string `json:"control"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
}

// ReconReport aggregates parallel port scanning, TLS handshake, subdomains, headers, security posture, and ad/tracker telemetry.
type ReconReport struct {
	Domain             string                    `json:"domain"`
	LiveSiteURL        string                    `json:"liveSiteUrl"`
	WaybackCalendarURL string                    `json:"waybackCalendarUrl"`
	TargetIP           string                    `json:"targetIp"`
	ReversePTR         string                    `json:"reversePtr,omitempty"`
	OpenPortsCount     int                       `json:"openPortsCount"`
	PortsScanned       int                       `json:"portsScanned"`
	Ports              []PortProbe               `json:"ports"`
	TLS                TLSInfo                   `json:"tls"`
	Subdomains         []SubdomainHit            `json:"subdomains"`
	HTTPHeaders        map[string]string         `json:"httpHeaders,omitempty"`
	HasRobotsTxt       bool                      `json:"hasRobotsTxt"`
	HasSecurityTxt     bool                      `json:"hasSecurityTxt"`
	HasSitemapXml      bool                      `json:"hasSitemapXml"`
	SecurityGrade      string                    `json:"securityGrade"`
	SecurityScore      int                       `json:"securityScore"`
	SecurityChecks     []SecurityCheck           `json:"securityChecks"`
	Trackers           webintel.TrackerTelemetry `json:"trackers"`
	DurationMs         int64                     `json:"durationMs"`
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
	"www", "api", "app", "dev", "staging", "docs", "mail", "status", "cdn", "auth", "admin", "blog", "portal", "beta",
}

// RunRecon executes a full parallel port scan, TLS inspection, subdomain enumeration, and security audit.
func RunRecon(ctx context.Context, rawTarget string) ReconReport {
	start := time.Now()
	clean := domain.CleanDomainName(rawTarget)

	report := ReconReport{
		Domain:             clean,
		LiveSiteURL:        "https://" + clean,
		WaybackCalendarURL: fmt.Sprintf("https://web.archive.org/web/*/%s", clean),
		PortsScanned:       len(standardPorts),
		HTTPHeaders:        make(map[string]string),
	}

	// Resolve primary IP & PTR
	ips, _ := net.DefaultResolver.LookupHost(ctx, clean)
	if len(ips) > 0 {
		report.TargetIP = ips[0]
		if ptrs, err := net.DefaultResolver.LookupAddr(ctx, ips[0]); err == nil && len(ptrs) > 0 {
			report.ReversePTR = strings.TrimSuffix(ptrs[0], ".")
		}
	}

	var (
		wg             sync.WaitGroup
		ports          []PortProbe
		tlsInfo        TLSInfo
		subdomains     []SubdomainHit
		secChecks      []SecurityCheck
		secScore       int
		secGrade       string
		headersMap     map[string]string
		hasRobots      bool
		hasSecurityTxt bool
		hasSitemap     bool
		trackers       webintel.TrackerTelemetry
	)

	wg.Add(4)

	// 1. Parallel Port Scanner
	go func() {
		defer wg.Done()
		ports = scanPortsParallel(ctx, clean)
	}()

	// 2. TLS Handshake & Certificate Inspector
	go func() {
		defer wg.Done()
		tlsInfo = inspectTLS(ctx, clean)
	}()

	// 3. Parallel Subdomain Enumerator + Takeover Check
	go func() {
		defer wg.Done()
		subdomains = discoverSubdomainsParallel(ctx, clean)
	}()

	// 4. HTTP Security Headers + Governance Files + SPF/DMARC + Ad/Tracker Audit
	go func() {
		defer wg.Done()
		secChecks, secScore, secGrade, headersMap, hasRobots, hasSecurityTxt, hasSitemap, trackers = auditSecurityPosture(ctx, clean)
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
	report.HTTPHeaders = headersMap
	report.HasRobotsTxt = hasRobots
	report.HasSecurityTxt = hasSecurityTxt
	report.HasSitemapXml = hasSitemap
	report.SecurityChecks = secChecks
	report.SecurityScore = secScore
	report.SecurityGrade = secGrade
	report.Trackers = trackers
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
			if i >= 10 {
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

	takeoverProviders := []string{
		"github.io", "s3.amazonaws.com", "herokuapp.com", "azurewebsites.net",
		"pantheonsite.io", "ghost.io", "vercel-dns.com", " cname.vercel-dns.com",
	}

	// Step 1: Detect Wildcard DNS (*.baseDomain) using random canary hostnames
	wildcardIPs := make(map[string]bool)
	var wildcardCNAME string
	canaryCtx, cancelCanary := context.WithTimeout(ctx, 1200*time.Millisecond)
	for _, canary := range []string{"_wildcard-canary-9821x." + baseDomain, "_nx-probe-4719z." + baseDomain} {
		if cn, err := net.DefaultResolver.LookupCNAME(canaryCtx, canary); err == nil {
			cleanCn := strings.TrimSuffix(strings.ToLower(cn), ".")
			if cleanCn != canary {
				wildcardCNAME = cleanCn
			}
		}
		if ips, err := net.DefaultResolver.LookupHost(canaryCtx, canary); err == nil {
			for _, ip := range ips {
				wildcardIPs[ip] = true
			}
		}
	}
	cancelCanary()
	hasWildcardDNS := len(wildcardIPs) > 0

	httpClient := &http.Client{
		Timeout: 1400 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirects blindly to root domain
		},
	}

	for _, sub := range commonSubPrefixes {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()
			fqdn := prefix + "." + baseDomain
			dnsCtx, cancel := context.WithTimeout(ctx, 1400*time.Millisecond)
			defer cancel()

			var cnameStr string
			if cn, err := net.DefaultResolver.LookupCNAME(dnsCtx, fqdn); err == nil {
				cleanCn := strings.TrimSuffix(strings.ToLower(cn), ".")
				if cleanCn != fqdn {
					cnameStr = cleanCn
				}
			}

			ips, err := net.DefaultResolver.LookupHost(dnsCtx, fqdn)
			if err != nil || len(ips) == 0 {
				// Check if it has a CNAME that failed to resolve IP (true dangling CNAME takeover candidate!)
				if cnameStr != "" {
					for _, prov := range takeoverProviders {
						if strings.Contains(cnameStr, prov) {
							mu.Lock()
							hits = append(hits, SubdomainHit{
								Subdomain:    fqdn,
								IPs:          []string{},
								CNAME:        cnameStr,
								TakeoverRisk: "HIGH RISK: Dangling CNAME (" + prov + " — NXDOMAIN)",
							})
							mu.Unlock()
							return
						}
					}
				}
				return
			}

			// Step 2: Check if all resolved IPs match the Wildcard DNS canary
			allMatchWildcard := hasWildcardDNS
			var cleanIPs []string
			for i, ip := range ips {
				if !wildcardIPs[ip] {
					allMatchWildcard = false
				}
				if i < 2 {
					cleanIPs = append(cleanIPs, ip)
				}
			}

			// If wildcard DNS is active and this subdomain has no distinct IP or distinct CNAME,
			// only keep "www" if it responds cleanly; filter out phantom subdomains (admin, grafana, etc.)
			if allMatchWildcard && (cnameStr == "" || cnameStr == wildcardCNAME) {
				if prefix != "www" {
					return
				}
			}

			// Step 3: Verify live service responsiveness & filter out 404 / deployment-not-found ghosts
			if prefix != "mail" && prefix != "smtp" {
				probeReq, reqErr := http.NewRequestWithContext(dnsCtx, http.MethodGet, "https://"+fqdn, nil)
				if reqErr == nil {
					probeReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
					resp, httpErr := httpClient.Do(probeReq)
					if httpErr != nil {
						// Try HTTP port 80 before discarding
						probeReq80, _ := http.NewRequestWithContext(dnsCtx, http.MethodGet, "http://"+fqdn, nil)
						resp, httpErr = httpClient.Do(probeReq80)
					}
					if httpErr != nil {
						// Does not respond to HTTPS or HTTP at all — skip non-responsive subdomain
						return
					}
					statusCode := resp.StatusCode
					vercelError := resp.Header.Get("x-vercel-error")
					resp.Body.Close()

					// Discard 404 Not Found, 502/504 Bad Gateway, 530 Cloudflare DNS Error, or Vercel DEPLOYMENT_NOT_FOUND
					if statusCode == 404 || statusCode == 410 || statusCode == 502 || statusCode == 530 || vercelError != "" {
						return
					}
				}
			}

			risk := "Verified Active Endpoint"
			for _, prov := range takeoverProviders {
				if strings.Contains(cnameStr, prov) {
					risk = "External Cloud CNAME (" + prov + ")"
				}
			}

			mu.Lock()
			hits = append(hits, SubdomainHit{
				Subdomain:    fqdn,
				IPs:          cleanIPs,
				CNAME:        cnameStr,
				TakeoverRisk: risk,
			})
			mu.Unlock()
		}(sub)
	}

	wg.Wait()
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Subdomain < hits[j].Subdomain
	})
	return hits
}

func auditSecurityPosture(ctx context.Context, host string) (
	checks []SecurityCheck,
	score int,
	grade string,
	headers map[string]string,
	hasRobots, hasSecurityTxt, hasSitemap bool,
	trackers webintel.TrackerTelemetry,
) {
	headers = make(map[string]string)

	reqCtx, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	client := &http.Client{Timeout: 4000 * time.Millisecond}
	req, _ := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://"+host, nil)
	if req != nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Scanner-Security-Auditor/2.0)")
	}

	var htmlSample string
	var respHeader http.Header

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		respHeader = resp.Header
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 300*1024))
		htmlSample = string(bodyBytes)

		interestingHeaders := []string{
			"Server", "X-Powered-By", "CF-Ray", "Cache-Control",
			"Strict-Transport-Security", "Content-Security-Policy",
			"X-Frame-Options", "X-Content-Type-Options", "Referrer-Policy", "Permissions-Policy",
		}
		for _, hk := range interestingHeaders {
			if val := resp.Header.Get(hk); val != "" {
				if len(val) > 85 {
					val = val[:82] + "..."
				}
				headers[hk] = val
			}
		}

		hsts := resp.Header.Get("Strict-Transport-Security") != ""
		if hsts {
			score += 20
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
			Detail:  ternaryStr(csp, "Active XSS & script policy header", "No CSP header configured"),
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

		xcto := strings.EqualFold(resp.Header.Get("X-Content-Type-Options"), "nosniff")
		if xcto {
			score += 10
		}
		checks = append(checks, SecurityCheck{
			Control: "MIME-Sniffing Protection (X-Content-Type-Options)",
			Passed:  xcto,
			Detail:  ternaryStr(xcto, "nosniff header active", "Missing nosniff protection"),
		})
	} else {
		checks = append(checks, SecurityCheck{
			Control: "HTTPS Web Listener",
			Passed:  false,
			Detail:  "No HTTPS web listener responded on port 443",
		})
	}

	// Probe robots.txt, security.txt, sitemap.xml in parallel
	var probeWg sync.WaitGroup
	probeWg.Add(3)
	go func() {
		defer probeWg.Done()
		hasRobots = checkEndpointStatus(reqCtx, client, "https://"+host+"/robots.txt")
	}()
	go func() {
		defer probeWg.Done()
		hasSecurityTxt = checkEndpointStatus(reqCtx, client, "https://"+host+"/.well-known/security.txt")
	}()
	go func() {
		defer probeWg.Done()
		hasSitemap = checkEndpointStatus(reqCtx, client, "https://"+host+"/sitemap.xml")
	}()
	probeWg.Wait()

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
		score += 18
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
		score += 17
	}
	checks = append(checks, SecurityCheck{
		Control: "DNS DMARC Anti-Spoofing Policy",
		Passed:  hasDMARC,
		Detail:  ternaryStr(hasDMARC, "Active _dmarc policy protecting domain spoofing", "No _dmarc record published"),
	})

	switch {
	case score >= 88:
		grade = "A+"
	case score >= 74:
		grade = "A"
	case score >= 58:
		grade = "B"
	case score >= 40:
		grade = "C"
	default:
		grade = "D"
	}

	trackers = webintel.DetectTrackersFromHTML([]string{htmlSample}, respHeader)
	return checks, score, grade, headers, hasRobots, hasSecurityTxt, hasSitemap, trackers
}

func checkEndpointStatus(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
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
