package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rny/scanner/internal/webintel"
)

// CrawlOptions configures a website structure crawl.
type CrawlOptions struct {
	TargetURL string `json:"targetUrl"`
	MaxPages  int    `json:"maxPages"`
	MaxDepth  int    `json:"maxDepth"`
}

// PageInfo holds metadata extracted from a single crawled URL.
type PageInfo struct {
	URL           string   `json:"url"`
	Path          string   `json:"path"`
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	H1            string   `json:"h1,omitempty"`
	StatusCode    int      `json:"statusCode"`
	ContentType   string   `json:"contentType"`
	LatencyMs     int64    `json:"latencyMs"`
	Depth         int      `json:"depth"`
	InternalLinks int      `json:"internalLinks"`
	ExternalLinks int      `json:"externalLinks"`
	ChildrenPaths []string `json:"childrenPaths,omitempty"`
}

// SiteNode represents a node in the hierarchical URL path tree of the site.
type SiteNode struct {
	Segment     string      `json:"segment"`
	FullPath    string      `json:"fullPath"`
	Title       string      `json:"title,omitempty"`
	StatusCode  int         `json:"statusCode,omitempty"`
	LatencyMs   int64       `json:"latencyMs,omitempty"`
	InternalOut int         `json:"internalOut,omitempty"`
	ExternalOut int         `json:"externalOut,omitempty"`
	Children    []*SiteNode `json:"children,omitempty"`
}

// CrawlReport contains the complete crawl summary, flat pages list, hierarchical tree structure, and tracker telemetry.
type CrawlReport struct {
	RootURL       string                    `json:"rootUrl"`
	SeedPath      string                    `json:"seedPath"`
	Host          string                    `json:"host"`
	PagesCrawled  int                       `json:"pagesCrawled"`
	TotalLinks    int                       `json:"totalLinks"`
	ExternalCount int                       `json:"externalCount"`
	TechHeaders   []string                  `json:"techHeaders,omitempty"`
	DurationMs    int64                     `json:"durationMs"`
	Pages         []PageInfo                  `json:"pages"`
	Tree          *SiteNode                   `json:"tree"`
	Trackers      webintel.TrackerTelemetry   `json:"trackers"`
	TechStack     webintel.TechStackTelemetry `json:"techStack"`
}

var (
	reTitle = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	reH1    = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	reDesc  = regexp.MustCompile(`(?is)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	reHref  = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"'#]+)["']`)
	reTags  = regexp.MustCompile(`<[^>]*>`)
)

