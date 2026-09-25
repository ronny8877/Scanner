package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var validLabelRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
var validTLDRegex = regexp.MustCompile(`^[a-z]{2,24}$`)

// ValidationResult describes whether a domain or URL input is valid and offers a typo correction if possible.
type ValidationResult struct {
	Valid      bool   `json:"valid"`
	Normalized string `json:"normalized"`
	Host       string `json:"host"`
	Error      string `json:"error,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidateTarget checks whether a raw domain or URL (e.g. "xyz,com" or "svelte.dev/docs") has valid RFC hostname syntax.
func ValidateTarget(raw string) ValidationResult {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ValidationResult{
			Valid: false,
			Error: "Please enter a domain name (e.g. svelte.dev or cloudflare.com).",
		}
	}

	// Check for common comma-instead-of-dot typo (e.g. "xyz,com")
	if strings.Contains(trimmed, ",") {
		suggested := strings.ReplaceAll(trimmed, ",", ".")
		suggested = strings.ReplaceAll(suggested, " ", "")
		return ValidationResult{
			Valid:      false,
			Error:      fmt.Sprintf("Invalid domain syntax '%s': contains a comma ',' instead of a dot '.'.", trimmed),
			Suggestion: suggested,
		}
	}

	if strings.Contains(trimmed, " ") {
		suggested := strings.ReplaceAll(trimmed, " ", "")
		return ValidationResult{
			Valid:      false,
			Error:      fmt.Sprintf("Invalid domain syntax '%s': domain names cannot contain spaces.", trimmed),
			Suggestion: suggested,
		}
	}

	clean := strings.ToLower(trimmed)
	clean = strings.TrimPrefix(clean, "https://")
	clean = strings.TrimPrefix(clean, "http://")
	clean = strings.TrimPrefix(clean, "https//")
	clean = strings.TrimPrefix(clean, "http//")

	slashIdx := strings.Index(clean, "/")
	host := clean
	pathPart := ""
	if slashIdx != -1 {
		host = clean[:slashIdx]
		pathPart = clean[slashIdx:]
	}

	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	if !strings.Contains(host, ".") {
		return ValidationResult{
			Valid:      false,
			Error:      fmt.Sprintf("'%s' is missing a Top-Level Domain extension (like .com, .ai, .dev).", host),
			Suggestion: host + ".com",
		}
	}

	if strings.Contains(host, "..") {
		suggested := strings.ReplaceAll(host, "..", ".") + pathPart
		return ValidationResult{
			Valid:      false,
			Error:      fmt.Sprintf("Invalid domain '%s': contains consecutive dots '..'.", host),
			Suggestion: suggested,
		}
	}

	labels := strings.Split(host, ".")
	tld := labels[len(labels)-1]
	if !validTLDRegex.MatchString(tld) {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("Invalid TLD extension '.%s' in '%s'. Extensions must be 2–24 letters.", tld, host),
		}
	}

	for _, label := range labels {
		if !validLabelRegex.MatchString(label) {
			return ValidationResult{
				Valid: false,
				Error: fmt.Sprintf("Invalid domain segment '%s' in '%s'. Only letters, digits, and interior hyphens are valid.", label, host),
			}
		}
	}

	return ValidationResult{
		Valid:      true,
		Normalized: host + pathPart,
		Host:       host,
	}
}
