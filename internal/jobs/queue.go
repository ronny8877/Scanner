package jobs

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rny/scanner/internal/crawler"
	"github.com/rny/scanner/internal/domain"
	"github.com/rny/scanner/internal/recon"
)

// JobStatus represents the lifecycle state of a queued scan task.
type JobStatus string

const (
	StatusQueued    JobStatus = "QUEUED"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

// Job represents a single tracked enterprise job in the queue.
type Job struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"` // "scan" | "inspect" | "history" | "recon" | "crawl" | "parallel_suite"
	Title         string      `json:"title"`
	Target        string      `json:"target"`
	Status        JobStatus   `json:"status"`
	Progress      int         `json:"progress"`      // 0 - 100
	Phase         string      `json:"phase"`         // Current execution step description
	Workers       int         `json:"workers"`       // Concurrent workers assigned
	ResultSummary string      `json:"resultSummary"` // Concise outcome pill text
	CreatedAt     string      `json:"createdAt"`
	DurationMs    int64       `json:"durationMs"`
	Result        interface{} `json:"result,omitempty"`
}

// ParallelSuiteResult holds the combined output of running all 4 engines concurrently on a target.
type ParallelSuiteResult struct {
	Domain  string               `json:"domain"`
	Inquiry domain.DomainInquiry `json:"inquiry"`
	History domain.HistoryReport `json:"history"`
	Recon   recon.ReconReport    `json:"recon"`
	Crawl   crawler.CrawlReport  `json:"crawl"`
}

// Manager coordinates concurrent job execution and telemetry.
type Manager struct {
	mu      sync.RWMutex
	counter uint64
	jobs    map[string]*Job
}

// NewManager creates a new enterprise Job Manager.
func NewManager() *Manager {
	return &Manager{
		jobs: make(map[string]*Job),
	}
}

// CreateJob registers a new job in QUEUED state and transitions it to RUNNING.
func (m *Manager) CreateJob(jobType, title, target string, workers int) *Job {
	seq := atomic.AddUint64(&m.counter, 1)
	id := fmt.Sprintf("job-%04d", seq)
	j := &Job{
		ID:        id,
		Type:      jobType,
		Title:     title,
		Target:    target,
		Status:    StatusRunning,
		Progress:  12,
		Phase:     "Dispatching concurrent worker pool…",
		Workers:   workers,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	m.mu.Lock()
	m.jobs[id] = j
	m.mu.Unlock()
	return j
}

// UpdateProgress updates the live progress percentage and phase description of a running job.
func (m *Manager) UpdateProgress(id string, progress int, phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok {
		j.Progress = progress
		j.Phase = phase
	}
}

// CompleteJob marks a job as COMPLETED with its summary and result payload.
func (m *Manager) CompleteJob(id, summary string, durationMs int64, result interface{}) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil
	}
	j.Status = StatusCompleted
	j.Progress = 100
	j.Phase = "Completed all parallel workers"
	j.ResultSummary = summary
	j.DurationMs = durationMs
	j.Result = result
	return j
}

// List returns the most recent jobs ordered newest first.
func (m *Manager) List() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []*Job
	for _, j := range m.jobs {
		copyJob := *j
		// Keep list lightweight while preserving metadata
		out = append(out, &copyJob)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID > out[j].ID
	})
	if len(out) > 35 {
		out = out[:35]
	}
	return out
}

// Get returns a single job including its full Result payload.
func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil, false
	}
	copyJob := *j
	return &copyJob, true
}

// RunParallelSuite executes RDAP Inquiry, Past History (Wayback+CT), Port/TLS Recon, and Site Cartography
// all in parallel on a target domain while updating job telemetry.
func (m *Manager) RunParallelSuite(ctx context.Context, rawTarget string) (*Job, ParallelSuiteResult) {
	start := time.Now()
	clean := domain.CleanDomainName(rawTarget)
	job := m.CreateJob("parallel_suite", "Full Parallel Surface & History Suite", clean, 32)

	var (
		wg          sync.WaitGroup
		doneCount   int32
		inquiryRes  domain.DomainInquiry
		historyRes  domain.HistoryReport
		reconRes    recon.ReconReport
		crawlRes    crawler.CrawlReport
	)

	advanceStep := func(stepLabel string) {
		c := atomic.AddInt32(&doneCount, 1)
		pct := int(c) * 25
		if pct >= 100 {
			pct = 95
		}
		m.UpdateProgress(job.ID, pct, fmt.Sprintf("Finished %s (%d/4 pipelines complete)", stepLabel, c))
	}

	wg.Add(4)

	go func() {
		defer wg.Done()
		inquiryRes = domain.InspectDomain(ctx, clean)
		advanceStep("RDAP & DNS Dossier")
	}()

	go func() {
		defer wg.Done()
		historyRes = domain.CheckDomainHistory(ctx, clean)
		advanceStep("Wayback & CT Archive History")
	}()

	go func() {
		defer wg.Done()
		reconRes = recon.RunRecon(ctx, clean)
		advanceStep("Port, TLS & Security Audit")
	}()

	go func() {
		defer wg.Done()
		crawlRes = crawler.CrawlSite(ctx, crawler.CrawlOptions{
			TargetURL: clean,
			MaxPages:  15,
			MaxDepth:  2,
		})
		advanceStep("Site Cartography Tree")
	}()

	wg.Wait()

	suite := ParallelSuiteResult{
		Domain:  clean,
		Inquiry: inquiryRes,
		History: historyRes,
		Recon:   reconRes,
		Crawl:   crawlRes,
	}

	summary := fmt.Sprintf(
		"%s · %d open ports · %d Wayback snaps · %d routes",
		inquiryRes.StatusSummary,
		reconRes.OpenPortsCount,
		historyRes.WaybackSnapshots,
		crawlRes.PagesCrawled,
	)

	completedJob := m.CompleteJob(job.ID, summary, time.Since(start).Milliseconds(), suite)
	return completedJob, suite
}
