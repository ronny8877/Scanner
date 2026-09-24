package domain

import (
	"fmt"
	"strings"
	"unicode"
)

// Valuation holds realistic registration costs and aftermarket valuation metrics for a domain.
type Valuation struct {
	Score            int      `json:"score"`            // 0 - 100 composite quality index
	Tier             string   `json:"tier"`             // "Category Killer", "Prime Brandable", "Solid Brand", "Standard Reg"
	RegFeeUSD        int      `json:"regFeeUsd"`        // Standard 1st-year registrar fee in USD
	RegFeeDisplay    string   `json:"regFeeDisplay"`    // e.g. "$10/yr Reg Fee"
	EstimatedMinUSD  int      `json:"estimatedMinUsd"`  // Realistic aftermarket resale low
	EstimatedMaxUSD  int      `json:"estimatedMaxUsd"`  // Realistic aftermarket resale high
	EstimatedDisplay string   `json:"estimatedDisplay"` // e.g. "$180 - $550" or "$12/yr (Reg)"
	LengthScore      int      `json:"lengthScore"`      // 0 - 30
	TLDScore         int      `json:"tldScore"`         // 0 - 25
	PhoneticScore    int      `json:"phoneticScore"`    // 0 - 25
	KeywordScore     int      `json:"keywordScore"`     // 0 - 20
	IsDictionaryWord bool     `json:"isDictionaryWord"` // True if exact root is a known dictionary word
	Highlights       []string `json:"highlights"`       // Reasons contributing to value
}

type tldMeta struct {
	weight int
	regFee int // Standard annual registration cost in USD
}

var tldCatalog = map[string]tldMeta{
	"com":     {25, 10},
	"ai":      {24, 68},
	"io":      {21, 38},
	"dev":     {19, 12},
	"co":      {19, 24},
	"app":     {18, 14},
	"net":     {16, 12},
	"org":     {16, 11},
	"sh":      {16, 32},
	"gg":      {16, 42},
	"so":      {15, 48},
	"cloud":   {15, 16},
	"tech":    {14, 14},
	"studio":  {15, 22},
	"design":  {14, 28},
	"tools":   {14, 24},
	"codes":   {14, 28},
	"xyz":     {13, 2},
	"me":      {15, 16},
	"vc":      {18, 55},
	"finance": {15, 35},
	"store":   {13, 8},
	"shop":    {13, 8},
	"build":   {14, 22},
}

var coreDictionaryWords = map[string]bool{
	"nova": true, "pulse": true, "apex": true, "flux": true, "mint": true, "loom": true,
	"cove": true, "kiln": true, "helm": true, "arch": true, "dune": true, "vibe": true,
	"echo": true, "tide": true, "zeal": true, "grid": true, "node": true, "folk": true,
	"wren": true, "spark": true, "forge": true, "orbit": true, "nexus": true, "scale": true,
	"vector": true, "agent": true, "neural": true, "cortex": true, "tensor": true, "prompt": true,
	"token": true, "synth": true, "model": true, "deploy": true, "socket": true, "kernel": true,
	"runtime": true, "trace": true, "beacon": true, "proxy": true, "vault": true, "cache": true,
	"mesh": true, "ledger": true, "yield": true, "settle": true, "float": true, "escrow": true,
	"capital": true, "basis": true, "atelier": true, "foundry": true, "canvas": true, "press": true,
	"folio": true, "forma": true, "motif": true, "serif": true, "svelte": true, "golang": true,
	"cloud": true, "data": true, "flow": true, "sync": true, "core": true, "stack": true,
}

var commercialRoots = []string{
	"cloud", "data", "pay", "flow", "sync", "pulse", "core", "hq",
	"labs", "base", "stack", "grid", "hub", "scale", "nexus", "mint",
	"nova", "apex", "flux", "orbit", "forge", "spark", "vector", "agent",
	"ai", "super", "color", "coloring", "studio", "craft", "dev", "sec",
}

// EvaluateDomain computes a realistic valuation for a domain name.
func EvaluateDomain(fullDomain string) Valuation {
	return EvaluateDomainWithStatus(fullDomain, true)
}

