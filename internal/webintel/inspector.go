package webintel

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DetectedTracker represents an Ad Network, Analytics Engine, Marketing Pixel, or Telemetry SDK detected on a site.
type DetectedTracker struct {
	Name        string `json:"name"`
	Category    string `json:"category"` // "AD_NETWORK", "ANALYTICS", "PIXEL_TRACKER", "SDK_TELEMETRY"
	Provider    string `json:"provider"`
	MatchedRule string `json:"matchedRule"`
	RiskLevel   string `json:"riskLevel"` // "LOW", "MODERATE", "HIGH"
	Description string `json:"description"`
}

// TrackerTelemetry summarizes all ad networks, analytics tools, and trackers found on a website.
type TrackerTelemetry struct {
	PrivacyGrade     string            `json:"privacyGrade"` // "A+", "A", "B", "C", "D"
	Verdict          string            `json:"verdict"`
	Summary          string            `json:"summary"`
	AdNetworksCount  int               `json:"adNetworksCount"`
	AnalyticsCount   int               `json:"analyticsCount"`
	PixelsCount      int               `json:"pixelsCount"`
	TelemetryCount   int               `json:"telemetryCount"`
	TotalDetected    int               `json:"totalDetected"`
	DetectedTrackers []DetectedTracker `json:"detectedTrackers"`
}

// CrawlerPermission represents whether a well-known bot (Googlebot, GPTBot, ClaudeBot, etc.) is allowed or blocked by robots.txt.
type CrawlerPermission struct {
	BotName     string `json:"botName"`
	Category    string `json:"category"` // "AI_LLM", "SEARCH_ENGINE", "SEO_BACKLINK", "SOCIAL_PREVIEW"
	Status      string `json:"status"`   // "ALLOWED", "PARTIAL", "BLOCKED"
	MatchedRule string `json:"matchedRule"`
}

// RobotsAgentGroup holds rules for a specific User-agent block in robots.txt.
type RobotsAgentGroup struct {
	UserAgent  string   `json:"userAgent"`
	Disallow   []string `json:"disallow"`
	Allow      []string `json:"allow"`
	CrawlDelay string   `json:"crawlDelay,omitempty"`
}

// SitemapEntry represents a single <url> or child <sitemap> entry with its update date.
type SitemapEntry struct {
	Loc        string `json:"loc"`
	Path       string `json:"path"`
	LastMod    string `json:"lastMod,omitempty"`
	AgeLabel   string `json:"ageLabel,omitempty"`
	ChangeFreq string `json:"changeFreq,omitempty"`
	Priority   string `json:"priority,omitempty"`
	IsChildMap bool   `json:"isChildMap,omitempty"`
}

// RobotsSitemapReport holds the complete analysis of robots.txt and sitemap.xml.
type RobotsSitemapReport struct {
	TargetURL          string              `json:"targetUrl"`
	Host               string              `json:"host"`
	CheckedAt          string              `json:"checkedAt"`
	DurationMs         int64               `json:"durationMs"`
	RobotsFound        bool                `json:"robotsFound"`
	RobotsURL          string              `json:"robotsUrl"`
	RobotsStatus       int                 `json:"robotsStatus"`
	RobotsSizeBytes    int                 `json:"robotsSizeBytes"`
	TotalDisallowCount int                 `json:"totalDisallowCount"`
	TotalAllowCount    int                 `json:"totalAllowCount"`
	DeclaredSitemaps   []string            `json:"declaredSitemaps"`
	AgentGroups        []RobotsAgentGroup  `json:"agentGroups"`
	BotMatrix          []CrawlerPermission `json:"botMatrix"`
	RawRobotsPreview   string              `json:"rawRobotsPreview"`
	SitemapFound       bool                `json:"sitemapFound"`
	SitemapURL         string              `json:"sitemapUrl"`
	SitemapStatus      int                 `json:"sitemapStatus"`
	IsSitemapIndex     bool                `json:"isSitemapIndex"`
	ChildSitemaps      []SitemapEntry      `json:"childSitemaps"`
	TotalUrlsCount     int                 `json:"totalUrlsCount"`
	NewestLastMod      string              `json:"newestLastMod,omitempty"`
	OldestLastMod      string              `json:"oldestLastMod,omitempty"`
	UpdatedLast7Days   int                 `json:"updatedLast7Days"`
	UpdatedLast30Days  int                 `json:"updatedLast30Days"`
	UpdatedLastYear    int                 `json:"updatedLastYear"`
	Entries            []SitemapEntry      `json:"entries"`
}

// MetaAuditCheck represents a single validation check on the page's meta tags.
type MetaAuditCheck struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Status  string `json:"status"` // "PASS", "WARN", "MISSING"
	Details string `json:"details"`
}

