package domain

import (
	"context"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScanOptions configures a bulk domain availability & valuation scan.
type ScanOptions struct {
	Keywords       []string `json:"keywords"`
	TLDs           []string `json:"tlds"`
	DictionaryPack string   `json:"dictionaryPack,omitempty"`
	Mutations      bool     `json:"mutations"`
	OnlyAvailable  bool     `json:"onlyAvailable"`
	MinScore       int      `json:"minScore"`
	Concurrency    int      `json:"concurrency"`
	MaxResults     int      `json:"maxResults"`
}

// ScanResultItem holds availability, registry metadata, and valuation for a single domain candidate.
type ScanResultItem struct {
	Domain       string    `json:"domain"`
	RootName     string    `json:"rootName"`
	TLD          string    `json:"tld"`
	Available    bool      `json:"available"`
	Status       string    `json:"status"` // "Available" | "Registered"
	RegisteredAt string    `json:"registeredAt,omitempty"`
	Registrar    string    `json:"registrar,omitempty"`
	Nameservers  []string  `json:"nameservers,omitempty"`
	Valuation    Valuation `json:"valuation"`
	LatencyMs    int64     `json:"latencyMs"`
}

// ScanReport aggregates the full parallel scan execution report.
type ScanReport struct {
	SeedKeywords   []string         `json:"seedKeywords"`
	DictionaryUsed string           `json:"dictionaryUsed,omitempty"`
	TotalChecked   int              `json:"totalChecked"`
	AvailableCount int              `json:"availableCount"`
	TakenCount     int              `json:"takenCount"`
	HighValueCount int              `json:"highValueCount"`
	DurationMs     int64            `json:"durationMs"`
	Items          []ScanResultItem `json:"items"`
}

var (
	brandPrefixes = []string{"get", "use", "try", "go", "open", "my", "hey"}
	brandSuffixes = []string{"hq", "labs", "studio", "ai", "cloud", "app", "flow", "hub", "ops", "dev"}
)

// ScanDomains generates domain candidates (from seed keywords and/or dictionary packs) and checks availability in parallel.
func ScanDomains(ctx context.Context, opts ScanOptions) ScanReport {
	start := time.Now()
	if len(opts.TLDs) == 0 {
		opts.TLDs = []string{"com", "ai", "io", "dev", "co", "app"}
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 18
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 72
	}
	if opts.DictionaryPack != "" && opts.DictionaryPack != "none" && opts.MaxResults < 84 {
		opts.MaxResults = 84
	}

	candidates := generateCandidates(opts.Keywords, opts.TLDs, opts.DictionaryPack, opts.Mutations, opts.MaxResults)
	results := make([]ScanResultItem, len(candidates))

	sem := make(chan struct{}, opts.Concurrency)
	var wg sync.WaitGroup

	for idx, cand := range candidates {
		select {
		case <-ctx.Done():
			break
		default:
		}
		wg.Add(1)
		go func(i int, fullDomain string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			results[i] = checkSingleDomainAuthoritative(ctx, fullDomain)
		}(idx, cand)
	}
	wg.Wait()

	var (
		allItems       []ScanResultItem
		availableCount int
		takenCount     int
		highValueCount int
	)

	for _, item := range results {
		if item.Domain == "" {
			continue
		}
		if item.Available {
			availableCount++
		} else {
			takenCount++
		}
		if item.Valuation.Score >= 82 {
			highValueCount++
		}

		if opts.OnlyAvailable && !item.Available {
			continue
		}
		if opts.MinScore > 0 && item.Valuation.Score < opts.MinScore {
			continue
		}
		allItems = append(allItems, item)
	}

	// Sort: Exact user inputs & dictionary words first, then Available domains, then by Valuation Score descending
	seedSet := make(map[string]bool)
	for _, kw := range opts.Keywords {
		cleanKw := strings.ToLower(strings.TrimSpace(kw))
		if cleanKw != "" {
			seedSet[cleanKw] = true
			if !strings.Contains(cleanKw, ".") {
				for _, t := range opts.TLDs {
					seedSet[cleanKw+"."+strings.TrimPrefix(strings.ToLower(t), ".")] = true
				}
			}
		}
	}

	sort.SliceStable(allItems, func(i, j int) bool {
		iExact := seedSet[allItems[i].Domain]
		jExact := seedSet[allItems[j].Domain]
		if iExact != jExact {
			return iExact
		}
		if allItems[i].Valuation.IsDictionaryWord != allItems[j].Valuation.IsDictionaryWord && opts.DictionaryPack != "" && opts.DictionaryPack != "none" {
			return allItems[i].Valuation.IsDictionaryWord
		}
		if allItems[i].Available != allItems[j].Available {
			return allItems[i].Available
		}
		if allItems[i].Valuation.Score != allItems[j].Valuation.Score {
			return allItems[i].Valuation.Score > allItems[j].Valuation.Score
		}
		return len(allItems[i].Domain) < len(allItems[j].Domain)
	})

	return ScanReport{
		SeedKeywords:   opts.Keywords,
		DictionaryUsed: opts.DictionaryPack,
		TotalChecked:   len(candidates),
		AvailableCount: availableCount,
		TakenCount:     takenCount,
		HighValueCount: highValueCount,
		DurationMs:     time.Since(start).Milliseconds(),
		Items:          allItems,
	}
}

func generateCandidates(keywords []string, tlds []string, dictPack string, mutations bool, limit int) []string {
	seen := make(map[string]bool)
	var list []string

	addCandidate := func(root, tld string) {
		root = sanitizeLabel(root)
		tld = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tld)), ".")
		if root == "" || tld == "" {
			return
		}
		full := root + "." + tld
		if !seen[full] && len(list) < limit {
			seen[full] = true
			list = append(list, full)
		}
	}

	dictWords := GetDictionaryWords(dictPack)

	// 1. Exact user keywords across requested TLDs
	for _, rawKw := range keywords {
		rawKw = strings.ToLower(strings.TrimSpace(rawKw))
		if rawKw == "" {
			continue
		}
		if strings.Contains(rawKw, ".") {
			parts := strings.SplitN(rawKw, ".", 2)
			addCandidate(parts[0], parts[1])
			root := parts[0]
			for _, t := range tlds {
				addCandidate(root, t)
			}
			continue
		}
		for _, t := range tlds {
			addCandidate(rawKw, t)
		}
	}

	// 2. If a Dictionary Pack is selected, immediately prioritize pure dictionary words across the top TLDs
	if len(dictWords) > 0 {
		dictTLDs := tlds
		if len(dictTLDs) > 4 {
			dictTLDs = dictTLDs[:4]
		}
		for _, dw := range dictWords {
			for _, t := range dictTLDs {
				addCandidate(dw, t)
			}
		}

		// Also combine user seed + dictionary word (e.g. "nova" + "loom" -> "novaloom.com")
		for _, rawKw := range keywords {
			cleanRoot := sanitizeLabel(strings.Split(rawKw, ".")[0])
			if cleanRoot == "" {
				continue
			}
			for i, dw := range dictWords {
				if i >= 10 {
					break
				}
				for _, t := range dictTLDs[:minInt(2, len(dictTLDs))] {
					addCandidate(cleanRoot+dw, t)
				}
			}
		}
	}

	// 3. Brandable combinations (suffixes & prefixes)
	if mutations {
		primaryTLDs := tlds
		if len(primaryTLDs) > 4 {
			primaryTLDs = primaryTLDs[:4]
		}
		for _, rawKw := range keywords {
			rawKw = sanitizeLabel(strings.Split(rawKw, ".")[0])
			if rawKw == "" {
				continue
			}
			for _, sfx := range brandSuffixes {
				for _, t := range primaryTLDs {
					addCandidate(rawKw+sfx, t)
				}
			}
			for _, pfx := range brandPrefixes {
				for _, t := range primaryTLDs {
					addCandidate(pfx+rawKw, t)
				}
			}
		}
	}

	return list
}