// CrawlSite crawls a domain or specific sub-URL (e.g. svelte.dev/docs/kit) and builds its hierarchical structure tree.
func CrawlSite(ctx context.Context, opts CrawlOptions) CrawlReport {
	start := time.Now()
	if opts.MaxPages <= 0 {
		opts.MaxPages = 25
	}
	if opts.MaxPages > 300 {
		opts.MaxPages = 300
	}
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 3
	}
	if opts.MaxDepth > 8 {
		opts.MaxDepth = 8
	}

	rawTarget := strings.TrimSpace(opts.TargetURL)
	if !strings.HasPrefix(rawTarget, "http://") && !strings.HasPrefix(rawTarget, "https://") {
		rawTarget = "https://" + rawTarget
	}

	parsedRoot, err := url.Parse(rawTarget)
	if err != nil || parsedRoot.Host == "" {
		return CrawlReport{RootURL: rawTarget, SeedPath: "/"}
	}
	rootHost := strings.ToLower(parsedRoot.Host)
	baseRoot := parsedRoot.Scheme + "://" + parsedRoot.Host

	seedPath := parsedRoot.Path
	if seedPath == "" {
		seedPath = "/"
	}
	if !strings.HasPrefix(seedPath, "/") {
		seedPath = "/" + seedPath
	}
	if len(seedPath) > 1 {
		seedPath = strings.TrimSuffix(seedPath, "/")
	}

	seedFullURL := baseRoot + seedPath
	if parsedRoot.RawQuery != "" {
		seedFullURL += "?" + parsedRoot.RawQuery
	}

	type queueItem struct {
		fullURL string
		path    string
		depth   int
	}

	visited := make(map[string]bool)
	var mu sync.Mutex
	var pages []PageInfo
	var htmlSamples []string
	var firstHeader http.Header
	techSet := make(map[string]bool)
	totalInternal := 0
	totalExternal := 0

	// Start directly from the exact requested URL/path (e.g. /@nyx)
	queue := []queueItem{{fullURL: seedFullURL, path: seedPath, depth: 0}}
	visited[seedPath] = true

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for len(queue) > 0 && len(pages) < opts.MaxPages {
		select {
		case <-ctx.Done():
			break
		default:
		}

		batchSize := 6
		if len(queue) < batchSize {
			batchSize = len(queue)
		}
		batch := queue[:batchSize]
		queue = queue[batchSize:]

		var wg sync.WaitGroup
		var nextWavePriority []queueItem
		var nextWaveGeneral []queueItem

		for _, item := range batch {
			if len(pages) >= opts.MaxPages {
				break
			}
			wg.Add(1)
			go func(qi queueItem) {
				defer wg.Done()
				page, discoveredPaths, headers, rawHTML, respHdr := fetchAndParsePage(ctx, client, qi.fullURL, qi.path, qi.depth, rootHost, baseRoot)

				mu.Lock()
				defer mu.Unlock()
				if len(pages) >= opts.MaxPages {
					return
				}
				pages = append(pages, page)
				if len(htmlSamples) < 8 && rawHTML != "" {
					htmlSamples = append(htmlSamples, rawHTML)
				}
				if firstHeader == nil && respHdr != nil {
					firstHeader = respHdr
				}
				totalInternal += page.InternalLinks
				totalExternal += page.ExternalLinks
				for _, h := range headers {
					techSet[h] = true
				}

				if qi.depth < opts.MaxDepth {
					for _, dp := range discoveredPaths {
						if !visited[dp] && len(visited) < opts.MaxPages*3 {
							visited[dp] = true
							nextItem := queueItem{
								fullURL: baseRoot + dp,
								path:    dp,
								depth:   qi.depth + 1,
							}
							// Prioritize sub-routes under the user's starting path (e.g. /@nyx/*)
							if seedPath != "/" && strings.HasPrefix(dp, seedPath) {
								nextWavePriority = append(nextWavePriority, nextItem)
							} else {
								nextWaveGeneral = append(nextWaveGeneral, nextItem)
							}
						}
					}
				}
			}(item)
		}
		wg.Wait()
		queue = append(queue, nextWavePriority...)
		queue = append(queue, nextWaveGeneral...)
	}

	// Sort pages so depth 0 (seedPath) is first, then by depth & path
	sort.Slice(pages, func(i, j int) bool {
		if pages[i].Path == seedPath {
			return true
		}
		if pages[j].Path == seedPath {
			return false
		}
		if pages[i].Depth != pages[j].Depth {
			return pages[i].Depth < pages[j].Depth
		}
		return pages[i].Path < pages[j].Path
	})

	var techHeaders []string
	for k := range techSet {
		techHeaders = append(techHeaders, k)
	}
	sort.Strings(techHeaders)

	tree := buildSiteTree(rootHost, seedPath, pages)
	trackers := webintel.DetectTrackersFromHTML(htmlSamples, firstHeader)
	techStack := webintel.DetectTechStackFromHTML(htmlSamples, nil, firstHeader)

	return CrawlReport{
		RootURL:       seedFullURL,
		SeedPath:      seedPath,
		Host:          rootHost,
		PagesCrawled:  len(pages),
		TotalLinks:    totalInternal,
		ExternalCount: totalExternal,
		TechHeaders:   techHeaders,
		DurationMs:    time.Since(start).Milliseconds(),
		Pages:         pages,
		Tree:          tree,
		Trackers:      trackers,
		TechStack:     techStack,
	}
}

