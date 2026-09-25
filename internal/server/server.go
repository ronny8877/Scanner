package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rny/scanner/internal/crawler"
	"github.com/rny/scanner/internal/domain"
	"github.com/rny/scanner/internal/jobs"
	"github.com/rny/scanner/internal/recon"
	"github.com/rny/scanner/internal/traffic"
	"github.com/rny/scanner/internal/watchlist"
	"github.com/rny/scanner/internal/webintel"
	"github.com/rny/scanner/web"
)

// Server wraps the HTTP API server, Enterprise Job Manager, and Watchlist Store.
type Server struct {
	Addr      string
	Jobs      *jobs.Manager
	Watchlist *watchlist.Store
}

// New creates a new API server instance with initialized Job Queue and Watchlist Vault.
func New(addr string) *Server {
	if addr == "" {
		addr = ":8080"
	}
	return &Server{
		Addr:      addr,
		Jobs:      jobs.NewManager(),
		Watchlist: watchlist.NewStore("data/watchlist.json"),
	}
}

// Start runs the HTTP server until context cancellation or error.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.withCORS(s.handleHealth))
	mux.HandleFunc("/api/dictionary", s.withCORS(s.handleDictionary))
	mux.HandleFunc("/api/scan", s.withCORS(s.handleScan))
	mux.HandleFunc("/api/inspect", s.withCORS(s.handleInspect))
	mux.HandleFunc("/api/history", s.withCORS(s.handleHistory))
	mux.HandleFunc("/api/recon", s.withCORS(s.handleRecon))
	mux.HandleFunc("/api/crawl", s.withCORS(s.handleCrawl))
	mux.HandleFunc("/api/robots", s.withCORS(s.handleRobots))
	mux.HandleFunc("/api/meta", s.withCORS(s.handleMeta))
	mux.HandleFunc("/api/traffic", s.withCORS(s.handleTraffic))
	mux.HandleFunc("/api/parallel-suite", s.withCORS(s.handleParallelSuite))
	mux.HandleFunc("/api/jobs", s.withCORS(s.handleJobs))
	mux.HandleFunc("/api/jobs/cancel", s.withCORS(s.handleJobCancel))
	mux.HandleFunc("/api/watchlist", s.withCORS(s.handleWatchlist))
	mux.HandleFunc("/api/watchlist/recheck", s.withCORS(s.handleWatchlistRecheck))

	// Serve embedded Svelte 5 Web Studio directly from the single binary at /
	if distFS, err := web.DistFS(); err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	srv := &http.Server{
		Addr:         s.Addr,
		Handler:      mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 65 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *Server) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "ok",
		"service":        "scanner-api",
		"version":        "2.1.0",
		"activeJobs":     len(s.Jobs.List()),
		"savedWatchlist": len(s.Watchlist.List()),
		"modes":          []string{"scan", "inspect", "history", "recon", "crawl", "parallel_suite", "watchlist"},
	})
}

