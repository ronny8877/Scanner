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
)

// Server wraps the HTTP API server for the Svelte UI.
type Server struct {
	Addr string
}

// New creates a new API server instance.
func New(addr string) *Server {
	if addr == "" {
		addr = ":8080"
	}
	return &Server{Addr: addr}
}

// Start runs the HTTP server until context cancellation or error.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.withCORS(s.handleHealth))
	mux.HandleFunc("/api/scan", s.withCORS(s.handleScan))
	mux.HandleFunc("/api/inspect", s.withCORS(s.handleInspect))
	mux.HandleFunc("/api/crawl", s.withCORS(s.handleCrawl))

	srv := &http.Server{
		Addr:         s.Addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *Server) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
		"status":  "ok",
		"service": "scanner-api",
		"version": "1.0.0",
		"modes":   []string{"scan", "inspect", "crawl"},
	})
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	opts := domain.ScanOptions{
		Keywords:      []string{"nova", "pulse"},
		TLDs:          []string{"com", "ai", "io", "dev", "co", "app"},
		Mutations:     true,
		OnlyAvailable: false,
		MinScore:      0,
		Concurrency:   14,
		MaxResults:    48,
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

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	report := domain.ScanDomains(ctx, opts)
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

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	result := domain.InspectDomain(ctx, target)
	writeJSON(w, http.StatusOK, result)
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

	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()

	report := crawler.CrawlSite(ctx, opts)
	writeJSON(w, http.StatusOK, report)
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