// EvaluateDomainWithStatus computes a realistic valuation calibrated by whether the domain is currently available or registered.
func EvaluateDomainWithStatus(fullDomain string, isAvailable bool) Valuation {
	fullDomain = strings.ToLower(strings.TrimSpace(fullDomain))
	parts := strings.Split(fullDomain, ".")
	name := parts[0]
	tld := "com"
	if len(parts) > 1 {
		tld = parts[len(parts)-1]
	}

	var highlights []string

	meta, ok := tldCatalog[tld]
	if !ok {
		meta = tldMeta{weight: 11, regFee: 14}
	}

	isDict := coreDictionaryWords[name]

	// 1. Length score (0 - 30)
	length := len(name)
	lengthScore := 0
	switch {
	case length <= 3:
		lengthScore = 30
		highlights = append(highlights, fmt.Sprintf("Rare %d-character ultra-short root", length))
	case length == 4:
		lengthScore = 27
		highlights = append(highlights, "Short 4-letter root scarcity")
	case length <= 6:
		lengthScore = 23
		highlights = append(highlights, fmt.Sprintf("Compact %d-letter brandable length", length))
	case length <= 8:
		lengthScore = 18
		highlights = append(highlights, fmt.Sprintf("Memorable %d-character length", length))
	case length <= 11:
		lengthScore = 13
	case length <= 14:
		lengthScore = 9
	default:
		lengthScore = 5
	}

	// 2. TLD score (0 - 25)
	tldScore := meta.weight
	if tld == "com" {
		highlights = append(highlights, fmt.Sprintf("Global .com standard ($%d/yr reg)", meta.regFee))
	} else if tld == "ai" || tld == "io" || tld == "dev" || tld == "app" {
		highlights = append(highlights, fmt.Sprintf("High-adoption .%s tech extension ($%d/yr reg)", tld, meta.regFee))
	} else {
		highlights = append(highlights, fmt.Sprintf(".%s extension ($%d/yr standard renewal)", tld, meta.regFee))
	}

	// 3. Phonetic / Cleanliness score (0 - 25)
	phoneticScore := analyzePhonetics(name)
	hasHyphen := strings.Contains(name, "-")
	hasNum := hasDigits(name)
	if !hasHyphen && !hasNum {
		phoneticScore += 4
		if phoneticScore > 25 {
			phoneticScore = 25
		}
		if phoneticScore >= 19 {
			highlights = append(highlights, "Natural vowel-consonant phonetic cadence")
		}
	} else {
		phoneticScore -= 6
		if phoneticScore < 2 {
			phoneticScore = 2
		}
		highlights = append(highlights, "Contains hyphens or digits (discounts resale liquidity)")
	}

	// 4. Dictionary & Commercial Keyword Intent (0 - 20)
	keywordScore := 7
	if isDict {
		keywordScore = 20
		highlights = append(highlights, "Exact single-word English dictionary root")
	} else {
		matchedCount := 0
		for _, kw := range commercialRoots {
			if strings.Contains(name, kw) {
				matchedCount++
			}
		}
		if matchedCount >= 2 {
			keywordScore = 16
			highlights = append(highlights, "Compound commercial keyword pairing")
		} else if matchedCount == 1 {
			keywordScore = 13
			highlights = append(highlights, "Contains high-demand commercial stem")
		}
	}

	totalScore := lengthScore + tldScore + phoneticScore + keywordScore
	if !isAvailable && isDict && tld == "com" && totalScore < 92 {
		totalScore = 94
	}
	if totalScore > 99 {
		totalScore = 99
	}

	tier, minUSD, maxUSD, estDisplay := calibrateRealisticPricing(totalScore, tld, length, isDict, isAvailable, meta.regFee)

	return Valuation{
		Score:            totalScore,
		Tier:             tier,
		RegFeeUSD:        meta.regFee,
		RegFeeDisplay:    fmt.Sprintf("$%d/yr Reg", meta.regFee),
		EstimatedMinUSD:  minUSD,
		EstimatedMaxUSD:  maxUSD,
		EstimatedDisplay: estDisplay,
		LengthScore:      lengthScore,
		TLDScore:         tldScore,
		PhoneticScore:    phoneticScore,
		KeywordScore:     keywordScore,
		IsDictionaryWord: isDict,
		Highlights:       highlights,
	}
}

func calibrateRealisticPricing(
	score int,
	tld string,
	length int,
	isDict bool,
	isAvailable bool,
	regFee int,
) (tier string, minUSD, maxUSD int, display string) {
	// Case A: Domain is AVAILABLE (unregistered right now)
	// Its immediate cost is the standard registrar fee ($2 - $68/yr), and its aftermarket flip potential is realistic.
	if isAvailable {
		switch {
		case score >= 84 || (isDict && (tld == "com" || tld == "ai" || tld == "io")):
			minUSD = 350
			maxUSD = 1250
			tier = "Prime Unclaimed Gem"
			display = fmt.Sprintf("$%d/yr · Flip $%s–$%s", regFee, formatInt(minUSD), formatInt(maxUSD))
		case score >= 73:
			minUSD = 95
			maxUSD = 340
			tier = "High-Potential Brandable"
			display = fmt.Sprintf("$%d/yr · Flip $%d–$%d", regFee, minUSD, maxUSD)
		case score >= 62:
			minUSD = 30
			maxUSD = 120
			tier = "Clean Available Name"
			display = fmt.Sprintf("$%d/yr · Flip $%d–$%d", regFee, minUSD, maxUSD)
		default:
			minUSD = regFee
			maxUSD = regFee * 2
			tier = "Standard Registration"
			display = fmt.Sprintf("$%d/yr Standard Reg", regFee)
		}
		return tier, minUSD, maxUSD, display
	}

	// Case B: Domain is REGISTERED (Active / Taken in the aftermarket)
	if isDict && length <= 5 && tld == "com" {
		return "Institutional .COM Asset", 35000, 140000, "$35,000 - $140,000+"
	}
	if isDict && length <= 6 && (tld == "ai" || tld == "io" || tld == "dev") {
		return "Category-Defining Tech Asset", 4500, 18000, "$4,500 - $18,000"
	}

	tldMult := 1.0
	switch tld {
	case "com":
		tldMult = 2.2
	case "ai":
		tldMult = 1.7
	case "io", "co", "dev":
		tldMult = 1.3
	default:
		tldMult = 0.85
	}

	switch {
	case score >= 85:
		minUSD = int(1800 * tldMult)
		maxUSD = int(5800 * tldMult)
		tier = "Premium Registered Asset"
	case score >= 74:
		minUSD = int(550 * tldMult)
		maxUSD = int(1950 * tldMult)
		tier = "Established Brand Domain"
	case score >= 62:
		minUSD = int(140 * tldMult)
		maxUSD = int(480 * tldMult)
		tier = "Registered Brandable"
	default:
		minUSD = 25
		maxUSD = 110
		tier = "Standard Registered"
	}

	return tier, minUSD, maxUSD, fmt.Sprintf("$%s - $%s", formatInt(minUSD), formatInt(maxUSD))
}

func analyzePhonetics(word string) int {
	if len(word) == 0 {
		return 0
	}
	vowels := 0
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
		score = 21
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
