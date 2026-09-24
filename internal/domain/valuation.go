package domain

import (
	"fmt"
	"strings"
	"unicode"
)

// Valuation holds the calculated value metrics for a domain name.
type Valuation struct {
	Score            int      `json:"score"`            // 0 - 100 composite score
	Tier             string   `json:"tier"`             // "Ultra Premium", "High Value", "Brandable", "Standard"
	EstimatedMinUSD  int      `json:"estimatedMinUsd"`  // Estimated aftermarket minimum USD
	EstimatedMaxUSD  int      `json:"estimatedMaxUsd"`  // Estimated aftermarket maximum USD
	EstimatedDisplay string   `json:"estimatedDisplay"` // Formatted string e.g. "$1,200 - $3,500"
	LengthScore      int      `json:"lengthScore"`      // 0 - 30
	TLDScore         int      `json:"tldScore"`         // 0 - 25
	PhoneticScore    int      `json:"phoneticScore"`    // 0 - 25
	KeywordScore     int      `json:"keywordScore"`     // 0 - 20
	Highlights       []string `json:"highlights"`       // Reasons contributing to value
}

var tldWeights = map[string]int{
	"com": 25,
	"ai":  24,
	"io":  21,
	"dev": 19,
	"co":  18,
	"app": 18,
	"net": 15,
	"org": 14,
	"xyz": 12,
	"sh":  15,
}

var highValueAffixes = []string{
	"cloud", "data", "pay", "flow", "sync", "pulse", "core", "hq",
	"labs", "base", "stack", "grid", "hub", "scale", "nexus", "mint",
	"nova", "apex", "flux", "orbit", "forge", "spark", "vector", "agent",
}

// EvaluateDomain computes a comprehensive valuation for a full domain name (e.g. "novapulse.io").
func EvaluateDomain(fullDomain string) Valuation {
	fullDomain = strings.ToLower(strings.TrimSpace(fullDomain))
	parts := strings.Split(fullDomain, ".")
	name := parts[0]
	tld := "com"
	if len(parts) > 1 {
		tld = parts[len(parts)-1]
	}

	var highlights []string

	// 1. Length score (max 30)
	length := len(name)
	lengthScore := 0
	switch {
	case length <= 4:
		lengthScore = 30
		highlights = append(highlights, fmt.Sprintf("Ultra-short %d-letter root", length))
	case length <= 6:
		lengthScore = 26
		highlights = append(highlights, fmt.Sprintf("Compact %d-letter brandable length", length))
	case length <= 8:
		lengthScore = 22
		highlights = append(highlights, "Ideal 7-8 char memorable length")
	case length <= 10:
		lengthScore = 16
	case length <= 13:
		lengthScore = 10
	default:
		lengthScore = 5
	}

	// 2. TLD score (max 25)
	tldScore, ok := tldWeights[tld]
	if !ok {
		tldScore = 10
	}
	if tld == "com" {
		highlights = append(highlights, "Flagship .com extension")
	} else if tld == "ai" || tld == "io" || tld == "dev" {
		highlights = append(highlights, fmt.Sprintf("High-demand tech TLD (.%s)", tld))
	}

	// 3. Phonetic / Cleanliness score (max 25)
	phoneticScore := analyzePhonetics(name)
	if !strings.Contains(name, "-") && !hasDigits(name) {
		phoneticScore += 5
		if phoneticScore > 25 {
			phoneticScore = 25
		}
		if phoneticScore >= 20 {
			highlights = append(highlights, "Clean phonetic cadence (no hyphens/digits)")
		}
	} else {
		highlights = append(highlights, "Contains hyphen or numeric characters")
	}

	// 4. Commercial / Keyword relevance (max 20)
	keywordScore := 10
	for _, kw := range highValueAffixes {
		if strings.Contains(name, kw) {
			keywordScore += 5
			highlights = append(highlights, fmt.Sprintf("Strong commercial keyword (%s)", kw))
			break
		}
	}
	if keywordScore > 20 {
		keywordScore = 20
	}

	totalScore := lengthScore + tldScore + phoneticScore + keywordScore
	if totalScore > 100 {
		totalScore = 100
	}

	tier, minUSD, maxUSD := scoreToPriceRange(totalScore, tld, length)

	return Valuation{
		Score:            totalScore,
		Tier:             tier,
		EstimatedMinUSD:  minUSD,
		EstimatedMaxUSD:  maxUSD,
		EstimatedDisplay: fmt.Sprintf("$%s - $%s", formatInt(minUSD), formatInt(maxUSD)),
		LengthScore:      lengthScore,
		TLDScore:         tldScore,
		PhoneticScore:    phoneticScore,
		KeywordScore:     keywordScore,
		Highlights:       highlights,
	}
}

func analyzePhonetics(word string) int {
	if len(word) == 0 {
		return 0
	}
	vowels := 0
	consonants := 0
	maxConsecutiveConsonants := 0
	curConsonants := 0

	for _, r := range word {
		if !unicode.IsLetter(r) {
			continue
		}
		if strings.ContainsRune("aeiouy", r) {
			vowels++
			curConsonants = 0
		} else {
			consonants++
			curConsonants++
			if curConsonants > maxConsecutiveConsonants {
				maxConsecutiveConsonants = curConsonants
			}
		}
	}

	if vowels == 0 {
		return 4
	}
	ratio := float64(vowels) / float64(len(word))
	score := 15
	if ratio >= 0.30 && ratio <= 0.55 {
		score = 20
	}
	if maxConsecutiveConsonants >= 4 {
		score -= 8
	} else if maxConsecutiveConsonants == 3 {
		score -= 3
	}
	if score < 4 {
		score = 4
	}
	return score
}

func hasDigits(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func scoreToPriceRange(score int, tld string, length int) (string, int, int) {
	multiplier := 1.0
	if tld == "com" {
		multiplier = 1.6
	} else if tld == "ai" {
		multiplier = 1.4
	} else if tld == "io" {
		multiplier = 1.2
	}
	if length <= 5 {
		multiplier *= 1.5
	}

	switch {
	case score >= 85:
		min := int(float64(2200) * multiplier)
		max := int(float64(6500) * multiplier)
		return "Ultra Premium", min, max
	case score >= 74:
		min := int(float64(650) * multiplier)
		max := int(float64(2100) * multiplier)
		return "High Value", min, max
	case score >= 62:
		min := int(float64(180) * multiplier)
		max := int(float64(600) * multiplier)
		return "Brandable", min, max
	default:
		return "Standard", 15, 120
	}
}

func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	rem := len(s) % 3
	if rem > 0 {
		out = append(out, s[:rem]...)
		if len(s) > rem {
			out = append(out, ',')
		}
	}
	for i := rem; i < len(s); i += 3 {
		out = append(out, s[i:i+3]...)
		if i+3 < len(s) {
			out = append(out, ',')
		}
	}
	return string(out)
}