// MetaSocialReport holds all HTML meta tags, OpenGraph/Twitter card values, platform preview payloads, and ad/tracker telemetry.
type MetaSocialReport struct {
	TargetURL           string            `json:"targetUrl"`
	FinalURL            string            `json:"finalUrl"`
	Host                string            `json:"host"`
	StatusCode          int               `json:"statusCode"`
	DurationMs          int64             `json:"durationMs"`
	CheckedAt           string            `json:"checkedAt"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	CanonicalURL        string            `json:"canonicalUrl"`
	FaviconURL          string            `json:"faviconUrl"`
	ThemeColor          string            `json:"themeColor"`
	RobotsMeta          string            `json:"robotsMeta"`
	Author              string            `json:"author"`
	Generator           string            `json:"generator"`
	Language            string            `json:"language"`
	Charset             string            `json:"charset"`
	Viewport            string            `json:"viewport"`
	OGTitle             string            `json:"ogTitle"`
	OGDescription       string            `json:"ogDescription"`
	OGImage             string            `json:"ogImage"`
	OGURL               string            `json:"ogUrl"`
	OGSiteName          string            `json:"ogSiteName"`
	OGType              string            `json:"ogType"`
	OGLocale            string            `json:"ogLocale"`
	TwitterCard         string            `json:"twitterCard"`
	TwitterTitle        string            `json:"twitterTitle"`
	TwitterDescription  string            `json:"twitterDescription"`
	TwitterImage        string            `json:"twitterImage"`
	TwitterSite         string            `json:"twitterSite"`
	TwitterCreator      string            `json:"twitterCreator"`
	ResolvedTitle       string            `json:"resolvedTitle"`
	ResolvedDescription string            `json:"resolvedDescription"`
	ResolvedImage       string            `json:"resolvedImage"`
	ResolvedSiteName    string            `json:"resolvedSiteName"`
	ResolvedThemeColor  string            `json:"resolvedThemeColor"`
	SocialScore         int               `json:"socialScore"`
	SocialGrade         string            `json:"socialGrade"`
	AuditChecks         []MetaAuditCheck  `json:"auditChecks"`
	AllMetaTags         map[string]string `json:"allMetaTags"`
	Trackers            TrackerTelemetry  `json:"trackers"`
}

type trackerSignature struct {
	name        string
	category    string
	provider    string
	riskLevel   string
	description string
	patterns    []string
}

var knownSignatures = []trackerSignature{
	// Ad Networks
	{
		name:        "Google AdSense",
		category:    "AD_NETWORK",
		provider:    "Google",
		riskLevel:   "MODERATE",
		description: "Contextual display advertising & programmatic monetization script",
		patterns:    []string{"pagead2.googlesyndication.com", "adsbygoogle.js", "adsbygoogle"},
	},
	{
		name:        "Google Publisher Tag (DoubleClick)",
		category:    "AD_NETWORK",
		provider:    "Google Ad Manager",
		riskLevel:   "MODERATE",
		description: "Enterprise programmatic ad server & header bidding slots",
		patterns:    []string{"securepubads.g.doubleclick.net", "gpt.js", "googletag.pubads"},
	},
	{
		name:        "Mediavine Ad Management",
		category:    "AD_NETWORK",
		provider:    "Mediavine",
		riskLevel:   "MODERATE",
		description: "Full-service publisher display ad wrapper & auction script",
		patterns:    []string{"scripts.mediavine.com", "mediavine.com"},
	},
	{
		name:        "Raptive / AdThrive Ads",
		category:    "AD_NETWORK",
		provider:    "Raptive",
		riskLevel:   "MODERATE",
		description: "Premium publisher programmatic ad network",
		patterns:    []string{"ads.adthrive.com", "raptive.com"},
	},
	{
		name:        "Ezoic Ad Optimization",
		category:    "AD_NETWORK",
		provider:    "Ezoic",
		riskLevel:   "MODERATE",
		description: "AI-driven ad slot testing and programmatic monetization",
		patterns:    []string{"ezojs.com", "g.ezoic.net"},
	},
	{
		name:        "Carbon Ads",
		category:    "AD_NETWORK",
		provider:    "BuySellAds",
		riskLevel:   "LOW",
		description: "Developer & design-focused minimalist single-slot ad network",
		patterns:    []string{"cdn.carbonads.com", "_carbonads_js"},
	},
	{
		name:        "Taboola Content Discovery",
		category:    "AD_NETWORK",
		provider:    "Taboola",
		riskLevel:   "HIGH",
		description: "Sponsored content feed & behavioral recommendation widget",
		patterns:    []string{"cdn.taboola.com", "_taboola"},
	},
	{
		name:        "Outbrain Native Ads",
		category:    "AD_NETWORK",
		provider:    "Outbrain",
		riskLevel:   "HIGH",
		description: "Sponsored native article widgets & behavioral ad network",
		patterns:    []string{"widgets.outbrain.com", "outbrain.js"},
	},
	{
		name:        "Amazon Publisher Services (APS)",
		category:    "AD_NETWORK",
		provider:    "Amazon",
		riskLevel:   "MODERATE",
		description: "Unified Ad Marketplace & header bidding library",
		patterns:    []string{"c.amazon-adsystem.com", "apstag.js"},
	},
	{
		name:        "Criteo Retargeting Ads",
		category:    "AD_NETWORK",
		provider:    "Criteo",
		riskLevel:   "HIGH",
		description: "Cross-site dynamic retargeting & commerce display bidder",
		patterns:    []string{"static.criteo.net", "criteo_q"},
	},

	// Analytics & Telemetry
	{
		name:        "Google Analytics 4 (GA4)",
		category:    "ANALYTICS",
		provider:    "Google",
		riskLevel:   "LOW",
		description: "Event-based traffic, acquisition & user engagement analytics",
		patterns:    []string{"googletagmanager.com/gtag/js", "google-analytics.com/analytics.js", "gtag('config'"},
	},
	{
		name:        "Google Tag Manager (GTM)",
		category:    "ANALYTICS",
		provider:    "Google",
		riskLevel:   "MODERATE",
		description: "Dynamic container injecting marketing, analytics, and event scripts",
		patterns:    []string{"googletagmanager.com/gtm.js", "GTM-"},
	},
	{
		name:        "Cloudflare Web Analytics",
		category:    "ANALYTICS",
		provider:    "Cloudflare",
		riskLevel:   "LOW",
		description: "Privacy-first cookieless edge performance & visitor beacon",
		patterns:    []string{"static.cloudflareinsights.com/beacon.min.js", "__cfBeacon"},
	},
	{
		name:        "Plausible Analytics",
		category:    "ANALYTICS",
		provider:    "Plausible",
		riskLevel:   "LOW",
		description: "Lightweight open-source cookieless web analytics",
		patterns:    []string{"plausible.io/js/"},
	},
	{
		name:        "PostHog Product Analytics",
		category:    "ANALYTICS",
		provider:    "PostHog",
		riskLevel:   "LOW",
		description: "Product analytics, feature flags & session replay suite",
		patterns:    []string{"app.posthog.com", "us.i.posthog.com", "eu.i.posthog.com", "posthog.init"},
	},
	{
		name:        "Vercel Web Analytics",
		category:    "ANALYTICS",
		provider:    "Vercel",
		riskLevel:   "LOW",
		description: "First-party edge visitor & Web Vitals telemetry",
		patterns:    []string{"/_vercel/insights/script.js", "va.vercel-scripts.com"},
	},
	{
		name:        "Umami Analytics",
		category:    "ANALYTICS",
		provider:    "Umami",
		riskLevel:   "LOW",
		description: "Privacy-focused cookieless web telemetry",
		patterns:    []string{"analytics.umami.is", "umami.is/script.js", "data-website-id="},
	},
	{
		name:        "Mixpanel",
		category:    "ANALYTICS",
		provider:    "Mixpanel",
		riskLevel:   "MODERATE",
		description: "Funnel & user cohort event analytics",
		patterns:    []string{"cdn.mxpnl.com", "mixpanel.init"},
	},
	{
		name:        "Amplitude Analytics",
		category:    "ANALYTICS",
		provider:    "Amplitude",
		riskLevel:   "MODERATE",
		description: "Product intelligence & behavioral event tracking",
		patterns:    []string{"cdn.amplitude.com"},
	},
	{
		name:        "Segment Customer Data Hub",
		category:    "ANALYTICS",
		provider:    "Twilio Segment",
		riskLevel:   "MODERATE",
		description: "Customer data pipeline routing events to downstream trackers",
		patterns:    []string{"cdn.segment.com/analytics.js"},
	},
	{
		name:        "Microsoft Clarity",
		category:    "ANALYTICS",
		provider:    "Microsoft",
		riskLevel:   "MODERATE",
		description: "Heatmaps and full DOM session recording telemetry",
		patterns:    []string{"clarity.ms/tag/"},
	},
	{
		name:        "Hotjar Session Recording",
		category:    "ANALYTICS",
		provider:    "Hotjar",
		riskLevel:   "MODERATE",
		description: "Click heatmaps, scroll depth & user session replay",
		patterns:    []string{"static.hotjar.com", "_hjSettings"},
	},

	// Pixels & Retargeting
	{
		name:        "Meta / Facebook Pixel",
		category:    "PIXEL_TRACKER",
		provider:    "Meta",
		riskLevel:   "HIGH",
		description: "Cross-site conversion tracking & custom audience retargeting beacon",
		patterns:    []string{"connect.facebook.net", "fbevents.js", "fbq('init'"},
	},
	{
		name:        "TikTok Ads Pixel",
		category:    "PIXEL_TRACKER",
		provider:    "TikTok",
		riskLevel:   "HIGH",
		description: "Ad attribution and behavioral conversion pixel",
		patterns:    []string{"analytics.tiktok.com/i18n/pixel"},
	},
	{
		name:        "LinkedIn Insight Tag",
		category:    "PIXEL_TRACKER",
		provider:    "LinkedIn",
		riskLevel:   "MODERATE",
		description: "B2B demographic & campaign conversion pixel",
		patterns:    []string{"snap.licdn.com/li.lms-analytics", "_linkedin_partner_id"},
	},
	{
		name:        "X / Twitter Universal Website Tag",
		category:    "PIXEL_TRACKER",
		provider:    "X Corp",
		riskLevel:   "MODERATE",
		description: "Conversion tracking & audience tailoring script",
		patterns:    []string{"static.ads-twitter.com/uwt.js", "twq('init'"},
	},
	{
		name:        "HubSpot Marketing Beacon",
		category:    "PIXEL_TRACKER",
		provider:    "HubSpot",
		riskLevel:   "MODERATE",
		description: "CRM visitor identification, forms & lead tracking",
		patterns:    []string{"js.hs-scripts.com", "js.hs-analytics.net"},
	},

	// SDKs & Consent
	{
		name:        "Sentry Error Monitoring",
		category:    "SDK_TELEMETRY",
		provider:    "Sentry",
		riskLevel:   "LOW",
		description: "Client-side exception & crash diagnostics",
		patterns:    []string{"browser.sentry-cdn.com", "sentry.io"},
	},
	{
		name:        "Datadog RUM",
		category:    "SDK_TELEMETRY",
		provider:    "Datadog",
		riskLevel:   "LOW",
		description: "Real User Monitoring & frontend performance tracing",
		patterns:    []string{"datadoghq-browser-agent.com"},
	},
	{
		name:        "Intercom Messenger",
		category:    "SDK_TELEMETRY",
		provider:    "Intercom",
		riskLevel:   "LOW",
		description: "Live customer support widget & visitor ping",
		patterns:    []string{"widget.intercom.io", "intercomSettings"},
	},
	{
		name:        "Crisp Live Chat",
		category:    "SDK_TELEMETRY",
		provider:    "Crisp",
		riskLevel:   "LOW",
		description: "Customer chat widget & visitor session beacon",
		patterns:    []string{"client.crisp.chat", "CRISP_WEBSITE_ID"},
	},
	{
		name:        "OneTrust / Cookiebot Consent",
		category:    "SDK_TELEMETRY",
		provider:    "CMP",
		riskLevel:   "LOW",
		description: "GDPR/CCPA cookie consent management banner",
		patterns:    []string{"cdn.cookielaw.org", "consent.cookiebot.com"},
	},
}

// DetectTrackersFromHTML scans one or more HTML documents (and optional headers) to detect Ad Networks, Analytics, and Pixels.
func DetectTrackersFromHTML(htmlBlobs []string, headers http.Header) TrackerTelemetry {
	combined := strings.ToLower(strings.Join(htmlBlobs, "\n"))
	var detected []DetectedTracker
	seen := make(map[string]bool)

	for _, sig := range knownSignatures {
		for _, pat := range sig.patterns {
			if strings.Contains(combined, strings.ToLower(pat)) {
				if !seen[sig.name] {
					seen[sig.name] = true
					detected = append(detected, DetectedTracker{
						Name:        sig.name,
						Category:    sig.category,
						Provider:    sig.provider,
						MatchedRule: pat,
						RiskLevel:   sig.riskLevel,
						Description: sig.description,
					})
				}
				break
			}
		}
	}

	// Check Cloudflare header for edge insights
	if headers != nil && headers.Get("CF-Ray") != "" && !seen["Cloudflare Edge Network"] {
		// Only note if no other analytics found or as edge telemetry
	}

	adCount := 0
	analyticsCount := 0
	pixelCount := 0
	sdkCount := 0

	for _, d := range detected {
		switch d.Category {
		case "AD_NETWORK":
			adCount++
		case "ANALYTICS":
			analyticsCount++
		case "PIXEL_TRACKER":
			pixelCount++
		case "SDK_TELEMETRY":
			sdkCount++
		}
	}

	total := len(detected)
	grade := "A+"
	verdict := "Zero-Ad & Zero-Tracker Clean Footprint"
	summary := "No commercial ad networks, behavioral retargeting pixels, or third-party tracking scripts detected in page source."

	if adCount == 0 && pixelCount == 0 && analyticsCount == 1 {
		grade = "A"
		verdict = "Minimal First-Party Analytics Only"
		summary = fmt.Sprintf("Clean ad-free surface using %d analytics tool (%s) with zero behavioral retargeting pixels.", analyticsCount, detected[0].Name)
	} else if adCount == 0 && pixelCount == 0 && analyticsCount > 1 {
		grade = "A-"
		verdict = "Ad-Free · Product Telemetry Stack"
		summary = fmt.Sprintf("No display ad networks or third-party ad pixels; runs %d analytics/telemetry scripts.", analyticsCount+sdkCount)
	} else if adCount > 0 || pixelCount > 0 {
		if adCount+pixelCount <= 2 {
			grade = "B"
			verdict = "Monetized / Marketing-Tracked Surface"
			summary = fmt.Sprintf("Detected %d ad network(s), %d analytics engine(s), and %d retargeting pixel(s).", adCount, analyticsCount, pixelCount)
		} else {
			grade = "C"
			verdict = "Heavy Ad Network & Multi-Tracker Footprint"
			summary = fmt.Sprintf("Page embeds %d ad network(s), %d analytics tool(s), and %d behavioral marketing pixel(s).", adCount, analyticsCount, pixelCount)
		}
	}

	return TrackerTelemetry{
		PrivacyGrade:     grade,
		Verdict:          verdict,
		Summary:          summary,
		AdNetworksCount:  adCount,
		AnalyticsCount:   analyticsCount,
		PixelsCount:      pixelCount,
		TelemetryCount:   sdkCount,
		TotalDetected:    total,
		DetectedTrackers: detected,
	}
}

// InspectRobotsAndSitemap fetches and parses both robots.txt and sitemap.xml for a domain or URL.
func InspectRobotsAndSitemap(ctx context.Context, target string) RobotsSitemapReport {
	start := time.Now()
	_, host, _ := normalizeURLAndHost(target)
	baseURL := "https://" + host

	report := RobotsSitemapReport{
		TargetURL: baseURL,
		Host:      host,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		RobotsURL: baseURL + "/robots.txt",
	}

	client := &http.Client{Timeout: 6 * time.Second}

	// 1. Fetch & parse robots.txt using adaptive multi-profile HTTP strategy
	bodyBytes, statusCode, _ := fetchWithAdaptiveStrategy(ctx, client, report.RobotsURL, 256*1024)
	report.RobotsStatus = statusCode
	report.RobotsSizeBytes = len(bodyBytes)
	bodyStr := string(bodyBytes)
	if statusCode >= 200 && statusCode < 300 && len(bodyBytes) > 0 && !strings.HasPrefix(strings.TrimSpace(strings.ToLower(bodyStr)), "<!doctype html") && !isEdgeChallengePage(bodyStr) {
		report.RobotsFound = true
		parseRobotsTxt(bodyStr, &report)
	}

	// Always compute the 12-Bot Permission Matrix (even if robots.txt is missing -> all allowed)
	report.BotMatrix = evaluateBotMatrix(report.AgentGroups, report.RobotsFound)

	// 2. Discover & fetch Sitemap.xml
	candidateSitemaps := append([]string{}, report.DeclaredSitemaps...)
	defaultCandidates := []string{
		baseURL + "/sitemap.xml",
		baseURL + "/sitemap_index.xml",
		baseURL + "/sitemap-0.xml",
	}
	for _, dc := range defaultCandidates {
		if !containsStr(candidateSitemaps, dc) {
			candidateSitemaps = append(candidateSitemaps, dc)
		}
	}

	for _, smURL := range candidateSitemaps {
		if fetchAndParseSitemap(ctx, client, smURL, &report, true) {
			break
		}
	}

	report.DurationMs = time.Since(start).Milliseconds()
	return report
}

func parseRobotsTxt(raw string, report *RobotsSitemapReport) {
	lines := strings.Split(raw, "\n")
	if len(raw) > 4500 {
		report.RawRobotsPreview = raw[:4500] + "\n... (truncated)"
	} else {
		report.RawRobotsPreview = strings.TrimSpace(raw)
	}

	var groups []RobotsAgentGroup
	var currentAgents []string
	groupMap := make(map[string]*RobotsAgentGroup)

	getOrCreateGroup := func(ua string) *RobotsAgentGroup {
		if g, ok := groupMap[ua]; ok {
			return g
		}
		g := &RobotsAgentGroup{
			UserAgent: ua,
			Disallow:  []string{},
			Allow:     []string{},
		}
		groupMap[ua] = g
		return g
	}

	prevWasUserAgent := false

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			prevWasUserAgent = false
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			if !prevWasUserAgent {
				currentAgents = nil
			}
			ua := val
			if ua == "" {
				ua = "*"
			}
			currentAgents = append(currentAgents, ua)
			getOrCreateGroup(ua)
			prevWasUserAgent = true
		case "disallow":
			prevWasUserAgent = false
			if len(currentAgents) == 0 {
				currentAgents = []string{"*"}
			}
			if val != "" {
				for _, ua := range currentAgents {
					g := getOrCreateGroup(ua)
					g.Disallow = append(g.Disallow, val)
					report.TotalDisallowCount++
				}
			}
		case "allow":
			prevWasUserAgent = false
			if len(currentAgents) == 0 {
				currentAgents = []string{"*"}
			}
			if val != "" {
				for _, ua := range currentAgents {
					g := getOrCreateGroup(ua)
					g.Allow = append(g.Allow, val)
					report.TotalAllowCount++
				}
			}
		case "crawl-delay":
			prevWasUserAgent = false
			for _, ua := range currentAgents {
				g := getOrCreateGroup(ua)
				g.CrawlDelay = val
			}
		case "sitemap":
			prevWasUserAgent = false
			if val != "" && !containsStr(report.DeclaredSitemaps, val) {
				report.DeclaredSitemaps = append(report.DeclaredSitemaps, val)
			}
		}
	}

	for _, g := range groupMap {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].UserAgent == "*" {
			return true
		}
		if groups[j].UserAgent == "*" {
			return false
		}
		return groups[i].UserAgent < groups[j].UserAgent
	})
	report.AgentGroups = groups
}

func evaluateBotMatrix(groups []RobotsAgentGroup, robotsFound bool) []CrawlerPermission {
	bots := []struct {
		name     string
		category string
	}{
		{"Googlebot", "SEARCH_ENGINE"},
		{"Bingbot", "SEARCH_ENGINE"},
		{"GPTBot (OpenAI)", "AI_LLM"},
		{"ClaudeBot (Anthropic)", "AI_LLM"},
		{"Google-Extended (Gemini)", "AI_LLM"},
		{"CCBot (CommonCrawl)", "AI_LLM"},
		{"PerplexityBot", "AI_LLM"},
		{"Bytespider (ByteDance)", "AI_LLM"},
		{"AhrefsBot", "SEO_BACKLINK"},
		{"SemrushBot", "SEO_BACKLINK"},
		{"Twitterbot (X Card)", "SOCIAL_PREVIEW"},
		{"facebookexternalhit", "SOCIAL_PREVIEW"},
	}

	var wildcardGroup *RobotsAgentGroup
	for i := range groups {
		if groups[i].UserAgent == "*" {
			wildcardGroup = &groups[i]
			break
		}
	}

	var matrix []CrawlerPermission
	for _, b := range bots {
		if !robotsFound {
			matrix = append(matrix, CrawlerPermission{
				BotName:     b.name,
				Category:    b.category,
				Status:      "ALLOWED",
				MatchedRule: "No robots.txt restriction (Default Allow)",
			})
			continue
		}

		token := strings.ToLower(strings.Split(b.name, " ")[0])
		var matchedGroup *RobotsAgentGroup
		for i := range groups {
			uaLower := strings.ToLower(groups[i].UserAgent)
			if uaLower == token || strings.Contains(token, uaLower) && uaLower != "*" {
				matchedGroup = &groups[i]
				break
			}
		}
		if matchedGroup == nil {
			matchedGroup = wildcardGroup
		}

		if matchedGroup == nil || len(matchedGroup.Disallow) == 0 {
			matrix = append(matrix, CrawlerPermission{
				BotName:     b.name,
				Category:    b.category,
				Status:      "ALLOWED",
				MatchedRule: "Allowed (0 Disallow rules)",
			})
			continue
		}

		rootBlocked := false
		for _, d := range matchedGroup.Disallow {
			if strings.TrimSpace(d) == "/" {
				rootBlocked = true
				break
			}
		}

		if rootBlocked && len(matchedGroup.Allow) == 0 {
			matrix = append(matrix, CrawlerPermission{
				BotName:     b.name,
				Category:    b.category,
				Status:      "BLOCKED",
				MatchedRule: fmt.Sprintf("User-agent: %s → Disallow: /", matchedGroup.UserAgent),
			})
		} else {
			sampleRule := matchedGroup.Disallow[0]
			if len(matchedGroup.Disallow) > 1 {
				sampleRule = fmt.Sprintf("%s (+%d paths)", sampleRule, len(matchedGroup.Disallow)-1)
			}
			matrix = append(matrix, CrawlerPermission{
				BotName:     b.name,
				Category:    b.category,
				Status:      "PARTIAL",
				MatchedRule: fmt.Sprintf("User-agent: %s → Disallow: %s", matchedGroup.UserAgent, sampleRule),
			})
		}
	}

	return matrix
}

type xmlURLSet struct {
	URLs []struct {
		Loc        string `xml:"loc"`
		LastMod    string `xml:"lastmod"`
		ChangeFreq string `xml:"changefreq"`
		Priority   string `xml:"priority"`
	} `xml:"url"`
}

type xmlSitemapIndex struct {
	Sitemaps []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod"`
	} `xml:"sitemap"`
}

