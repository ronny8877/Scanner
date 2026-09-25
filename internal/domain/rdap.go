package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DomainInquiry holds detailed registration, RDAP/WHOIS, DNS, and valuation info for a single domain.
type DomainInquiry struct {
	Domain             string     `json:"domain"`
	Available          bool       `json:"available"`
	StatusSummary      string     `json:"statusSummary"` // "Available" or "Registered"
	LiveSiteURL        string     `json:"liveSiteUrl"`
	WaybackCalendarURL string     `json:"waybackCalendarUrl"`
	RegisteredAt       string     `json:"registeredAt,omitempty"`
	UpdatedAt          string     `json:"updatedAt,omitempty"`
	ExpiresAt          string     `json:"expiresAt,omitempty"`
	DomainAge          string     `json:"domainAge,omitempty"`
	DaysToExpiry       int        `json:"daysToExpiry,omitempty"`
	Registrar          string     `json:"registrar,omitempty"`
	RegistrarIANA      string     `json:"registrarIana,omitempty"`
	RegistryHandle     string     `json:"registryHandle,omitempty"`
	DNSSEC             string     `json:"dnssec,omitempty"`
	StatusFlags        []string   `json:"statusFlags,omitempty"`
	Nameservers        []string   `json:"nameservers,omitempty"`
	DNS                DNSRecords `json:"dns"`
	Valuation          Valuation  `json:"valuation"`
	CheckedAt          string     `json:"checkedAt"`
	CheckLatencyMs     int64      `json:"checkLatencyMs"`
}

// DNSRecords stores resolved DNS records for a domain.
type DNSRecords struct {
	A     []string `json:"a,omitempty"`
	AAAA  []string `json:"aaaa,omitempty"`
	CNAME string   `json:"cname,omitempty"`
	PTR   []string `json:"ptr,omitempty"`
	MX    []string `json:"mx,omitempty"`
	NS    []string `json:"ns,omitempty"`
	TXT   []string `json:"txt,omitempty"`
	DMARC []string `json:"dmarc,omitempty"`
	SPF   string   `json:"spf,omitempty"`
}