func (s *Server) handleDictionary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"packs": domain.BuiltinDictionaryPacks,
	})
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	opts := domain.ScanOptions{
		Keywords:      []string{"veltrix", "nova"},
		TLDs:          []string{"com", "ai", "io", "dev", "co", "app"},
		Mutations:     true,
		OnlyAvailable: false,
		MinScore:      0,
		Concurrency:   18,
		MaxResults:    64,
	}

	if r.Method == http.MethodPost {
		_ = json.NewDecoder(r.Body).Decode(&opts)
	} else {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q != "" {
			opts.Keywords = splitCSV(q)
		}
		if tlds := r.URL.Query().Get("tlds"); tlds != "" {
			opts.TLDs = splitCSV(tlds)
		}
		if dict := r.URL.Query().Get("dictionaryPack"); dict != "" {
			opts.DictionaryPack = dict
		}
		if r.URL.Query().Get("onlyAvailable") == "true" {
			opts.OnlyAvailable = true
		}
		if r.URL.Query().Get("mutations") == "false" {
			opts.Mutations = false
		}
		if minStr := r.URL.Query().Get("minScore"); minStr != "" {
			if v, err := strconv.Atoi(minStr); err == nil {
				opts.MinScore = v
			}
		}
	}

	targetDesc := strings.Join(opts.Keywords, ", ")
	if opts.DictionaryPack != "" {
		targetDesc += " + [" + opts.DictionaryPack + "]"
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 28*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "scan", "Bulk Availability & Valuation Scan", targetDesc, 18)
	defer jobCancel()

	s.Jobs.UpdateProgress(job.ID, 45, fmt.Sprintf("Resolving %d TLDs across 18 parallel workers…", len(opts.TLDs)))

	report := domain.ScanDomains(jobCtx, opts)
	summary := fmt.Sprintf("%d unclaimed · %d prime gems · %d checked", report.AvailableCount, report.HighValueCount, report.TotalChecked)
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleInspect(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("domain"))
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Domain string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		target = body.Domain
	}
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain parameter is required"})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 14*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "inspect", "RDAP Registration & DNS Dossier", target, 6)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 55, "Querying ICANN RDAP gateway & DNS NS/MX/TXT/DMARC…")

	result := domain.InspectDomain(jobCtx, target)
	summary := fmt.Sprintf("%s · Score %d/100", result.StatusSummary, result.Valuation.Score)
	if result.DomainAge != "" {
		summary = fmt.Sprintf("%s · Age %s", result.StatusSummary, result.DomainAge)
	}
	s.Jobs.CompleteJob(job.ID, summary, result.CheckLatencyMs, result)

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("domain"))
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Domain string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		target = body.Domain
	}
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain parameter is required"})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 16*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "history", "Past Registration & Archive History", target, 6)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 50, "Querying Wayback CDX + Availability API + RDAP + CT logs…")

	report := domain.CheckDomainHistory(jobCtx, target)
	summary := "Clean Virgin History"
	if report.PreviouslyRegistered {
		summary = fmt.Sprintf("Active/Past Record (%d–%d · %d snaps)", report.FirstSeenYear, report.LastSeenYear, report.WaybackSnapshots)
	}
	s.Jobs.CompleteJob(job.ID, summary, report.CheckLatencyMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleRecon(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("domain"))
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Domain string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		target = body.Domain
	}
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain parameter is required"})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 16*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "recon", "Parallel Port, TLS & Security Recon", target, 28)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 40, "Probing 14 TCP ports, TLS 1.3 handshake, headers & subdomains…")

	report := recon.RunRecon(jobCtx, target)
	summary := fmt.Sprintf("%d open ports · Grade %s · %d subdomains", report.OpenPortsCount, report.SecurityGrade, len(report.Subdomains))
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleCrawl(w http.ResponseWriter, r *http.Request) {
	opts := crawler.CrawlOptions{
		TargetURL: r.URL.Query().Get("url"),
		MaxPages:  20,
		MaxDepth:  2,
	}
	if r.Method == http.MethodPost {
		_ = json.NewDecoder(r.Body).Decode(&opts)
	} else {
		if mp := r.URL.Query().Get("maxPages"); mp != "" {
			if v, err := strconv.Atoi(mp); err == nil {
				opts.MaxPages = v
			}
		}
		if md := r.URL.Query().Get("maxDepth"); md != "" {
			if v, err := strconv.Atoi(md); err == nil {
				opts.MaxDepth = v
			}
		}
	}

	if strings.TrimSpace(opts.TargetURL) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "targetUrl is required"})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "crawl", "Site Structure & Link Graph Crawl", opts.TargetURL, 8)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 50, fmt.Sprintf("Crawling up to %d pages (depth %d)…", opts.MaxPages, opts.MaxDepth))

	report := crawler.CrawlSite(jobCtx, opts)
	summary := fmt.Sprintf("%d pages crawled · %d internal routes", report.PagesCrawled, report.TotalLinks)
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleRobots(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	if target == "" {
		target = strings.TrimSpace(r.URL.Query().Get("domain"))
	}
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Target string `json:"target"`
			Domain string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Target != "" {
			target = body.Target
		} else {
			target = body.Domain
		}
	}
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target is required"})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 22*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "robots", "Robots.txt & Sitemap.xml Audit", target, 6)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 50, "Parsing robots.txt User-agent rules & XML sitemaps…")

	report := webintel.InspectRobotsAndSitemap(jobCtx, target)
	summary := fmt.Sprintf("Robots: %v (%d rules) · Sitemap: %d URLs", report.RobotsFound, report.TotalDisallowCount+report.TotalAllowCount, report.TotalUrlsCount)
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	if target == "" {
		target = strings.TrimSpace(r.URL.Query().Get("url"))
	}
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Target string `json:"target"`
			URL    string `json:"url"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Target != "" {
			target = body.Target
		} else {
			target = body.URL
		}
	}
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target URL is required"})
		return
	}

	val := domain.ValidateTarget(target)
	if !val.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":      val.Error,
			"suggestion": val.Suggestion,
		})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "meta", "Social Meta & Ad/Tracker Inspection", val.Normalized, 4)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 55, "Extracting OpenGraph/Twitter cards & scanning Ad/Tracker scripts…")

	report := webintel.InspectMetaAndSocial(jobCtx, val.Normalized)
	summary := fmt.Sprintf("Social Grade %s · Trackers Grade %s (%d detected)", report.SocialGrade, report.Trackers.PrivacyGrade, report.Trackers.TotalDetected)
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleTraffic(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("domain"))
	if target == "" {
		target = strings.TrimSpace(r.URL.Query().Get("target"))
	}
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Domain string `json:"domain"`
			Target string `json:"target"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Domain != "" {
			target = body.Domain
		} else {
			target = body.Target
		}
	}
	val := domain.ValidateTarget(target)
	if !val.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":      val.Error,
			"suggestion": val.Suggestion,
		})
		return
	}

	baseCtx, baseCancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer baseCancel()

	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "traffic", "Traffic & Popularity Intelligence", val.Host, 6)
	defer jobCancel()
	s.Jobs.UpdateProgress(job.ID, 50, "Correlating Tranco Top-1M + Cloudflare Radar + Index Footprint…")

	report := traffic.EstimateDomainTraffic(jobCtx, val.Host)
	summary := fmt.Sprintf("%s · %s · %s", report.DomainPopularity, report.EstimatedTrafficRange, report.PopularityTrend)
	s.Jobs.CompleteJob(job.ID, summary, report.DurationMs, report)

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleParallelSuite(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("domain"))
	if target == "" && r.Method == http.MethodPost {
		var body struct {
			Domain string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		target = body.Domain
	}
	val := domain.ValidateTarget(target)
	if !val.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":      val.Error,
			"suggestion": val.Suggestion,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()

	_, suite := s.Jobs.RunParallelSuite(ctx, val.Host)
	writeJSON(w, http.StatusOK, suite)
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id != "" {
		if j, ok := s.Jobs.Get(id); ok {
			writeJSON(w, http.StatusOK, j)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs": s.Jobs.List(),
	})
}