func fetchAndParsePage(
	ctx context.Context,
	client *http.Client,
	targetURL, relPath string,
	depth int,
	rootHost, baseRoot string,
) (PageInfo, []string, []string, string, http.Header) {
	start := time.Now()
	info := PageInfo{
		URL:   targetURL,
		Path:  relPath,
		Depth: depth,
	}

	reqCtx, cancel := context.WithTimeout(ctx, 4800*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return info, nil, nil, "", nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		info.LatencyMs = time.Since(start).Milliseconds()
		return info, nil, nil, "", nil
	}
	defer resp.Body.Close()

	info.StatusCode = resp.StatusCode
	info.ContentType = resp.Header.Get("Content-Type")
	info.LatencyMs = time.Since(start).Milliseconds()

	var headers []string
	if srv := resp.Header.Get("Server"); srv != "" {
		headers = append(headers, "Server: "+srv)
	}
	if pwr := resp.Header.Get("X-Powered-By"); pwr != "" {
		headers = append(headers, "Powered-By: "+pwr)
	}
	if cf := resp.Header.Get("CF-Ray"); cf != "" {
		headers = append(headers, "CDN: Cloudflare")
	}
	if vercel := resp.Header.Get("X-Vercel-Id"); vercel != "" {
		headers = append(headers, "Platform: Vercel Edge")
	}
	if vMit := resp.Header.Get("X-Vercel-Mitigated"); vMit != "" {
		headers = append(headers, "Edge Shield: Vercel Challenge ("+vMit+")")
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 300*1024))
	if err != nil {
		return info, nil, headers, "", resp.Header
	}
	html := string(bodyBytes)

	if m := reTitle.FindStringSubmatch(html); len(m) > 1 {
		info.Title = cleanText(m[1])
	}
	if m := reH1.FindStringSubmatch(html); len(m) > 1 {
		info.H1 = cleanText(m[1])
	}
	if m := reDesc.FindStringSubmatch(html); len(m) > 1 {
		info.Description = cleanText(m[1])
	}

	// If edge firewall challenged the request (e.g. outbid.lol HTTP 429 Vercel Challenge), label cleanly
	if resp.StatusCode == 429 || resp.StatusCode == 403 || resp.Header.Get("X-Vercel-Mitigated") != "" || strings.Contains(strings.ToLower(info.Title), "security checkpoint") || strings.Contains(strings.ToLower(info.Title), "just a moment") {
		shieldLabel := "Edge WAF Challenge"
		if resp.Header.Get("X-Vercel-Id") != "" || resp.Header.Get("X-Vercel-Mitigated") != "" {
			shieldLabel = "Vercel Security Checkpoint (Active Edge Shield)"
		} else if resp.Header.Get("CF-Ray") != "" {
			shieldLabel = "Cloudflare Bot Mitigation Shield"
		}
		info.Title = fmt.Sprintf("%s%s — Protected Route (%s)", rootHost, relPath, shieldLabel)
		info.H1 = shieldLabel
		info.Description = fmt.Sprintf("Origin route %s is actively protected by %s (HTTP %d).", relPath, shieldLabel, resp.StatusCode)
	}

	basePageURL, _ := url.Parse(targetURL)
	matches := reHref.FindAllStringSubmatch(html, -1)
	seenPaths := make(map[string]bool)
	var discovered []string

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		href := strings.TrimSpace(m[1])
		if href == "" || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			continue
		}

		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		if basePageURL != nil {
			u = basePageURL.ResolveReference(u)
		}

		if u.Host != "" && strings.ToLower(u.Host) != rootHost && strings.ToLower(u.Host) != "www."+rootHost && "www."+strings.ToLower(u.Host) != rootHost {
			info.ExternalLinks++
			continue
		}

		cleanPath := u.Path
		if cleanPath == "" {
			cleanPath = "/"
		}
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}
		if len(cleanPath) > 1 {
			cleanPath = strings.TrimSuffix(cleanPath, "/")
		}

		lowerP := strings.ToLower(cleanPath)
		if strings.HasSuffix(lowerP, ".png") || strings.HasSuffix(lowerP, ".jpg") || strings.HasSuffix(lowerP, ".svg") ||
			strings.HasSuffix(lowerP, ".css") || strings.HasSuffix(lowerP, ".js") || strings.HasSuffix(lowerP, ".pdf") ||
			strings.HasSuffix(lowerP, ".xml") || strings.HasSuffix(lowerP, ".ico") || strings.HasSuffix(lowerP, ".webp") {
			continue
		}

		info.InternalLinks++
		if !seenPaths[cleanPath] {
			seenPaths[cleanPath] = true
			discovered = append(discovered, cleanPath)
		}
	}

	info.ChildrenPaths = discovered
	return info, discovered, headers, html, resp.Header
}

func buildSiteTree(host, seedPath string, pages []PageInfo) *SiteNode {
	rootSegment := host
	if seedPath != "" && seedPath != "/" {
		rootSegment = host + seedPath
	}
	root := &SiteNode{
		Segment:    rootSegment,
		FullPath:   seedPath,
		StatusCode: 200,
	}

	pageByPath := make(map[string]PageInfo)
	for _, p := range pages {
		pageByPath[p.Path] = p
		if p.Path == seedPath || (seedPath == "/" && p.Path == "/") {
			root.Title = p.Title
			root.StatusCode = p.StatusCode
			root.LatencyMs = p.LatencyMs
			root.InternalOut = p.InternalLinks
			root.ExternalOut = p.ExternalLinks
		}
	}

	allPaths := make(map[string]bool)
	for _, p := range pages {
		allPaths[p.Path] = true
		for _, child := range p.ChildrenPaths {
			if len(allPaths) < 90 {
				allPaths[child] = true
			}
		}
	}

	var sortedPaths []string
	for k := range allPaths {
		if k != "/" && k != seedPath {
			sortedPaths = append(sortedPaths, k)
		}
	}
	sort.Strings(sortedPaths)

	for _, pathStr := range sortedPaths {
		// If pathStr is under seedPath (e.g. /@nyx/gallery under /@nyx), attach relative to root
		relTrim := strings.Trim(pathStr, "/")
		if seedPath != "/" && strings.HasPrefix(pathStr, seedPath+"/") {
			relTrim = strings.TrimPrefix(pathStr, seedPath+"/")
		}
		segments := strings.Split(relTrim, "/")
		curr := root
		accumPath := ""
		if seedPath != "/" && strings.HasPrefix(pathStr, seedPath+"/") {
			accumPath = seedPath
		}
		for _, seg := range segments {
			if seg == "" {
				continue
			}
			accumPath += "/" + seg
			var found *SiteNode
			for _, c := range curr.Children {
				if c.Segment == seg {
					found = c
					break
				}
			}
			if found == nil {
				found = &SiteNode{
					Segment:  seg,
					FullPath: accumPath,
				}
				if p, ok := pageByPath[accumPath]; ok {
					found.Title = p.Title
					found.StatusCode = p.StatusCode
					found.LatencyMs = p.LatencyMs
					found.InternalOut = p.InternalLinks
					found.ExternalOut = p.ExternalLinks
				}
				curr.Children = append(curr.Children, found)
			}
			curr = found
		}
	}

	return root
}

func cleanText(s string) string {
	s = reTags.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 90 {
		return s[:87] + "..."
	}
	return s
}
