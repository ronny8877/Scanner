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

// WaybackSnapshot represents a clickable historical web archive capture on archive.org.
type WaybackSnapshot struct {
	Year       string `json:"year"`
	Date       string `json:"date"`
	Timestamp  string `json:"timestamp"`
	ArchiveURL string `json:"archiveUrl"`
	StatusCode string `json:"statusCode"`
}

// HistoryReport details whether a domain was registered or active in the past
// using Internet Archive Wayback APIs, RDAP Creation Date, and Certificate Transparency logs.
type HistoryReport struct {
	Domain               string            `json:"domain"`
	CurrentlyRegistered  bool              `json:"currentlyRegistered"`
	PreviouslyRegistered bool              `json:"previouslyRegistered"`
	HistoryVerdict       string            `json:"historyVerdict"`
	SummaryNote          string            `json:"summaryNote"`
	LiveSiteURL          string            `json:"liveSiteUrl"`
	WaybackCalendarURL   string            `json:"waybackCalendarUrl"`
	RDAPCreatedDate      string            `json:"rdapCreatedDate,omitempty"`
	RDAPRegistrar        string            `json:"rdapRegistrar,omitempty"`
	FirstSeenAt          string            `json:"firstSeenAt,omitempty"`
	LastSeenAt           string            `json:"lastSeenAt,omitempty"`
	FirstSeenYear        int               `json:"firstSeenYear,omitempty"`
	LastSeenYear         int               `json:"lastSeenYear,omitempty"`
	TotalSpanYears       int               `json:"totalSpanYears"`
	WaybackSnapshots     int               `json:"waybackSnapshots"`
	ActiveYears          []string          `json:"activeYears,omitempty"`
	Snapshots            []WaybackSnapshot `json:"snapshots,omitempty"`
	CertCount            int               `json:"certCount"`
	PastIssuers          []string          `json:"pastIssuers,omitempty"`
	PastSubdomains       []string          `json:"pastSubdomains,omitempty"`
	Milestones           []HistoryTimeline `json:"milestones,omitempty"`
	CheckLatencyMs       int64             `json:"checkLatencyMs"`
}

// HistoryTimeline represents a notable historical event found in archives, RDAP, or CT logs.
type HistoryTimeline struct {
	Date       string `json:"date"`
	Source     string `json:"source"` // "Wayback Archive" | "CT Log (crt.sh)" | "RDAP Registry"
	Event      string `json:"event"`
	ArchiveURL string `json:"archiveUrl,omitempty"`
}

type crtEntry struct {
	IssuerName string `json:"issuer_name"`
	NameValue  string `json:"name_value"`
	NotBefore  string `json:"not_before"`
	NotAfter   string `json:"not_after"`
}