type rdapResponse struct {
	LDHName     string       `json:"ldhName"`
	Handle      string       `json:"handle"`
	Status      []string     `json:"status"`
	Events      []rdapEvent  `json:"events"`
	Entities    []rdapEntity `json:"entities"`
	SecureDNS   *struct {
		DelegationSigned bool `json:"delegationSigned"`
	} `json:"secureDNS"`
	Nameservers []struct {
		LDHName string `json:"ldhName"`
	} `json:"nameservers"`
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

type rdapEntity struct {
	Roles      []string      `json:"roles"`
	VCardArray []interface{} `json:"vcardArray"`
	PublicIDs  []struct {
		Type       string `json:"type"`
		Identifier string `json:"identifier"`
	} `json:"publicIds"`
	Entities []rdapEntity `json:"entities"`
}

// InspectDomain performs a deep inquiry on a domain via parallel DNS resolution, RDAP, and Port-43 WHOIS lookup.
func InspectDomain(ctx context.Context, rawDomain string) DomainInquiry {
	start := time.Now()
	cleanDomain := CleanDomainName(rawDomain)

	var (
		dnsRecords DNSRecords
		rdapData   *rdapResponse
		rdapErr    error
		whoisRec   WhoisRecord
		whoisErr   error
		wg         sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		dnsRecords = resolveDNS(ctx, cleanDomain)
	}()
	go func() {
		defer wg.Done()
		rdapData, _, rdapErr = fetchRDAP(ctx, cleanDomain)
	}()
	go func() {
		defer wg.Done()
		whoisRec, whoisErr = QueryWhois(ctx, cleanDomain)
	}()
	wg.Wait()

	hasDNS := len(dnsRecords.NS) > 0 || len(dnsRecords.A) > 0 || len(dnsRecords.AAAA) > 0 || len(dnsRecords.MX) > 0
	isRegistered := hasDNS || (rdapErr == nil && rdapData != nil) || (whoisErr == nil && whoisRec.Registered)

	inquiry := DomainInquiry{
		Domain:             cleanDomain,
		Available:          !isRegistered,
		StatusSummary:      "Available",
		LiveSiteURL:        "https://" + cleanDomain,
		WaybackCalendarURL: fmt.Sprintf("https://web.archive.org/web/*/%s", cleanDomain),
		DNS:                dnsRecords,
		CheckedAt:          time.Now().Format(time.RFC3339),
		Nameservers:        dnsRecords.NS,
	}

	if rdapErr == nil && rdapData != nil {
		inquiry.Available = false
		inquiry.StatusSummary = "Registered"
		inquiry.RegistryHandle = rdapData.Handle
		inquiry.StatusFlags = rdapData.Status
		if rdapData.SecureDNS != nil {
			if rdapData.SecureDNS.DelegationSigned {
				inquiry.DNSSEC = "Signed (DNSSEC Active)"
			} else {
				inquiry.DNSSEC = "Unsigned"
			}
		}

		if len(rdapData.Nameservers) > 0 && len(inquiry.Nameservers) == 0 {
			for _, ns := range rdapData.Nameservers {
				inquiry.Nameservers = append(inquiry.Nameservers, strings.ToLower(ns.LDHName))
			}
		}

		for _, ev := range rdapData.Events {
			switch ev.EventAction {
			case "registration":
				inquiry.RegisteredAt = ev.EventDate
				inquiry.DomainAge = computeHumanAge(ev.EventDate)
			case "expiration":
				inquiry.ExpiresAt = ev.EventDate
				inquiry.DaysToExpiry = computeDaysUntil(ev.EventDate)
			case "last changed", "last update of RDAP database":
				if inquiry.UpdatedAt == "" {
					inquiry.UpdatedAt = ev.EventDate
				}
			}
		}

		inquiry.Registrar, inquiry.RegistrarIANA = extractRegistrar(rdapData.Entities)
	}

	// Merge authoritative Port-43 WHOIS fields (essential for .co, .ai, .io, .gg, .so, and parked/lame-NS domains like agent.co)
	if whoisErr == nil && whoisRec.Registered {
		inquiry.Available = false
		if inquiry.RegisteredAt == "" && whoisRec.CreatedDate != "" {
			inquiry.RegisteredAt = whoisRec.CreatedDate
			inquiry.DomainAge = computeHumanAge(whoisRec.CreatedDate)
		}
		if inquiry.UpdatedAt == "" && whoisRec.UpdatedDate != "" {
			inquiry.UpdatedAt = whoisRec.UpdatedDate
		}
		if inquiry.ExpiresAt == "" && whoisRec.ExpiryDate != "" {
			inquiry.ExpiresAt = whoisRec.ExpiryDate
			inquiry.DaysToExpiry = computeDaysUntil(whoisRec.ExpiryDate)
		}
		if inquiry.Registrar == "" && whoisRec.Registrar != "" {
			inquiry.Registrar = whoisRec.Registrar
		}
		if inquiry.RegistrarIANA == "" && whoisRec.RegistrarIANA != "" {
			inquiry.RegistrarIANA = whoisRec.RegistrarIANA
		}
		if len(inquiry.Nameservers) == 0 && len(whoisRec.Nameservers) > 0 {
			inquiry.Nameservers = whoisRec.Nameservers
			inquiry.DNS.NS = whoisRec.Nameservers
		}
		if len(inquiry.StatusFlags) == 0 && len(whoisRec.StatusFlags) > 0 {
			inquiry.StatusFlags = whoisRec.StatusFlags
		}
		if inquiry.DNSSEC == "" && whoisRec.DNSSEC != "" {
			inquiry.DNSSEC = whoisRec.DNSSEC
		}
	}

	if !inquiry.Available {
		hasWebIP := len(dnsRecords.A) > 0 || len(dnsRecords.AAAA) > 0
		if !hasWebIP {
			inquiry.StatusSummary = "Registered (Parked / No Active A Record)"
		} else {
			inquiry.StatusSummary = "Registered"
		}
		if inquiry.Registrar == "" {
			inquiry.Registrar = "Active Registry Delegation"
		}
	}

	inquiry.Valuation = EvaluateDomainWithStatus(cleanDomain, inquiry.Available)
	inquiry.CheckLatencyMs = time.Since(start).Milliseconds()
	return inquiry
}

func fetchRDAP(ctx context.Context, domain string) (*rdapResponse, int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	parts := strings.Split(domain, ".")
	tld := parts[len(parts)-1]

	var candidateURLs []string
	if directBase, ok := tldRDAPEndpoints[tld]; ok {
		candidateURLs = append(candidateURLs, directBase+domain)
	}
	candidateURLs = append(candidateURLs, fmt.Sprintf("https://rdap.org/domain/%s", domain))

	client := &http.Client{
		Timeout: 4000 * time.Millisecond,
	}

	for _, u := range candidateURLs {
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/rdap+json, application/json")
		req.Header.Set("User-Agent", "Scanner-Domain-Intelligence/2.0")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		var parsed rdapResponse
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		resp.Body.Close()
		if err == nil && parsed.LDHName != "" {
			return &parsed, http.StatusOK, nil
		}
	}

	return nil, http.StatusNotFound, fmt.Errorf("rdap not found")
}

func resolveDNS(ctx context.Context, domain string) DNSRecords {
	r := net.DefaultResolver
	dnsCtx, cancel := context.WithTimeout(ctx, 2600*time.Millisecond)
	defer cancel()

	var rec DNSRecords

	if ips, err := r.LookupIPAddr(dnsCtx, domain); err == nil {
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				rec.A = append(rec.A, ip.IP.String())
			} else {
				rec.AAAA = append(rec.AAAA, ip.IP.String())
			}
		}
	}

	// Reverse DNS PTR lookup on primary IPv4
	if len(rec.A) > 0 {
		if names, err := r.LookupAddr(dnsCtx, rec.A[0]); err == nil {
			for i, n := range names {
				if i >= 2 {
					break
				}
				rec.PTR = append(rec.PTR, strings.TrimSuffix(n, "."))
			}
		}
	}

	if cname, err := r.LookupCNAME(dnsCtx, "www."+domain); err == nil && cname != "" {
		cleanCname := strings.TrimSuffix(strings.ToLower(cname), ".")
		if cleanCname != "www."+domain {
			rec.CNAME = cleanCname
		}
	}

	if nss, err := r.LookupNS(dnsCtx, domain); err == nil {
		for _, ns := range nss {
			rec.NS = append(rec.NS, strings.TrimSuffix(strings.ToLower(ns.Host), "."))
		}
	}

	if mxs, err := r.LookupMX(dnsCtx, domain); err == nil {
		for _, mx := range mxs {
			rec.MX = append(rec.MX, fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(strings.ToLower(mx.Host), ".")))
		}
	}

	if txts, err := r.LookupTXT(dnsCtx, domain); err == nil {
		for i, t := range txts {
			if strings.Contains(strings.ToLower(t), "v=spf1") {
				rec.SPF = t
			}
			if i < 8 {
				rec.TXT = append(rec.TXT, t)
			}
		}
	}

	if dmarcs, err := r.LookupTXT(dnsCtx, "_dmarc."+domain); err == nil {
		for _, d := range dmarcs {
			rec.DMARC = append(rec.DMARC, d)
		}
	}

	return rec
}