func (s *Server) handleJobCancel(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" && r.Method == http.MethodPost {
		var body struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		id = body.ID
	}

	if id == "all" || id == "" {
		canceledCount := s.Jobs.CancelAllRunning()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"canceled": canceledCount,
			"jobs":     s.Jobs.List(),
		})
		return
	}

	ok := s.Jobs.CancelJob(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"canceled": ok,
		"jobs":     s.Jobs.List(),
	})
}

func (s *Server) handleWatchlist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": s.Watchlist.List(),
		})
	case http.MethodPost:
		var body struct {
			Domain    string   `json:"domain"`
			Available *bool    `json:"available"`
			Notes     string   `json:"notes"`
			Tags      []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Domain) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid domain is required"})
			return
		}
		saved := s.Watchlist.Upsert(body.Domain, body.Available, body.Notes, body.Tags)
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		dom := strings.TrimSpace(r.URL.Query().Get("domain"))
		if dom == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain query parameter is required"})
			return
		}
		s.Watchlist.Remove(dom)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": s.Watchlist.List(),
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleWatchlistRecheck(w http.ResponseWriter, r *http.Request) {
	baseCtx, baseCancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer baseCancel()

	items := s.Watchlist.List()
	job, jobCtx, jobCancel := s.Jobs.CreateJobWithCancel(baseCtx, "watchlist_recheck", "Parallel Watchlist Vault Verification", fmt.Sprintf("%d saved domains", len(items)), len(items)*2)
	defer jobCancel()

	t0 := time.Now()
	updated := s.Watchlist.RecheckAllParallel(jobCtx)
	s.Jobs.CompleteJob(job.ID, fmt.Sprintf("Re-verified %d saved domains", len(updated)), time.Since(t0).Milliseconds(), updated)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": updated,
	})
}

func splitCSV(s string) []string {
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []string
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		fmt.Printf("json encode error: %v\n", err)
	}
}
