package watchlist

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rny/scanner/internal/domain"
)

// SavedDomain represents a domain saved by the user in the Watchlist Vault.
type SavedDomain struct {
	Domain               string           `json:"domain"`
	Available            bool             `json:"available"`
	Status               string           `json:"status"`
	Valuation            domain.Valuation `json:"valuation"`
	Notes                string           `json:"notes,omitempty"`
	Tags                 []string         `json:"tags,omitempty"`
	PreviouslyRegistered *bool            `json:"previouslyRegistered,omitempty"`
	FirstSeenYear        int              `json:"firstSeenYear,omitempty"`
	SavedAt              string           `json:"savedAt"`
	LastCheckedAt        string           `json:"lastCheckedAt"`
}

// Store manages thread-safe persistent storage of saved domains.
type Store struct {
	mu       sync.RWMutex
	filePath string
	items    map[string]SavedDomain
}

// NewStore initializes a Watchlist store backed by a JSON file.
func NewStore(filePath string) *Store {
	if filePath == "" {
		filePath = filepath.Join("data", "watchlist.json")
	}
	s := &Store{
		filePath: filePath,
		items:    make(map[string]SavedDomain),
	}
	s.load()
	return s
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// Seed with 3 helpful starter examples if no file exists yet
		now := time.Now().Format(time.RFC3339)
		starters := []struct {
			dom   string
			avail bool
			note  string
			tags  []string
		}{
			{"veltrixhq.ai", true, "Short AI infrastructure brand candidate", []string{"Unclaimed Gem", "AI"}},
			{"nexoralabs.io", true, "Clean phonetic pattern for devtool launch", []string{"Shortlist", "DevTools"}},
			{"svelte.dev", false, "Reference architecture & benchmark domain", []string{"Benchmark", "Registered"}},
		}
		for _, st := range starters {
			val := domain.EvaluateDomain(st.dom)
			status := "Registered"
			if st.avail {
				status = "Available"
			}
			s.items[st.dom] = SavedDomain{
				Domain:        st.dom,
				Available:     st.avail,
				Status:        status,
				Valuation:     val,
				Notes:         st.note,
				Tags:          st.tags,
				SavedAt:       now,
				LastCheckedAt: now,
			}
		}
		_ = s.saveUnlocked()
		return
	}

	var list []SavedDomain
	if err := json.Unmarshal(data, &list); err == nil {
		for _, item := range list {
			s.items[item.Domain] = item
		}
	}
}

func (s *Store) saveUnlocked() error {
	_ = os.MkdirAll(filepath.Dir(s.filePath), 0o755)
	list := s.listUnlocked()
	raw, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, raw, 0o644)
}

func (s *Store) listUnlocked() []SavedDomain {
	var out []SavedDomain
	for _, v := range s.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Available != out[j].Available {
			return out[i].Available
		}
		return out[i].Valuation.Score > out[j].Valuation.Score
	})
	return out
}

// List returns all saved domains sorted by availability and valuation score.
func (s *Store) List() []SavedDomain {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listUnlocked()
}

// Upsert adds or updates a domain in the watchlist.
func (s *Store) Upsert(rawDomain string, available *bool, notes string, tags []string) SavedDomain {
	clean := domain.CleanDomainName(rawDomain)
	now := time.Now().Format(time.RFC3339)
	val := domain.EvaluateDomain(clean)

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.items[clean]
	isAvail := true
	if available != nil {
		isAvail = *available
	} else if exists {
		isAvail = existing.Available
	}

	status := "Registered"
	if isAvail {
		status = "Available"
	}

	savedAt := now
	if exists && existing.SavedAt != "" {
		savedAt = existing.SavedAt
	}
	if notes == "" && exists {
		notes = existing.Notes
	}
	if len(tags) == 0 && exists {
		tags = existing.Tags
	}
	if len(tags) == 0 {
		if isAvail {
			tags = []string{"Unclaimed"}
		} else {
			tags = []string{"Watch"}
		}
	}

	entry := SavedDomain{
		Domain:               clean,
		Available:            isAvail,
		Status:               status,
		Valuation:            val,
		Notes:                strings.TrimSpace(notes),
		Tags:                 tags,
		PreviouslyRegistered: existing.PreviouslyRegistered,
		FirstSeenYear:        existing.FirstSeenYear,
		SavedAt:              savedAt,
		LastCheckedAt:        now,
	}

	s.items[clean] = entry
	_ = s.saveUnlocked()
	return entry
}

// Remove deletes a domain from the watchlist.
func (s *Store) Remove(rawDomain string) bool {
	clean := domain.CleanDomainName(rawDomain)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[clean]; !ok {
		return false
	}
	delete(s.items, clean)
	_ = s.saveUnlocked()
	return true
}

// RecheckAllParallel concurrently verifies registration & archive status for all saved domains.
func (s *Store) RecheckAllParallel(ctx context.Context) []SavedDomain {
	current := s.List()
	if len(current) == 0 {
		return current
	}

	var wg sync.WaitGroup
	updated := make([]SavedDomain, len(current))

	for i, item := range current {
		wg.Add(1)
		go func(idx int, sd SavedDomain) {
			defer wg.Done()
			inq := domain.InspectDomain(ctx, sd.Domain)
			hist := domain.CheckDomainHistory(ctx, sd.Domain)
			prevReg := hist.PreviouslyRegistered

			sd.Available = inq.Available
			sd.Status = inq.StatusSummary
			sd.Valuation = inq.Valuation
			sd.PreviouslyRegistered = &prevReg
			sd.FirstSeenYear = hist.FirstSeenYear
			sd.LastCheckedAt = time.Now().Format(time.RFC3339)
			updated[idx] = sd
		}(i, item)
	}

	wg.Wait()

	s.mu.Lock()
	for _, u := range updated {
		s.items[u.Domain] = u
	}
	_ = s.saveUnlocked()
	out := s.listUnlocked()
	s.mu.Unlock()

	return out
}
