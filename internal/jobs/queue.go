package jobs

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rny/scanner/internal/crawler"
	"github.com/rny/scanner/internal/domain"
	"github.com/rny/scanner/internal/recon"
	"github.com/rny/scanner/internal/traffic"
	"github.com/rny/scanner/internal/webintel"
)

// JobStatus represents the lifecycle state of a queued scan task.
type JobStatus string

const (
	StatusQueued    JobStatus = "QUEUED"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusCanceled  JobStatus = "CANCELED"
	StatusFailed    JobStatus = "FAILED"
)

// Job represents a single tracked enterprise job in the queue.
type Job struct {
	ID            string             `json:"id"`
	Type          string             `json:"type"`
	Title         string             `json:"title"`
	Target        string             `json:"target"`
	Status        JobStatus          `json:"status"`
	Progress      int                `json:"progress"`
	Phase         string             `json:"phase"`
	Workers       int                `json:"workers"`
	ResultSummary string             `json:"resultSummary"`
	CreatedAt     string             `json:"createdAt"`
	DurationMs    int64              `json:"durationMs"`
	Result        interface{}        `json:"result,omitempty"`
	cancelFunc    context.CancelFunc `json:"-"`
}

// ParallelSuiteResult holds the complete Executive Domain Dossier compiled across all 7 engines in parallel.
type ParallelSuiteResult struct {
	ReportID    string                       `json:"reportId"`
	Domain      string                       `json:"domain"`
	GeneratedAt string                       `json:"generatedAt"`
	DurationMs  int64                        `json:"durationMs"`
	Inquiry     domain.DomainInquiry         `json:"inquiry"`
	History     domain.HistoryReport         `json:"history"`
	Traffic     traffic.TrafficReport        `json:"traffic"`
	Recon       recon.ReconReport            `json:"recon"`
	Crawl       crawler.CrawlReport          `json:"crawl"`
	Robots      webintel.RobotsSitemapReport `json:"robots"`
	Meta        webintel.MetaSocialReport    `json:"meta"`
}

// Manager coordinates concurrent job execution, cancellation, and telemetry.
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

// CreateJobWithCancel registers a new running job with an attached context.CancelFunc so it can be aborted mid-flight.
func (m *Manager) CreateJobWithCancel(parentCtx context.Context, jobType, title, target string, workers int) (*Job, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)
	seq := atomic.AddUint64(&m.counter, 1)
	id := fmt.Sprintf("job-%04d", seq)
	j := &Job{
		ID:         id,
		Type:       jobType,
		Title:      title,
		Target:     target,
		Status:     StatusRunning,
		Progress:   15,
		Phase:      "Dispatching concurrent worker pool…",
		Workers:    workers,
		CreatedAt:  time.Now().Format(time.RFC3339),
		cancelFunc: cancel,
	}

	m.mu.Lock()
	m.jobs[id] = j
	m.mu.Unlock()
	return j, ctx, cancel
}

// CreateJob registers a new job in RUNNING state.
func (m *Manager) CreateJob(jobType, title, target string, workers int) *Job {
	j, _, _ := m.CreateJobWithCancel(context.Background(), jobType, title, target, workers)
	return j
}

// CancelJob aborts a running job by invoking its context.CancelFunc and marking it CANCELED.
func (m *Manager) CancelJob(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	j, ok := m.jobs[id]
	if !ok {
		return false
	}
	if j.Status == StatusRunning || j.Status == StatusQueued {
		if j.cancelFunc != nil {
			j.cancelFunc()
		}
		j.Status = StatusCanceled
		j.Phase = "Canceled by user"
		j.ResultSummary = "Aborted mid-flight by user"
		return true
	}
	return false
}

// CancelAllRunning aborts all currently running jobs.
func (m *Manager) CancelAllRunning() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, j := range m.jobs {
		if j.Status == StatusRunning || j.Status == StatusQueued {
			if j.cancelFunc != nil {
				j.cancelFunc()
			}
			j.Status = StatusCanceled
			j.Phase = "Canceled by user"
			j.ResultSummary = "Aborted mid-flight by user"
			count++
		}
	}
	return count
}

