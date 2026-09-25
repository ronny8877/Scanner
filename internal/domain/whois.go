package domain

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// WhoisRecord holds parsed authoritative fields from a TCP Port-43 WHOIS query.
type WhoisRecord struct {
	Registered    bool
	Registrar     string
	RegistrarIANA string
	CreatedDate   string
	UpdatedDate   string
	ExpiryDate    string
	Nameservers   []string
	StatusFlags   []string
	DNSSEC        string
}

// tldWhoisServers maps our 24 supported TLDs to their authoritative port-43 WHOIS servers.
var tldWhoisServers = map[string]string{
	"com":     "whois.verisign-grs.com",
	"net":     "whois.verisign-grs.com",
	"cc":      "ccwhois.verisign-grs.com",
	"org":     "whois.pir.org",
	"ai":      "whois.nic.ai",
	"io":      "whois.nic.io",
	"co":      "whois.registry.co",
	"dev":     "whois.nic.google",
	"app":     "whois.nic.google",
	"xyz":     "whois.nic.xyz",
	"sh":      "whois.nic.sh",
	"me":      "whois.nic.me",
	"gg":      "whois.gg",
	"so":      "whois.nic.so",
	"vc":      "whois.nic.vc",
	"tech":    "whois.nic.tech",
	"cloud":   "whois.nic.cloud",
	"studio":  "whois.nic.studio",
	"design":  "whois.nic.design",
	"tools":   "whois.nic.tools",
	"run":     "whois.nic.run",
	"build":   "whois.nic.build",
	"finance": "whois.nic.finance",
	"systems": "whois.nic.systems",
}

// tldRDAPEndpoints maps TLDs that rdap.org may miss (such as .co, .ai, .io, .sh, .me, .gg) to their direct registry RDAP base URLs.
var tldRDAPEndpoints = map[string]string{
	"co":      "https://rdap.centralnic.com/co/domain/",
	"ai":      "https://rdap.identitydigital.services/rdap/domain/",
	"io":      "https://rdap.identitydigital.services/rdap/domain/",
	"me":      "https://rdap.identitydigital.services/rdap/domain/",
	"sh":      "https://rdap.identitydigital.services/rdap/domain/",
	"vc":      "https://rdap.identitydigital.services/rdap/domain/",
	"studio":  "https://rdap.identitydigital.services/rdap/domain/",
	"tools":   "https://rdap.identitydigital.services/rdap/domain/",
	"run":     "https://rdap.identitydigital.services/rdap/domain/",
	"finance": "https://rdap.identitydigital.services/rdap/domain/",
	"systems": "https://rdap.identitydigital.services/rdap/domain/",
	"xyz":     "https://rdap.centralnic.com/xyz/domain/",
	"tech":    "https://rdap.centralnic.com/tech/domain/",
	"design":  "https://rdap.centralnic.com/design/domain/",
	"build":   "https://rdap.centralnic.com/build/domain/",
	"dev":     "https://pubapi.registry.google/rdap/domain/",
	"app":     "https://pubapi.registry.google/rdap/domain/",
}

// QueryWhois performs an authoritative port-43 WHOIS lookup for a domain.
func QueryWhois(ctx context.Context, fullDomain string) (WhoisRecord, error) {
	var rec WhoisRecord
	clean := CleanDomainName(fullDomain)
	parts := strings.Split(clean, ".")
	if len(parts) < 2 {
		return rec, fmt.Errorf("invalid domain: %s", fullDomain)
	}
	tld := parts[len(parts)-1]

	server, ok := tldWhoisServers[tld]
	if !ok {
		server = "whois.iana.org"
	}

	raw, err := dialWhoisServer(ctx, server, clean)
	if err != nil {
		return rec, err
	}

	// If we queried whois.iana.org and received a "refer:" line, follow the referral server
	if server == "whois.iana.org" {
		for _, line := range strings.Split(raw, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "refer:") {
				refParts := strings.SplitN(line, ":", 2)
				if len(refParts) == 2 {
					refSrv := strings.TrimSpace(refParts[1])
					if refSrv != "" {
						if refRaw, err := dialWhoisServer(ctx, refSrv, clean); err == nil {
							raw = refRaw
						}
					}
				}
				break
			}
		}
	}

	return parseWhoisOutput(clean, raw), nil
}

func dialWhoisServer(ctx context.Context, server, query string) (string, error) {
	dialer := net.Dialer{Timeout: 2500 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(server, "43"))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(2800 * time.Millisecond))
	if _, err := fmt.Fprintf(conn, "%s\r\n", query); err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(conn, 64*1024))
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func parseWhoisOutput(domain, raw string) WhoisRecord {
	var rec WhoisRecord
	lowerRaw := strings.ToLower(raw)

	notFoundMarkers := []string{
		"no match for",
		"not found",
		"no data found",
		"no entries found",
		"domain not found",
		"status: free",
		"status: available",
		"is available for registration",
		"no object found",
	}
	for _, marker := range notFoundMarkers {
		if strings.Contains(lowerRaw, marker) {
			rec.Registered = false
			return rec
		}
	}

	nsSeen := make(map[string]bool)
	statusSeen := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ">>>") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if val == "" {
			continue
		}

		switch key {
		case "domain name", "domain":
			if strings.EqualFold(val, domain) {
				rec.Registered = true
			}
		case "registrar", "sponsoring registrar":
			if rec.Registrar == "" {
				rec.Registrar = val
				rec.Registered = true
			}
		case "registrar iana id":
			if rec.RegistrarIANA == "" {
				rec.RegistrarIANA = val
			}
		case "creation date", "created on", "created", "registration time", "registered on":
			if rec.CreatedDate == "" {
				rec.CreatedDate = normalizeWhoisDate(val)
				rec.Registered = true
			}
		case "updated date", "last updated", "modified":
			if rec.UpdatedDate == "" {
				rec.UpdatedDate = normalizeWhoisDate(val)
			}
		case "registry expiry date", "registrar registration expiration date", "expiry date", "expires on", "paid-till":
			if rec.ExpiryDate == "" {
				rec.ExpiryDate = normalizeWhoisDate(val)
				rec.Registered = true
			}
		case "name server", "nserver", "nameserver":
			ns := strings.TrimSuffix(strings.ToLower(strings.Fields(val)[0]), ".")
			if ns != "" && !nsSeen[ns] {
				nsSeen[ns] = true
				rec.Nameservers = append(rec.Nameservers, ns)
				rec.Registered = true
			}
		case "domain status", "status":
			st := strings.Fields(val)[0]
			if st != "" && !statusSeen[st] {
				statusSeen[st] = true
				rec.StatusFlags = append(rec.StatusFlags, st)
			}
		case "dnssec":
			if rec.DNSSEC == "" {
				rec.DNSSEC = val
			}
		}
	}

	return rec
}

func normalizeWhoisDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 10 && raw[4] == '-' && raw[7] == '-' {
		return raw[:10]
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02-Jan-2006",
		"2006/01/02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return raw
}
