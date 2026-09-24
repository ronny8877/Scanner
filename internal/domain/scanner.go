package domain

import (
	"context"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScanOptions configures a bulk domain availability & value discovery scan.
type ScanOptions struct {
	Keywords       []string `json:"keywords"`
	TLDs           []string `json:"tlds"`
	DictionaryPack string   `json:"dictionaryPack,omitempty"` // Optional built-in dictionary pack ID
	Mutations      bool     `json:"mutations"`                // Generate brandable prefix/suffix combinations
	OnlyAvailable  bool     `json:"onlyAvailable"`            // Filter results to only available (empty) domains
	MinScore       int      `json:"minScore"`                 // Minimum valuation score (0-100)
	Concurrency    int      `json:"concurrency"`              // Worker pool size
	MaxResults     int      `json:"maxResults"`
}

// ScanResultItem represents one scanned domain candidate with availability and valuation.
type ScanResultItem struct {
	Domain        string    `json:"domain"`
	RootName      string    `json:"rootName"`
	TLD           string    `json:"tld"`
	Available     bool      `json:"available"`
	Status        string    `json:"status"` // "Available" | "Registered"
	RegisteredAt  string    `json:"registeredAt,omitempty"`
	Registrar     string    `json:"registrar,omitempty"`
	Nameservers   []string  `json:"nameservers,omitempty"`
	Valuation     Valuation `json:"valuation"`
	LatencyMs     int64     `json:"latencyMs"`
}

// ScanReport aggregates the full scan run statistics and sorted items.
type ScanReport struct {
	SeedKeywords   []string         `json:"seedKeywords"`
	DictionaryUsed string           `json:"dictionaryUsed,omitempty"`
	TotalChecked   int              `json:"totalChecked"`
	AvailableCount int              `json:"availableCount"`
	TakenCount     int              `json:"takenCount"`
	HighValueCount int              `json:"highValueCount"` // Available domains with score >= 73
	DurationMs     int64            `json:"durationMs"`
	Items          []ScanResultItem `json:"items"`
}

var defaultTLDs = []string{"com", "ai", "io", "dev", "co", "app"}

var brandPrefixes = []string{"get", "try", "use", "go", "open"}
var brandSuffixes = []string{"hq", "labs", "flow", "pulse", "hub", "core", "grid", "cloud", "sync", "base", "studio", "ai"}

// ScanDomains generates domain candidates from keywords/dictionary and checks their availability & value concurrently.
func ScanDomains(ctx context.Context, opts ScanOptions) ScanReport {
	start := time.Now()

	if len(opts.TLDs) == 0 {
		opts.TLDs = defaultTLDs
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 18
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 64
	}

	candidates := generateCandidates(opts.Keywords, opts.TLDs, opts.DictionaryPack, opts.Mutations, opts.MaxResults)

	type job struct {
		domain string
	}

	jobCh := make(chan job, len(candidates))
	resCh := make(chan ScanResultItem, len(candidates))

	var wg sync.WaitGroup
	workers := opts.Concurrency
	if workers > len(candidates) && len(candidates) > 0 {
		workers = len(candidates)
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				select {
				case <-ctx.Done():
					return
				default:
					resCh <- checkSingleDomainFast(ctx, j.domain)
				}
			}
		}()
	}

	for _, c := range candidates {
		jobCh <- job{domain: c}
	}
	close(jobCh)
	wg.Wait()
	close(resCh)

	var allItems []ScanResultItem
	availableCount := 0
	takenCount := 0
	highValueCount := 0

	for item := range resCh {
		if item.Available {
			availableCount++
			if item.Valuation.Score >= 73 {
				highValueCount++
			}
		} else {
			takenCount++
		}

		if opts.OnlyAvailable && !item.Available {
			continue
		}
		if opts.MinScore > 0 && item.Valuation.Score < opts.MinScore {
			continue
		}
		allItems = append(allItems, item)
	}

	// Sort: Available domains first, then by highest valuation score descending, then shortest length
	sort.Slice(allItems, func(i, j int) bool {
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

	// 1. Exact user keywords across all requested TLDs
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

		// If dictionary pack is active, combine user seed + dictionary word (e.g. "nova" + "loom" -> "novaloom.com")
		if len(dictWords) > 0 {
			topTLDs := tlds
			if len(topTLDs) > 3 {
				topTLDs = topTLDs[:3]
			}
			for i, dw := range dictWords {
				if i >= 12 {
					break
				}
				for _, t := range topTLDs {
					addCandidate(rawKw+dw, t)
				}
			}
		}
	}

	// 2. Dictionary words themselves across requested TLDs
	if len(dictWords) > 0 {
		for _, dw := range dictWords {
			for _, t := range tlds {
				addCandidate(dw, t)
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
			rawKw = sanitizeLabel(rawKw)
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

func checkSingleDomainFast(ctx context.Context, fullDomain string) ScanResultItem {
	start := time.Now()
	parts := strings.SplitN(fullDomain, ".", 2)
	root := parts[0]
	tld := "com"
	if len(parts) > 1 {
		tld = parts[1]
	}

	dnsCtx, cancel := context.WithTimeout(ctx, 1800*time.Millisecond)
	defer cancel()

	r := net.DefaultResolver
	var nameservers []string
	if nss, err := r.LookupNS(dnsCtx, fullDomain); err == nil && len(nss) > 0 {
		for _, ns := range nss {
			nameservers = append(nameservers, strings.TrimSuffix(strings.ToLower(ns.Host), "."))
		}
	}

	hasIP := false
	if len(nameservers) == 0 {
		if ips, err := r.LookupHost(dnsCtx, fullDomain); err == nil && len(ips) > 0 {
			hasIP = true
		}
	}

	registered := len(nameservers) > 0 || hasIP
	status := "Available"
	if registered {
		status = "Registered"
	}

	val := EvaluateDomainWithStatus(fullDomain, !registered)

	return ScanResultItem{
		Domain:      fullDomain,
		RootName:    root,
		TLD:         tld,
		Available:   !registered,
		Status:      status,
		Nameservers: nameservers,
		Valuation:   val,
		LatencyMs:   time.Since(start).Milliseconds(),
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