func fetchAndParseSitemap(ctx context.Context, client *http.Client, smURL string, report *RobotsSitemapReport, followIndex bool) bool {
	reqCtx, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, smURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Scanner-Studio-Sitemap-Inspector/2.0")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil || len(bodyBytes) == 0 {
		return false
	}

	bodyStr := string(bodyBytes)
	if !strings.Contains(bodyStr, "<urlset") && !strings.Contains(bodyStr, "<sitemapindex") {
		return false
	}

	report.SitemapFound = true
	report.SitemapURL = smURL
	report.SitemapStatus = resp.StatusCode

	now := time.Now()

	if strings.Contains(bodyStr, "<sitemapindex") {
		report.IsSitemapIndex = true
		var idx xmlSitemapIndex
		_ = xml.Unmarshal(bodyBytes, &idx)
		for _, sm := range idx.Sitemaps {
			loc := strings.TrimSpace(sm.Loc)
			if loc == "" {
				continue
			}
			cleanDate, ageLabel, daysAgo := parseSitemapDate(sm.LastMod, now)
			updateDateCounters(report, cleanDate, daysAgo)
			report.ChildSitemaps = append(report.ChildSitemaps, SitemapEntry{
				Loc:        loc,
				Path:       extractPathFromURL(loc),
				LastMod:    cleanDate,
				AgeLabel:   ageLabel,
				IsChildMap: true,
			})
		}

		// Follow first child sitemap so the user also sees actual page <url> entries and update dates
		if followIndex && len(report.ChildSitemaps) > 0 {
			firstChild := report.ChildSitemaps[0].Loc
			fetchChildURLSet(ctx, client, firstChild, report, now)
		}
		return true
	}

	var urlset xmlURLSet
	_ = xml.Unmarshal(bodyBytes, &urlset)
	report.TotalUrlsCount = len(urlset.URLs)

	for i, u := range urlset.URLs {
		loc := strings.TrimSpace(u.Loc)
		if loc == "" {
			continue
		}
		cleanDate, ageLabel, daysAgo := parseSitemapDate(u.LastMod, now)
		updateDateCounters(report, cleanDate, daysAgo)

		if i < 100 {
			report.Entries = append(report.Entries, SitemapEntry{
				Loc:        loc,
				Path:       extractPathFromURL(loc),
				LastMod:    cleanDate,
				AgeLabel:   ageLabel,
				ChangeFreq: strings.TrimSpace(u.ChangeFreq),
				Priority:   strings.TrimSpace(u.Priority),
			})
		}
	}

	return true
}