func extractRegistrar(entities []rdapEntity) (name string, iana string) {
	for _, ent := range entities {
		isRegistrar := false
		for _, role := range ent.Roles {
			if role == "registrar" {
				isRegistrar = true
				break
			}
		}
		if isRegistrar {
			for _, pid := range ent.PublicIDs {
				if pid.Identifier != "" {
					iana = pid.Identifier
				}
			}
			if len(ent.VCardArray) >= 2 {
				if props, ok := ent.VCardArray[1].([]interface{}); ok {
					for _, p := range props {
						if row, ok := p.([]interface{}); ok && len(row) >= 4 {
							if key, ok := row[0].(string); ok && key == "fn" {
								if val, ok := row[3].(string); ok && val != "" {
									name = val
								}
							}
						}
					}
				}
			}
			if name != "" {
				if iana != "" {
					return fmt.Sprintf("%s (IANA #%s)", name, iana), iana
				}
				return name, iana
			}
		}
	}
	return "", ""
}

func parseRegistryDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	if len(raw) >= 10 {
		return time.Parse("2006-01-02", raw[:10])
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func computeHumanAge(isoDate string) string {
	t, err := parseRegistryDate(isoDate)
	if err != nil {
		return ""
	}
	days := int(time.Since(t).Hours() / 24)
	if days < 0 {
		return "0 days"
	}
	years := days / 365
	months := (days % 365) / 30
	if years > 0 {
		return fmt.Sprintf("%dy %dm (%d days)", years, months, days)
	}
	return fmt.Sprintf("%dm (%d days)", months, days)
}

func computeDaysUntil(isoDate string) int {
	t, err := parseRegistryDate(isoDate)
	if err != nil {
		return 0
	}
	return int(time.Until(t).Hours() / 24)
}

// CleanDomainName strips protocol, path, and port from a raw input string.
func CleanDomainName(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if idx := strings.IndexByte(s, '/'); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.IndexByte(s, ':'); idx != -1 {
		s = s[:idx]
	}
	if !strings.Contains(s, ".") && s != "" {
		s = s + ".com"
	}
	return s
}
