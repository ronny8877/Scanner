package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HistoryReport details whether a domain was registered or active in the past
// using Internet Archive Wayback CDX records and Certificate Transparency (CT) logs.
type HistoryReport struct {
	Domain               string            `json:"domain"`
	PreviouslyRegistered bool              `json:"previouslyRegistered"`
	HistoryVerdict       string            `json:"historyVerdict"`
	SummaryNote          string            `json:"summaryNote"`
	FirstSeenAt          string            `json:"firstSeenAt,omitempty"`
	LastSeenAt           string            `json:"lastSeenAt,omitempty"`
	FirstSeenYear        int               `json:"firstSeenYear,omitempty"`
	LastSeenYear         int               `json:"lastSeenYear,omitempty"`
	TotalSpanYears       int               `json:"totalSpanYears"`
	WaybackSnapshots     int               `json:"waybackSnapshots"`
	ActiveYears          []string          `json:"activeYears,omitempty"`
	CertCount            int               `json:"certCount"`
	PastIssuers          []string          `json:"pastIssuers,omitempty"`
	PastSubdomains       []string          `json:"pastSubdomains,omitempty"`
	Milestones           []HistoryTimeline `json:"milestones,omitempty"`
	CheckLatencyMs       int64             `json:"checkLatencyMs"`
}

// HistoryTimeline represents a notable historical event found in archives or CT logs.
type HistoryTimeline struct {
	Date   string `json:"date"`
	Source string `json:"source"` // "Wayback Archive" | "CT Log (crt.sh)" | "RDAP Registry"
	Event  string `json:"event"`
}

type crtEntry struct {
	IssuerName string `json:"issuer_name"`
	NameValue  string `json:"name_value"`
	NotBefore  string `json:"not_before"`
	NotAfter   string `json:"not_after"`
}

// CheckDomainHistory queries Wayback Machine CDX and Certificate Transparency logs in parallel
// to determine if a domain was registered or hosted in the past.
func CheckDomainHistory(ctx context.Context, rawDomain string) HistoryReport {
	start := time.Now()
	clean := CleanDomainName(rawDomain)

	report := HistoryReport{
		Domain: clean,
	}

	var (
		wg               sync.WaitGroup
		wbFirst, wbLast  string
		wbSnapshots      int
		wbYears          []string
		ctCerts          int
		ctFirst, ctLast  string
		ctIssuers        []string
		ctSubdomains     []string
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		wbFirst, wbLast, wbSnapshots, wbYears = queryWaybackCDX(ctx, clean)
	}()

	go func() {
		defer wg.Done()
		ctCerts, ctFirst, ctLast, ctIssuers, ctSubdomains = queryCertTransparency(ctx, clean)
	}()

	wg.Wait()

	report.WaybackSnapshots = wbSnapshots
	report.ActiveYears = wbYears
	report.CertCount = ctCerts
	report.PastIssuers = ctIssuers
	report.PastSubdomains = ctSubdomains

	// Determine earliest and latest dates across both sources
	candidatesFirst := []string{}
	if wbFirst != "" {
		candidatesFirst = append(candidatesFirst, wbFirst)
	}
	if ctFirst != "" {
		candidatesFirst = append(candidatesFirst, ctFirst)
	}
	sort.Strings(candidatesFirst)
	if len(candidatesFirst) > 0 {
		report.FirstSeenAt = candidatesFirst[0]
		if len(report.FirstSeenAt) >= 4 {
			report.FirstSeenYear, _ = strconv.Atoi(report.FirstSeenAt[:4])
		}
	}

	candidatesLast := []string{}
	if wbLast != "" {
		candidatesLast = append(candidatesLast, wbLast)
	}
	if ctLast != "" {
		candidatesLast = append(candidatesLast, ctLast)
	}
	sort.Strings(candidatesLast)
	if len(candidatesLast) > 0 {
		report.LastSeenAt = candidatesLast[len(candidatesLast)-1]
		if len(report.LastSeenAt) >= 4 {
			report.LastSeenYear, _ = strconv.Atoi(report.LastSeenAt[:4])
		}
	}

	if report.FirstSeenYear > 0 && report.LastSeenYear >= report.FirstSeenYear {
		report.TotalSpanYears = report.LastSeenYear - report.FirstSeenYear + 1
	}

	// Build timeline milestones
	if wbFirst != "" {
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   wbFirst,
			Source: "Wayback Archive",
			Event:  fmt.Sprintf("First archived web snapshot captured on %s", wbFirst),
		})
	}
	if ctFirst != "" {
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   ctFirst,
			Source: "CT Log (crt.sh)",
			Event:  fmt.Sprintf("First TLS/SSL certificate issued (%d total historical certs)", ctCerts),
		})
	}
	if wbLast != "" && wbLast != wbFirst {
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   wbLast,
			Source: "Wayback Archive",
			Event:  fmt.Sprintf("Most recent web crawl capture (%d total snapshots across %d active years)", wbSnapshots, len(wbYears)),
		})
	}
	if ctLast != "" && ctLast != ctFirst {
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   ctLast,
			Source: "CT Log (crt.sh)",
			Event:  fmt.Sprintf("Most recent TLS certificate logged (%d subdomains observed)", len(ctSubdomains)),
		})
	}

	sort.Slice(report.Milestones, func(i, j int) bool {
		return report.Milestones[i].Date < report.Milestones[j].Date
	})

	if wbSnapshots > 0 || ctCerts > 0 {
		report.PreviouslyRegistered = true
		report.HistoryVerdict = "Previously Registered / Historical Footprint Found"
		report.SummaryNote = fmt.Sprintf(
			"Historical footprint confirmed between %d and %d (%d Wayback captures, %d TLS certificates, %d past subdomains).",
			report.FirstSeenYear, report.LastSeenYear, wbSnapshots, ctCerts, len(ctSubdomains),
		)
	} else {
		report.PreviouslyRegistered = false
		report.HistoryVerdict = "Clean Virgin Domain (No Past Registration Traces)"
		report.SummaryNote = "Zero historical snapshots in Internet Archive Wayback Machine and zero past SSL certificates in Certificate Transparency logs."
	}

	report.CheckLatencyMs = time.Since(start).Milliseconds()
	return report
}