func fetchChildURLSet(ctx context.Context, client *http.Client, childURL string, report *RobotsSitemapReport, now time.Time) {
	reqCtx, cancel := context.WithTimeout(ctx, 4000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, childURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Scanner-Studio-Sitemap-Inspector/2.0")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return
	}

	var urlset xmlURLSet
	if err := xml.Unmarshal(bodyBytes, &urlset); err != nil {
		return
	}
	report.TotalUrlsCount += len(urlset.URLs)

	for i, u := range urlset.URLs {
		loc := strings.TrimSpace(u.Loc)
		if loc == "" {
			continue
		}
		cleanDate, ageLabel, daysAgo := parseSitemapDate(u.LastMod, now)
		updateDateCounters(report, cleanDate, daysAgo)

		if i < 80 {
			report.Entries = append(report.Entries, SitemapEntry{
				Loc:        loc,
				Path:       extractPathFromURL(loc),
				LastMod:    cleanDate,
				AgeLabel:   ageLabel,
				ChangeFreq: strings.TrimSpace(u.ChangeFreq),
				Priority:   strings.TrimSpace(u.Priority),
			})
		}
	}
}

func parseSitemapDate(raw string, now time.Time) (string, string, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", -1
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			days := int(now.Sub(t).Hours() / 24)
			if days < 0 {
				days = 0
			}
			var age string
			switch {
			case days == 0:
				age = "Today"
			case days == 1:
				age = "Yesterday"
			case days < 30:
				age = fmt.Sprintf("%dd ago", days)
			case days < 365:
				age = fmt.Sprintf("%dmo ago", days/30)
			default:
				age = fmt.Sprintf("%dyr ago", days/365)
			}
			return t.Format("2006-01-02"), age, days
		}
	}
	if len(raw) >= 10 {
		return raw[:10], "", -1
	}
	return raw, "", -1
}