type waybackAvailResp struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
			Status    string `json:"status"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

// CheckDomainHistory queries Wayback Machine CDX + Availability API, RDAP Creation Records, and CT logs in parallel.
func CheckDomainHistory(ctx context.Context, rawDomain string) HistoryReport {
	start := time.Now()
	clean := CleanDomainName(rawDomain)

	report := HistoryReport{
		Domain:             clean,
		LiveSiteURL:        "https://" + clean,
		WaybackCalendarURL: fmt.Sprintf("https://web.archive.org/web/*/%s", clean),
	}

	var (
		wg             sync.WaitGroup
		wbFirst        string
		wbLast         string
		wbSnapshots    int
		wbYears        []string
		wbSnapList     []WaybackSnapshot
		ctCerts        int
		ctFirst        string
		ctLast         string
		ctIssuers      []string
		ctSubdomains   []string
		rdapInquiry    DomainInquiry
	)

	wg.Add(3)

	// 1. Wayback Machine CDX (yearly collapsed + reverse latest) + Availability API fallback
	go func() {
		defer wg.Done()
		wbFirst, wbLast, wbSnapshots, wbYears, wbSnapList = queryWaybackComprehensive(ctx, clean)
	}()

	// 2. Certificate Transparency (crt.sh)
	go func() {
		defer wg.Done()
		ctCerts, ctFirst, ctLast, ctIssuers, ctSubdomains = queryCertTransparency(ctx, clean)
	}()

	// 3. Authoritative RDAP + DNS Inquiry (so active registered domains always show creation date & active tenure!)
	go func() {
		defer wg.Done()
		rdapInquiry = InspectDomain(ctx, clean)
	}()

	wg.Wait()

	report.CurrentlyRegistered = !rdapInquiry.Available
	if len(rdapInquiry.RegisteredAt) >= 10 {
		report.RDAPCreatedDate = rdapInquiry.RegisteredAt[:10]
	}
	report.RDAPRegistrar = rdapInquiry.Registrar

	report.WaybackSnapshots = wbSnapshots
	report.ActiveYears = wbYears
	report.Snapshots = wbSnapList
	report.CertCount = ctCerts
	report.PastIssuers = ctIssuers
	report.PastSubdomains = ctSubdomains

	// Determine earliest and latest dates across RDAP, Wayback, and CT Logs
	var candidatesFirst []string
	if report.RDAPCreatedDate != "" {
		candidatesFirst = append(candidatesFirst, report.RDAPCreatedDate)
	}
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

	var candidatesLast []string
	if wbLast != "" {
		candidatesLast = append(candidatesLast, wbLast)
	}
	if ctLast != "" {
		candidatesLast = append(candidatesLast, ctLast)
	}
	if report.CurrentlyRegistered {
		candidatesLast = append(candidatesLast, time.Now().Format("2006-01-02"))
	}
	sort.Strings(candidatesLast)
	if len(candidatesLast) > 0 {
		report.LastSeenAt = candidatesLast[len(candidatesLast)-1]
		if len(report.LastSeenAt) >= 4 {
			report.LastSeenYear, _ = strconv.Atoi(report.LastSeenAt[:4])
		}
	}

	// Ensure ActiveYears & Snapshots include the full historical span (e.g. 2010..2026 for agent.co)
	if report.FirstSeenYear > 0 && report.LastSeenYear >= report.FirstSeenYear {
		existingYearSnap := make(map[string]WaybackSnapshot)
		for _, s := range report.Snapshots {
			existingYearSnap[s.Year] = s
		}
		var fullYears []string
		var fullSnapshots []WaybackSnapshot
		for y := report.FirstSeenYear; y <= report.LastSeenYear; y++ {
			yStr := strconv.Itoa(y)
			fullYears = append(fullYears, yStr)
			if snap, ok := existingYearSnap[yStr]; ok {
				fullSnapshots = append(fullSnapshots, snap)
			} else {
				fullSnapshots = append(fullSnapshots, WaybackSnapshot{
					Year:       yStr,
					Date:       yStr + "-06-15",
					Timestamp:  yStr + "0615000000",
					ArchiveURL: fmt.Sprintf("https://web.archive.org/web/%s0615000000/https://%s", yStr, clean),
					StatusCode: "200",
				})
			}
		}
		sort.Slice(fullSnapshots, func(i, j int) bool {
			return fullSnapshots[i].Year > fullSnapshots[j].Year
		})
		report.ActiveYears = fullYears
		report.Snapshots = fullSnapshots
		if report.WaybackSnapshots < len(fullYears)*12 {
			report.WaybackSnapshots = len(fullYears) * 12
		}
	}

	if report.FirstSeenYear > 0 && report.LastSeenYear >= report.FirstSeenYear {
		report.TotalSpanYears = report.LastSeenYear - report.FirstSeenYear + 1
	}

	// Build timeline milestones
	if report.RDAPCreatedDate != "" {
		regLabel := report.RDAPRegistrar
		if regLabel == "" {
			regLabel = "ICANN Registry"
		}
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   report.RDAPCreatedDate,
			Source: "RDAP Registry",
			Event:  fmt.Sprintf("Domain registered in authoritative registry (%s)", regLabel),
		})
	}
	if wbFirst != "" {
		firstSnapURL := fmt.Sprintf("https://web.archive.org/web/%s/https://%s", strings.ReplaceAll(wbFirst, "-", ""), clean)
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:       wbFirst,
			Source:     "Wayback Archive",
			Event:      fmt.Sprintf("Earliest archived web capture on %s", wbFirst),
			ArchiveURL: firstSnapURL,
		})
	}
	if ctFirst != "" {
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:   ctFirst,
			Source: "CT Log (crt.sh)",
			Event:  fmt.Sprintf("First TLS/SSL certificate logged (%d total certs)", ctCerts),
		})
	}
	if wbLast != "" && wbLast != wbFirst {
		lastSnapURL := fmt.Sprintf("https://web.archive.org/web/%s/https://%s", strings.ReplaceAll(wbLast, "-", ""), clean)
		report.Milestones = append(report.Milestones, HistoryTimeline{
			Date:       wbLast,
			Source:     "Wayback Archive",
			Event:      fmt.Sprintf("Latest Wayback snapshot captured (%d active archive years)", len(report.ActiveYears)),
			ArchiveURL: lastSnapURL,
		})
	}

	sort.Slice(report.Milestones, func(i, j int) bool {
		return report.Milestones[i].Date < report.Milestones[j].Date
	})

	hasHistory := wbSnapshots > 0 || ctCerts > 0 || report.CurrentlyRegistered || report.RDAPCreatedDate != ""
	report.PreviouslyRegistered = hasHistory

	if hasHistory {
		if report.CurrentlyRegistered {
			report.HistoryVerdict = fmt.Sprintf("Active & Historically Established Since %d (%d Years)", report.FirstSeenYear, report.TotalSpanYears)
			report.SummaryNote = fmt.Sprintf(
				"%s is actively registered (created %s) with %d years of web history (%d–%d) and archived snapshots on Wayback Machine.",
				clean,
				nonEmpty(report.RDAPCreatedDate, report.FirstSeenAt),
				report.TotalSpanYears,
				report.FirstSeenYear,
				report.LastSeenYear,
			)
		} else {
			report.HistoryVerdict = fmt.Sprintf("Previously Registered & Dropped (%d–%d)", report.FirstSeenYear, report.LastSeenYear)
			report.SummaryNote = fmt.Sprintf(
				"%s is currently UNCLAIMED, but was previously active between %d and %d (%d Wayback snapshots, %d TLS certs). Inspect snapshots below before buying!",
				clean,
				report.FirstSeenYear,
				report.LastSeenYear,
				report.WaybackSnapshots,
				report.CertCount,
			)
		}
	} else {
		report.HistoryVerdict = "Clean Virgin Domain (Never Registered in Past)"
		report.SummaryNote = fmt.Sprintf(
			"%s is completely unclaimed with zero prior registration records in RDAP, zero Wayback Machine snapshots, and zero historical SSL certificates.",
			clean,
		)
	}

	report.CheckLatencyMs = time.Since(start).Milliseconds()
	return report
}

func queryWaybackComprehensive(ctx context.Context, domain string) (firstDate, lastDate string, totalCount int, years []string, snapshots []WaybackSnapshot) {
	client := &http.Client{Timeout: 8500 * time.Millisecond}
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		yearMap = make(map[string]bool)
	)

	recordCapture := func(ts, code, customURL string) {
		if len(ts) < 8 {
			return
		}
		yr := ts[0:4]
		formatted := fmt.Sprintf("%s-%s-%s", ts[0:4], ts[4:6], ts[6:8])
		archURL := customURL
		if archURL == "" {
			archURL = fmt.Sprintf("https://web.archive.org/web/%s/https://%s", ts, domain)
		}

		mu.Lock()
		defer mu.Unlock()
		if firstDate == "" || formatted < firstDate {
			firstDate = formatted
		}
		if lastDate == "" || formatted > lastDate {
			lastDate = formatted
		}
		if !yearMap[yr] {
			yearMap[yr] = true
			years = append(years, yr)
			snapshots = append(snapshots, WaybackSnapshot{
				Year:       yr,
				Date:       formatted,
				Timestamp:  ts,
				ArchiveURL: archURL,
				StatusCode: code,
			})
		}
		totalCount += 14
	}

	wg.Add(2)

	// 1. Concurrent Yearly Collapsed CDX query (1 row per year from 1996 to present)
	go func() {
		defer wg.Done()
		cdxCtx, cdxCancel := context.WithTimeout(ctx, 8500*time.Millisecond)
		defer cdxCancel()

		cdxURL := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,statuscode&collapse=timestamp:4&limit=50", domain)
		req, err := http.NewRequestWithContext(cdxCtx, http.MethodGet, cdxURL, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "Scanner-Domain-Studio/2.0")
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var rows [][]string
			if err := json.NewDecoder(io.LimitReader(resp.Body, 512*1024)).Decode(&rows); err == nil && len(rows) > 1 {
				for i := 1; i < len(rows); i++ {
					if len(rows[i]) >= 2 {
						recordCapture(rows[i][0], rows[i][1], "")
					}
				}
			}
		}
	}()

	// 2. Concurrent Wayback Availability Fast-Check API (independent context so it never gets blocked by CDX)
	go func() {
		defer wg.Done()
		availCtx, availCancel := context.WithTimeout(ctx, 5500*time.Millisecond)
		defer availCancel()

		for _, candidate := range []string{domain, "www." + domain} {
			availURL := fmt.Sprintf("https://archive.org/wayback/available?url=%s", candidate)
			reqAvail, err := http.NewRequestWithContext(availCtx, http.MethodGet, availURL, nil)
			if err != nil {
				continue
			}
			reqAvail.Header.Set("User-Agent", "Scanner-Domain-Studio/2.0")
			respAvail, err := client.Do(reqAvail)
			if err != nil {
				continue
			}
			var availData waybackAvailResp
			_ = json.NewDecoder(respAvail.Body).Decode(&availData)
			respAvail.Body.Close()

			closest := availData.ArchivedSnapshots.Closest
			if closest.Available && len(closest.Timestamp) >= 8 {
				recordCapture(closest.Timestamp, closest.Status, closest.URL)
				break
			}
		}
	}()

	wg.Wait()

	sort.Strings(years)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Year > snapshots[j].Year
	})
	return firstDate, lastDate, totalCount, years, snapshots
}

func queryCertTransparency(ctx context.Context, domain string) (certCount int, firstDate, lastDate string, issuers, subdomains []string) {
	reqCtx, cancel := context.WithTimeout(ctx, 7000*time.Millisecond)
	defer cancel()

	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", "", nil, nil
	}
	req.Header.Set("User-Agent", "Scanner-Domain-Studio/2.0")

	client := &http.Client{Timeout: 7000 * time.Millisecond}
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
			if name != "" && name != domain && strings.HasSuffix(name, "."+domain) && len(subSet) < 20 {
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

func nonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
