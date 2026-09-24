package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// DomainInquiry holds detailed registration, RDAP/WHOIS, DNS, and valuation info for a single domain.
type DomainInquiry struct {
	Domain         string      `json:"domain"`
	Available      bool        `json:"available"`
	StatusSummary  string      `json:"statusSummary"` // "Available" or "Registered"
	RegisteredAt   string      `json:"registeredAt,omitempty"`
	UpdatedAt      string      `json:"updatedAt,omitempty"`
	ExpiresAt      string      `json:"expiresAt,omitempty"`
	DomainAge      string      `json:"domainAge,omitempty"`
	DaysToExpiry   int         `json:"daysToExpiry,omitempty"`
	Registrar      string      `json:"registrar,omitempty"`
	RegistryHandle string      `json:"registryHandle,omitempty"`
	StatusFlags    []string    `json:"statusFlags,omitempty"`
	Nameservers    []string    `json:"nameservers,omitempty"`
	DNS            DNSRecords  `json:"dns"`
	Valuation      Valuation   `json:"valuation"`
	CheckedAt      string      `json:"checkedAt"`
	CheckLatencyMs int64       `json:"checkLatencyMs"`
}

// DNSRecords stores resolved DNS records for a domain.
type DNSRecords struct {
	A     []string `json:"a,omitempty"`
	AAAA  []string `json:"aaaa,omitempty"`
	MX    []string `json:"mx,omitempty"`
	NS    []string `json:"ns,omitempty"`
	TXT   []string `json:"txt,omitempty"`
	CNAME string   `json:"cname,omitempty"`
}

type rdapResponse struct {
	LDHName     string       `json:"ldhName"`
	Handle      string       `json:"handle"`
	Status      []string     `json:"status"`
	Events      []rdapEvent  `json:"events"`
	Entities    []rdapEntity `json:"entities"`
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
	Entities   []rdapEntity  `json:"entities"`
}

// InspectDomain performs a deep inquiry on a domain via DNS resolution and RDAP lookup.
func InspectDomain(ctx context.Context, rawDomain string) DomainInquiry {
	start := time.Now()
	cleanDomain := CleanDomainName(rawDomain)
	val := EvaluateDomain(cleanDomain)

	dnsRecords := resolveDNS(ctx, cleanDomain)
	hasDNS := len(dnsRecords.NS) > 0 || len(dnsRecords.A) > 0 || len(dnsRecords.AAAA) > 0 || len(dnsRecords.MX) > 0

	inquiry := DomainInquiry{
		Domain:         cleanDomain,
		Available:      !hasDNS,
		StatusSummary:  "Available",
		DNS:            dnsRecords,
		Valuation:      val,
		CheckedAt:      time.Now().Format(time.RFC3339),
		Nameservers:    dnsRecords.NS,
	}

	// Attempt RDAP query for authoritative registration dates & registrar details
	rdapData, rdapStatus, err := fetchRDAP(ctx, cleanDomain)
	if err == nil && rdapData != nil {
		inquiry.Available = false
		inquiry.StatusSummary = "Registered"
		inquiry.RegistryHandle = rdapData.Handle
		inquiry.StatusFlags = rdapData.Status

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

		inquiry.Registrar = extractRegistrar(rdapData.Entities)
	} else if rdapStatus == http.StatusNotFound && !hasDNS {
		inquiry.Available = true
		inquiry.StatusSummary = "Available"
	} else if hasDNS {
		inquiry.Available = false
		inquiry.StatusSummary = "Registered"
		if inquiry.Registrar == "" {
			inquiry.Registrar = "Active DNS Delegation (RDAP rate-limited or restricted)"
		}
	}

	inquiry.CheckLatencyMs = time.Since(start).Milliseconds()
	return inquiry
}

func fetchRDAP(ctx context.Context, domain string) (*rdapResponse, int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	// rdap.org automatically redirects to the authoritative registry RDAP endpoint (Verisign, Google, Identity Digital, etc.)
	url := fmt.Sprintf("https://rdap.org/domain/%s", domain)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")
	req.Header.Set("User-Agent", "Scanner-Domain-Intelligence/1.0")

	client := &http.Client{
		Timeout: 4500 * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("rdap status %d", resp.StatusCode)
	}

	var parsed rdapResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, resp.StatusCode, err
	}
	return &parsed, resp.StatusCode, nil
}

func resolveDNS(ctx context.Context, domain string) DNSRecords {
	r := net.DefaultResolver
	dnsCtx, cancel := context.WithTimeout(ctx, 2200*time.Millisecond)
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
			if i >= 5 {
				break
			}
			rec.TXT = append(rec.TXT, t)
		}
	}

	return rec
}

func extractRegistrar(entities []rdapEntity) string {
	for _, ent := range entities {
		isRegistrar := false
		for _, role := range ent.Roles {
			if role == "registrar" {
				isRegistrar = true
				break
			}
		}
		if isRegistrar && len(ent.VCardArray) >= 2 {
			if props, ok := ent.VCardArray[1].([]interface{}); ok {
				for _, p := range props {
					if row, ok := p.([]interface{}); ok && len(row) >= 4 {
						if key, ok := row[0].(string); ok && key == "fn" {
							if val, ok := row[3].(string); ok && val != "" {
								return val
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func computeHumanAge(isoDate string) string {
	t, err := time.Parse(time.RFC3339, isoDate)
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
	t, err := time.Parse(time.RFC3339, isoDate)
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
