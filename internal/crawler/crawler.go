package crawler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
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

// CrawlReport contains the complete crawl summary, flat pages list, and hierarchical tree structure.
type CrawlReport struct {
	RootURL       string     `json:"rootUrl"`
	Host          string     `json:"host"`
	PagesCrawled  int        `json:"pagesCrawled"`
	TotalLinks    int        `json:"totalLinks"`
	ExternalCount int        `json:"externalCount"`
	TechHeaders   []string   `json:"techHeaders,omitempty"`
	DurationMs    int64      `json:"durationMs"`
	Pages         []PageInfo `json:"pages"`
	Tree          *SiteNode  `json:"tree"`
}

var (
	reTitle = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	reH1    = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	reDesc  = regexp.MustCompile(`(?is)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	reHref  = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"'#]+)["']`)
	reTags  = regexp.MustCompile(`<[^>]*>`)
)

// CrawlSite crawls a domain/website and builds its hierarchical structure tree.
func CrawlSite(ctx context.Context, opts CrawlOptions) CrawlReport {
	start := time.Now()
	if opts.MaxPages <= 0 {
		opts.MaxPages = 25
	}
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 3
	}

	rawTarget := strings.TrimSpace(opts.TargetURL)
	if !strings.HasPrefix(rawTarget, "http://") && !strings.HasPrefix(rawTarget, "https://") {
		rawTarget = "https://" + rawTarget
	}

	parsedRoot, err := url.Parse(rawTarget)
	if err != nil || parsedRoot.Host == "" {
		return CrawlReport{RootURL: rawTarget}
	}
	rootHost := strings.ToLower(parsedRoot.Host)
	baseRoot := parsedRoot.Scheme + "://" + parsedRoot.Host

	type queueItem struct {
		fullURL string
		path    string
		depth   int
	}

	visited := make(map[string]bool)
	var mu sync.Mutex
	var pages []PageInfo
	techSet := make(map[string]bool)
	totalInternal := 0
	totalExternal := 0

	queue := []queueItem{{fullURL: baseRoot + "/", path: "/", depth: 0}}
	visited["/"] = true

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for len(queue) > 0 && len(pages) < opts.MaxPages {
		// Process up to 5 URLs in parallel per wave
		batchSize := 5
		if len(queue) < batchSize {
			batchSize = len(queue)
		}
		batch := queue[:batchSize]
		queue = queue[batchSize:]

		var wg sync.WaitGroup
		var nextWave []queueItem

		for _, item := range batch {
			if len(pages) >= opts.MaxPages {
				break
			}
			wg.Add(1)
			go func(qi queueItem) {
				defer wg.Done()
				page, discoveredPaths, headers := fetchAndParsePage(ctx, client, qi.fullURL, qi.path, qi.depth, rootHost, baseRoot)

				mu.Lock()
				defer mu.Unlock()
				if len(pages) >= opts.MaxPages {
					return
				}
				pages = append(pages, page)
				totalInternal += page.InternalLinks
				totalExternal += page.ExternalLinks
				for _, h := range headers {
					techSet[h] = true
				}

				if qi.depth < opts.MaxDepth {
					for _, dp := range discoveredPaths {
						if !visited[dp] && len(visited) < opts.MaxPages*2 {
							visited[dp] = true
							nextWave = append(nextWave, queueItem{
								fullURL: baseRoot + dp,
								path:    dp,
								depth:   qi.depth + 1,
							})
						}
					}
				}
			}(item)
		}
		wg.Wait()
		queue = append(queue, nextWave...)
	}

	// Sort pages by path depth then alphabetically
	sort.Slice(pages, func(i, j int) bool {
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

	tree := buildSiteTree(rootHost, pages)

	return CrawlReport{
		RootURL:       baseRoot,
		Host:          rootHost,
		PagesCrawled:  len(pages),
		TotalLinks:    totalInternal,
		ExternalCount: totalExternal,
		TechHeaders:   techHeaders,
		DurationMs:    time.Since(start).Milliseconds(),
		Pages:         pages,
		Tree:          tree,
	}
}

func fetchAndParsePage(
	ctx context.Context,
	client *http.Client,
	targetURL, relPath string,
	depth int,
	rootHost, baseRoot string,
) (PageInfo, []string, []string) {
	start := time.Now()
	info := PageInfo{
		URL:   targetURL,
		Path:  relPath,
		Depth: depth,
	}

	reqCtx, cancel := context.WithTimeout(ctx, 4500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return info, nil, nil
	}
	req.Header.Set("User-Agent", "Scanner-Site-Structure-Bot/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		info.LatencyMs = time.Since(start).Milliseconds()
		return info, nil, nil
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
		headers = append(headers, "Platform: Vercel")
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return info, nil, headers
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

		// Skip static binary extensions
		lowerP := strings.ToLower(cleanPath)
		if strings.HasSuffix(lowerP, ".png") || strings.HasSuffix(lowerP, ".jpg") || strings.HasSuffix(lowerP, ".svg") ||
			strings.HasSuffix(lowerP, ".css") || strings.HasSuffix(lowerP, ".js") || strings.HasSuffix(lowerP, ".pdf") ||
			strings.HasSuffix(lowerP, ".xml") || strings.HasSuffix(lowerP, ".ico") {
			continue
		}

		info.InternalLinks++
		if !seenPaths[cleanPath] {
			seenPaths[cleanPath] = true
			discovered = append(discovered, cleanPath)
		}
	}

	info.ChildrenPaths = discovered
	return info, discovered, headers
}

func buildSiteTree(host string, pages []PageInfo) *SiteNode {
	root := &SiteNode{
		Segment:    host,
		FullPath:   "/",
		StatusCode: 200,
	}

	pageByPath := make(map[string]PageInfo)
	for _, p := range pages {
		pageByPath[p.Path] = p
		if p.Path == "/" {
			root.Title = p.Title
			root.StatusCode = p.StatusCode
			root.LatencyMs = p.LatencyMs
			root.InternalOut = p.InternalLinks
			root.ExternalOut = p.ExternalLinks
		}
	}

	// Ensure all crawled paths & their discovered internal routes are represented in the tree
	allPaths := make(map[string]bool)
	for _, p := range pages {
		allPaths[p.Path] = true
		for _, child := range p.ChildrenPaths {
			if len(allPaths) < 60 {
				allPaths[child] = true
			}
		}
	}

	var sortedPaths []string
	for k := range allPaths {
		if k != "/" {
			sortedPaths = append(sortedPaths, k)
		}
	}
	sort.Strings(sortedPaths)

	for _, pathStr := range sortedPaths {
		segments := strings.Split(strings.Trim(pathStr, "/"), "/")
		curr := root
		accumPath := ""
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