// checkSingleDomainAuthoritative checks DNS (NS + A/AAAA + SERVFAIL detection) and falls back to TCP Port-43 WHOIS
// so parked/squatted domains without A/AAAA records (like agent.co) are never falsely marked Available.
func checkSingleDomainAuthoritative(ctx context.Context, fullDomain string) ScanResultItem {
	start := time.Now()
	parts := strings.SplitN(fullDomain, ".", 2)
	root := parts[0]
	tld := "com"
	if len(parts) > 1 {
		tld = parts[1]
	}

	dnsCtx, cancel := context.WithTimeout(ctx, 1600*time.Millisecond)
	defer cancel()

	r := net.DefaultResolver
	var nameservers []string
	dnsServFail := false

	if nss, err := r.LookupNS(dnsCtx, fullDomain); err == nil && len(nss) > 0 {
		for _, ns := range nss {
			nameservers = append(nameservers, strings.TrimSuffix(strings.ToLower(ns.Host), "."))
		}
	} else if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && !dnsErr.IsNotFound {
			// SERVFAIL / timeout in TLD zone means the domain IS delegated in the registry with lame/parked nameservers!
			dnsServFail = true
		}
	}

	hasIP := false
	if len(nameservers) == 0 {
		if ips, err := r.LookupHost(dnsCtx, fullDomain); err == nil && len(ips) > 0 {
			hasIP = true
		}
	}

	registered := len(nameservers) > 0 || hasIP
	var registeredAt, registrar string

	// If DNS returned no active NS/IP (or returned SERVFAIL like agent.co), verify with Port-43 WHOIS
	if !registered {
		whoisCtx, whoisCancel := context.WithTimeout(ctx, 2200*time.Millisecond)
		if wRec, wErr := QueryWhois(whoisCtx, fullDomain); wErr == nil && wRec.Registered {
			registered = true
			registeredAt = wRec.CreatedDate
			registrar = wRec.Registrar
			nameservers = wRec.Nameservers
		} else if dnsServFail {
			// Even if WHOIS rate-limits, a TLD zone SERVFAIL (not NXDOMAIN) indicates an existing registration delegation
			registered = true
			registrar = "Registered (Lame / Parked Nameservers)"
		}
		whoisCancel()
	}

	status := "Available"
	if registered {
		status = "Registered"
	}

	val := EvaluateDomainWithStatus(fullDomain, !registered)

	return ScanResultItem{
		Domain:       fullDomain,
		RootName:     root,
		TLD:          tld,
		Available:    !registered,
		Status:       status,
		RegisteredAt: registeredAt,
		Registrar:    registrar,
		Nameservers:  nameservers,
		Valuation:    val,
		LatencyMs:    time.Since(start).Milliseconds(),
	}
}

func sanitizeLabel(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