// UpdateProgress updates the live progress percentage and phase description of a running job.
func (m *Manager) UpdateProgress(id string, progress int, phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok && j.Status == StatusRunning {
		j.Progress = progress
		j.Phase = phase
	}
}

// CompleteJob marks a job as COMPLETED (unless it was already CANCELED).
func (m *Manager) CompleteJob(id, summary string, durationMs int64, result interface{}) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil
	}
	if j.Status == StatusCanceled {
		j.DurationMs = durationMs
		return j
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
		out = append(out, &copyJob)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID > out[j].ID
	})
	if len(out) > 40 {
		out = out[:40]
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
func (m *Manager) RunParallelSuite(parentCtx context.Context, rawTarget string) (*Job, ParallelSuiteResult) {
	start := time.Now()
	clean := domain.CleanDomainName(rawTarget)
	job, ctx, cancel := m.CreateJobWithCancel(parentCtx, "parallel_suite", "Executive Domain Intelligence Report", clean, 42)
	defer cancel()

	var (
		wg         sync.WaitGroup
		doneCount  int32
		inquiryRes domain.DomainInquiry
		historyRes domain.HistoryReport
		trafficRes traffic.TrafficReport
		reconRes   recon.ReconReport
		crawlRes   crawler.CrawlReport
		robotsRes  webintel.RobotsSitemapReport
		metaRes    webintel.MetaSocialReport
	)

	advanceStep := func(stepLabel string) {
		c := atomic.AddInt32(&doneCount, 1)
		pct := int(c) * 14
		if pct >= 100 {
			pct = 96
		}
		m.UpdateProgress(job.ID, pct, fmt.Sprintf("Completed %s (%d/7 report sections)", stepLabel, c))
	}

	wg.Add(7)

	go func() {
		defer wg.Done()
		inquiryRes = domain.InspectDomain(ctx, clean)
		advanceStep("RDAP, WHOIS & Valuation")
	}()

	go func() {
		defer wg.Done()
		historyRes = domain.CheckDomainHistory(ctx, clean)
		advanceStep("Wayback & CT Archive History")
	}()

	go func() {
		defer wg.Done()
		trafficRes = traffic.EstimateDomainTraffic(ctx, clean)
		advanceStep("Tranco & Radar Traffic Intelligence")
	}()

	go func() {
		defer wg.Done()
		reconRes = recon.RunRecon(ctx, clean)
		advanceStep("Port, TLS & Subdomain Recon")
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

	go func() {
		defer wg.Done()
		robotsRes = webintel.InspectRobotsAndSitemap(ctx, clean)
		advanceStep("Robots.txt & Sitemap.xml Audit")
	}()

	go func() {
		defer wg.Done()
		metaRes = webintel.InspectMetaAndSocial(ctx, clean)
		advanceStep("Social Cards & Tracker Telemetry")
	}()

	wg.Wait()

	duration := time.Since(start).Milliseconds()
	suite := ParallelSuiteResult{
		ReportID:    fmt.Sprintf("REP-%s-%d", strings.ToUpper(strings.Split(clean, ".")[0]), time.Now().Unix()%10000),
		Domain:      clean,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		DurationMs:  duration,
		Inquiry:     inquiryRes,
		History:     historyRes,
		Traffic:     trafficRes,
		Recon:       reconRes,
		Crawl:       crawlRes,
		Robots:      robotsRes,
		Meta:        metaRes,
	}

	summary := fmt.Sprintf(
		"%s · %s · %s · Grade %s",
		inquiryRes.StatusSummary,
		inquiryRes.Valuation.EstimatedDisplay,
		trafficRes.EstimatedTrafficRange,
		reconRes.SecurityGrade,
	)

	completedJob := m.CompleteJob(job.ID, summary, duration, suite)
	return completedJob, suite
}