func updateDateCounters(report *RobotsSitemapReport, cleanDate string, daysAgo int) {
	if cleanDate == "" {
		return
	}
	if report.NewestLastMod == "" || cleanDate > report.NewestLastMod {
		report.NewestLastMod = cleanDate
	}
	if report.OldestLastMod == "" || cleanDate < report.OldestLastMod {
		report.OldestLastMod = cleanDate
	}
	if daysAgo >= 0 {
		if daysAgo <= 7 {
			report.UpdatedLast7Days++
		}
		if daysAgo <= 30 {
			report.UpdatedLast30Days++
		}
		if daysAgo <= 365 {
			report.UpdatedLastYear++
		}
	}
}

var (
	reTitleTag = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	reHtmlLang = regexp.MustCompile(`(?is)<html[^>]+lang=["']([^"']+)["']`)
	reMetaTag  = regexp.MustCompile(`(?is)<meta\s+[^>]*>`)
	reLinkTag  = regexp.MustCompile(`(?is)<link\s+[^>]*>`)
	reAttrPair = regexp.MustCompile(`(?is)([a-zA-Z0-9_:-]+)\s*=\s*["']([^"']*)["']`)
)

// InspectMetaAndSocial fetches a URL, extracts all HTML meta/OpenGraph/Twitter card tags, builds platform previews, and detects Ad/Analytics trackers.
func InspectMetaAndSocial(ctx context.Context, target string) MetaSocialReport {
	start := time.Now()
	fullURL, host, _ := normalizeURLAndHost(target)

	report := MetaSocialReport{
		TargetURL:   fullURL,
		FinalURL:    fullURL,
		Host:        host,
		CheckedAt:   time.Now().UTC().Format(time.RFC3339),
		AllMetaTags: make(map[string]string),
	}

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	bodyBytes, statusCode, finalURL := fetchWithAdaptiveStrategy(ctx, client, fullURL, 512*1024)
	report.StatusCode = statusCode
	if finalURL != "" {
		report.FinalURL = finalURL
	}
	bodyStr := string(bodyBytes)

	if bodyStr == "" {
		report.ResolvedTitle = host
		report.ResolvedDescription = "Unable to fetch HTML head metadata from target URL."
		report.ResolvedSiteName = host
		report.ResolvedThemeColor = "#5366e8"
		report.Trackers = DetectTrackersFromHTML(nil, nil)
		report.DurationMs = time.Since(start).Milliseconds()
		return report
	}

	if m := reTitleTag.FindStringSubmatch(bodyStr); len(m) > 1 {
		report.Title = strings.TrimSpace(strings.Join(strings.Fields(m[1]), " "))
	}
	if m := reHtmlLang.FindStringSubmatch(bodyStr); len(m) > 1 {
		report.Language = strings.TrimSpace(m[1])
	}

	metaTags := reMetaTag.FindAllString(bodyStr, -1)
	for _, tag := range metaTags {
		attrs := parseAttributes(tag)
		if cs, ok := attrs["charset"]; ok && cs != "" {
			report.Charset = cs
		}
		key := attrs["property"]
		if key == "" {
			key = attrs["name"]
		}
		content := strings.TrimSpace(attrs["content"])
		if key == "" || content == "" {
			continue
		}
		keyLower := strings.ToLower(key)
		report.AllMetaTags[keyLower] = content

		switch keyLower {
		case "description":
			report.Description = content
		case "theme-color":
			report.ThemeColor = content
		case "robots":
			report.RobotsMeta = content
		case "author":
			report.Author = content
		case "generator":
			report.Generator = content
		case "viewport":
			report.Viewport = content
		case "og:title":
			report.OGTitle = content
		case "og:description":
			report.OGDescription = content
		case "og:image", "og:image:url", "og:image:secure_url":
			if report.OGImage == "" {
				report.OGImage = resolveRelativeAsset(fullURL, content)
			}
		case "og:url":
			report.OGURL = content
		case "og:site_name":
			report.OGSiteName = content
		case "og:type":
			report.OGType = content
		case "og:locale":
			report.OGLocale = content
		case "twitter:card":
			report.TwitterCard = content
		case "twitter:title":
			report.TwitterTitle = content
		case "twitter:description":
			report.TwitterDescription = content
		case "twitter:image", "twitter:image:src":
			if report.TwitterImage == "" {
				report.TwitterImage = resolveRelativeAsset(fullURL, content)
			}
		case "twitter:site":
			report.TwitterSite = content
		case "twitter:creator":
			report.TwitterCreator = content
		}
	}

	linkTags := reLinkTag.FindAllString(bodyStr, -1)
	for _, ltag := range linkTags {
		attrs := parseAttributes(ltag)
		rel := strings.ToLower(attrs["rel"])
		href := strings.TrimSpace(attrs["href"])
		if href == "" {
			continue
		}
		if strings.Contains(rel, "canonical") && report.CanonicalURL == "" {
			report.CanonicalURL = resolveRelativeAsset(fullURL, href)
		}
		if strings.Contains(rel, "icon") && report.FaviconURL == "" {
			report.FaviconURL = resolveRelativeAsset(fullURL, href)
		}
	}

	if report.FaviconURL == "" {
		report.FaviconURL = "https://" + host + "/favicon.ico"
	}

	// Build resolved social preview fields
	report.ResolvedTitle = firstNonEmpty(report.OGTitle, report.TwitterTitle, report.Title, host)
	report.ResolvedDescription = firstNonEmpty(report.OGDescription, report.TwitterDescription, report.Description, "No meta description or og:description provided on this page.")
	report.ResolvedImage = firstNonEmpty(report.OGImage, report.TwitterImage)
	report.ResolvedSiteName = firstNonEmpty(report.OGSiteName, report.TwitterSite, host)
	report.ResolvedThemeColor = firstNonEmpty(report.ThemeColor, "#5865F2")

	// Run Meta Audit Checks
	score := 0
	var checks []MetaAuditCheck

	if len(report.Title) >= 15 && len(report.Title) <= 70 {
		score += 20
		checks = append(checks, MetaAuditCheck{ID: "title", Label: "HTML <title> Tag", Status: "PASS", Details: fmt.Sprintf("%d chars (Optimal 15–70 chars)", len(report.Title))})
	} else if report.Title != "" {
		score += 10
		checks = append(checks, MetaAuditCheck{ID: "title", Label: "HTML <title> Tag", Status: "WARN", Details: fmt.Sprintf("%d chars (Recommended 15–70 chars)", len(report.Title))})
	} else {
		checks = append(checks, MetaAuditCheck{ID: "title", Label: "HTML <title> Tag", Status: "MISSING", Details: "Missing <title> element"})
	}

	if len(report.Description) >= 50 && len(report.Description) <= 170 {
		score += 20
		checks = append(checks, MetaAuditCheck{ID: "desc", Label: "Meta Description", Status: "PASS", Details: fmt.Sprintf("%d chars (Optimal 50–160 chars)", len(report.Description))})
	} else if report.Description != "" {
		score += 12
		checks = append(checks, MetaAuditCheck{ID: "desc", Label: "Meta Description", Status: "WARN", Details: fmt.Sprintf("%d chars", len(report.Description))})
	} else {
		checks = append(checks, MetaAuditCheck{ID: "desc", Label: "Meta Description", Status: "MISSING", Details: "Missing meta[name=description]"})
	}

	if report.OGTitle != "" && report.OGDescription != "" {
		score += 20
		checks = append(checks, MetaAuditCheck{ID: "og_text", Label: "OpenGraph Title & Description", Status: "PASS", Details: "og:title & og:description configured for Discord/WhatsApp/FB"})
	} else {
		checks = append(checks, MetaAuditCheck{ID: "og_text", Label: "OpenGraph Title & Description", Status: "WARN", Details: "Incomplete og:title or og:description (falling back to HTML tags)"})
	}

	if report.ResolvedImage != "" {
		score += 25
		checks = append(checks, MetaAuditCheck{ID: "og_image", Label: "Social Share Image (og:image)", Status: "PASS", Details: report.ResolvedImage})
	} else {
		checks = append(checks, MetaAuditCheck{ID: "og_image", Label: "Social Share Image (og:image)", Status: "MISSING", Details: "No og:image or twitter:image — shares will render without visual banner"})
	}

	if report.TwitterCard != "" {
		score += 15
		checks = append(checks, MetaAuditCheck{ID: "twitter_card", Label: "Twitter / Discord Card Mode", Status: "PASS", Details: "twitter:card = " + report.TwitterCard})
	} else {
		checks = append(checks, MetaAuditCheck{ID: "twitter_card", Label: "Twitter / Discord Card Mode", Status: "WARN", Details: "Missing twitter:card (Discord uses summary_large_image for big banners)"})
	}

	report.SocialScore = score
	switch {
	case score >= 90:
		report.SocialGrade = "A+"
	case score >= 75:
		report.SocialGrade = "A"
	case score >= 55:
		report.SocialGrade = "B"
	default:
		report.SocialGrade = "C"
	}
	report.AuditChecks = checks

	report.Trackers = DetectTrackersFromHTML([]string{bodyStr}, nil)
	report.DurationMs = time.Since(start).Milliseconds()

	return report
}