func queryWaybackCDX(ctx context.Context, domain string) (firstDate, lastDate string, count int, years []string) {
	reqCtx, cancel := context.WithTimeout(ctx, 5500*time.Millisecond)
	defer cancel()

	// Query up to 250 timestamps collapsed by year-month
	url := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,statuscode&collapse=timestamp:6&limit=250", domain)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, nil
	}
	req.Header.Set("User-Agent", "Scanner-Domain-Studio/1.0")

	client := &http.Client{Timeout: 5500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, nil
	}

	var rows [][]string
	if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&rows); err != nil {
		return "", "", 0, nil
	}

	if len(rows) <= 1 {
		return "", "", 0, nil
	}

	yearMap := make(map[string]bool)
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) == 0 {
			continue
		}
		ts := rows[i][0]
		if len(ts) >= 8 {
			formatted := fmt.Sprintf("%s-%s-%s", ts[0:4], ts[4:6], ts[6:8])
			if firstDate == "" || formatted < firstDate {
				firstDate = formatted
			}
			if lastDate == "" || formatted > lastDate {
				lastDate = formatted
			}
			yearMap[ts[0:4]] = true
			count++
		}
	}

	for y := range yearMap {
		years = append(years, y)
	}
	sort.Strings(years)
	return firstDate, lastDate, count, years
}

func queryCertTransparency(ctx context.Context, domain string) (certCount int, firstDate, lastDate string, issuers, subdomains []string) {
	reqCtx, cancel := context.WithTimeout(ctx, 5500*time.Millisecond)
	defer cancel()

	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", "", nil, nil
	}
	req.Header.Set("User-Agent", "Scanner-Domain-Studio/1.0")

	client := &http.Client{Timeout: 5500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", "", nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", "", nil, nil
	}

	var entries []crtEntry
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&entries); err != nil {
		return 0, "", "", nil, nil
	}

	certCount = len(entries)
	issuerSet := make(map[string]bool)
	subSet := make(map[string]bool)

	for _, e := range entries {
		if len(e.NotBefore) >= 10 {
			d := e.NotBefore[:10]
			if firstDate == "" || d < firstDate {
				firstDate = d
			}
			if lastDate == "" || d > lastDate {
				lastDate = d
			}
		}

		// Extract clean Organization from issuer_name (e.g. O=Let's Encrypt)
		for _, part := range strings.Split(e.IssuerName, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "O=") {
				org := strings.Trim(strings.TrimPrefix(part, "O="), `"`)
				if org != "" && len(issuerSet) < 6 {
					issuerSet[org] = true
				}
			}
		}

		for _, name := range strings.Split(e.NameValue, "\n") {
			name = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "*.")))
			if name != "" && name != domain && strings.HasSuffix(name, "."+domain) && len(subSet) < 18 {
				subSet[name] = true
			}
		}
	}

	for k := range issuerSet {
		issuers = append(issuers, k)
	}
	sort.Strings(issuers)

	for s := range subSet {
		subdomains = append(subdomains, s)
	}
	sort.Strings(subdomains)

	return certCount, firstDate, lastDate, issuers, subdomains
}