func normalizeURLAndHost(raw string) (string, string, string) {
	clean := strings.TrimSpace(raw)
	if !strings.HasPrefix(clean, "http://") && !strings.HasPrefix(clean, "https://") {
		clean = "https://" + clean
	}
	u, err := url.Parse(clean)
	if err != nil || u.Host == "" {
		return clean, clean, "/"
	}
	host := strings.ToLower(u.Host)
	path := u.Path
	if path == "" {
		path = "/"
	}
	base := u.Scheme + "://" + host
	full := base + path
	if u.RawQuery != "" {
		full += "?" + u.RawQuery
	}
	return full, host, path
}

func resolveRelativeAsset(pageURL, assetRef string) string {
	assetRef = strings.TrimSpace(assetRef)
	if assetRef == "" {
		return ""
	}
	if strings.HasPrefix(assetRef, "http://") || strings.HasPrefix(assetRef, "https://") {
		return assetRef
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return assetRef
	}
	ref, err := url.Parse(assetRef)
	if err != nil {
		return assetRef
	}
	return base.ResolveReference(ref).String()
}

func parseAttributes(tag string) map[string]string {
	res := make(map[string]string)
	matches := reAttrPair.FindAllStringSubmatch(tag, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			res[strings.ToLower(m[1])] = m[2]
		}
	}
	return res
}

func extractPathFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return raw
	}
	return u.Path
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// fetchWithAdaptiveStrategy attempts multiple standard HTTP negotiation profiles
// (Desktop Browser navigation -> Social Card Preview Bot -> Wayback Archive Snapshot fallback)
// when an origin server or CDN edge gateway returns 403/429/503 challenge responses.
func fetchWithAdaptiveStrategy(ctx context.Context, client *http.Client, targetURL string, maxBytes int64) ([]byte, int, string) {
	profiles := []map[string]string{
		{
			"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language":           "en-US,en;q=0.9",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Upgrade-Insecure-Requests": "1",
		},
		{
			"User-Agent":      "Twitterbot/1.0",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language": "en-US,en;q=0.9",
		},
		{
			"User-Agent": "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		},
	}

	var lastBody []byte
	var lastStatus int
	var lastFinalURL string

	for _, headers := range profiles {
		reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
		if err != nil {
			cancel()
			continue
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			cancel()
			continue
		}

		lastStatus = resp.StatusCode
		if resp.Request != nil && resp.Request.URL != nil {
			lastFinalURL = resp.Request.URL.String()
		}
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
		resp.Body.Close()
		cancel()

		lastBody = bodyBytes
		bodyStr := string(bodyBytes)

		if resp.StatusCode >= 200 && resp.StatusCode < 400 && !isEdgeChallengePage(bodyStr) {
			return bodyBytes, resp.StatusCode, lastFinalURL
		}
	}

	// Fallback Profile 3: Check Wayback Machine latest snapshot if live edge blocked all direct requests
	wbCtx, wbCancel := context.WithTimeout(ctx, 4*time.Second)
	defer wbCancel()
	availEndpoint := fmt.Sprintf("https://archive.org/wayback/available?url=%s", url.QueryEscape(targetURL))
	if req, err := http.NewRequestWithContext(wbCtx, http.MethodGet, availEndpoint, nil); err == nil {
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			var avail struct {
				ArchivedSnapshots struct {
					Closest struct {
						Available bool   `json:"available"`
						URL       string `json:"url"`
					} `json:"closest"`
				} `json:"archived_snapshots"`
			}
			if json.NewDecoder(resp.Body).Decode(&avail) == nil && avail.ArchivedSnapshots.Closest.Available && avail.ArchivedSnapshots.Closest.URL != "" {
				snapURL := strings.Replace(avail.ArchivedSnapshots.Closest.URL, "http://web.archive.org/web/", "https://web.archive.org/web/", 1)
				if snapReq, err := http.NewRequestWithContext(wbCtx, http.MethodGet, snapURL, nil); err == nil {
					if snapResp, err := client.Do(snapReq); err == nil {
						defer snapResp.Body.Close()
						if b, err := io.ReadAll(io.LimitReader(snapResp.Body, maxBytes)); err == nil && len(b) > 0 {
							return b, 200, targetURL
						}
					}
				}
			}
		}
	}

	return lastBody, lastStatus, lastFinalURL
}

func isEdgeChallengePage(body string) bool {
	lower := strings.ToLower(body)
	if len(lower) > 35000 {
		return false
	}
	return strings.Contains(lower, "just a moment...") ||
		strings.Contains(lower, "attention required! | cloudflare") ||
		strings.Contains(lower, "cf-browser-verification") ||
		strings.Contains(lower, "vercel security checkpoint") ||
		strings.Contains(lower, "ddos-guard")
}

