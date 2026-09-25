package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Valuation holds realistic registration costs and calibrated institutional/aftermarket valuation metrics for a domain.
type Valuation struct {
	Score            int      `json:"score"`            // 0 - 100 composite quality index
	Tier             string   `json:"tier"`             // e.g. "Global Enterprise Flagship", "Institutional .COM Asset", "Prime Unclaimed Gem"
	RegFeeUSD        int      `json:"regFeeUsd"`        // Standard 1st-year registrar fee in USD
	RegFeeDisplay    string   `json:"regFeeDisplay"`    // e.g. "$10/yr Reg Fee"
	EstimatedMinUSD  int      `json:"estimatedMinUsd"`  // Realistic aftermarket resale low
	EstimatedMaxUSD  int      `json:"estimatedMaxUsd"`  // Realistic aftermarket resale high
	EstimatedDisplay string   `json:"estimatedDisplay"` // e.g. "$5,000,000+ (Institutional Asset)" or "$10/yr · Flip $95–$340"
	LengthScore      int      `json:"lengthScore"`      // 0 - 30
	TLDScore         int      `json:"tldScore"`         // 0 - 25
	PhoneticScore    int      `json:"phoneticScore"`    // 0 - 25
	KeywordScore     int      `json:"keywordScore"`     // 0 - 20
	IsDictionaryWord bool     `json:"isDictionaryWord"` // True if exact root is a known dictionary word or flagship compound
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

// flagshipEnterpriseDomains maps iconic global companies & platforms to their baseline institutional valuation (in USD).
var flagshipEnterpriseDomains = map[string]struct {
	minUSD int
	maxUSD int
	label  string
}{
	"cloudflare.com":   {5000000, 25000000, "$5,000,000+ · Global Flagship"},
	"google.com":       {50000000, 100000000, "$50,000,000+ · Global Flagship"},
	"apple.com":        {50000000, 100000000, "$50,000,000+ · Global Flagship"},
	"microsoft.com":    {50000000, 100000000, "$50,000,000+ · Global Flagship"},
	"amazon.com":       {50000000, 100000000, "$50,000,000+ · Global Flagship"},
	"nvidia.com":       {25000000, 80000000, "$25,000,000+ · Global Flagship"},
	"openai.com":       {15000000, 50000000, "$15,000,000+ · Global Flagship"},
	"anthropic.com":    {5000000, 20000000, "$5,000,000+ · Global Flagship"},
	"stripe.com":       {10000000, 35000000, "$10,000,000+ · Global Flagship"},
	"github.com":       {10000000, 35000000, "$10,000,000+ · Global Flagship"},
	"vercel.com":       {2500000, 10000000, "$2,500,000+ · Global Flagship"},
	"supabase.com":     {1500000, 6000000, "$1,500,000+ · Global Flagship"},
	"linear.app":       {850000, 3200000, "$850,000+ · Category Flagship"},
	"svelte.dev":       {450000, 1800000, "$450,000+ · Ecosystem Flagship"},
	"golang.org":       {650000, 2400000, "$650,000+ · Ecosystem Flagship"},
	"figma.com":        {5000000, 18000000, "$5,000,000+ · Global Flagship"},
	"notion.so":        {1500000, 5000000, "$1,500,000+ · Category Flagship"},
	"notion.com":       {5000000, 15000000, "$5,000,000+ · Global Flagship"},
	"discord.com":      {8000000, 25000000, "$8,000,000+ · Global Flagship"},
	"shopify.com":      {10000000, 30000000, "$10,000,000+ · Global Flagship"},
	"datadog.com":      {4000000, 15000000, "$4,000,000+ · Global Flagship"},
	"snowflake.com":    {4000000, 15000000, "$4,000,000+ · Global Flagship"},
	"crowdstrike.com":  {3500000, 12000000, "$3,500,000+ · Global Flagship"},
	"palantir.com":     {3500000, 12000000, "$3,500,000+ · Global Flagship"},
	"coinbase.com":     {5000000, 20000000, "$5,000,000+ · Global Flagship"},
	"airbnb.com":       {10000000, 30000000, "$10,000,000+ · Global Flagship"},
	"uber.com":         {15000000, 45000000, "$15,000,000+ · Global Flagship"},
	"netflix.com":      {15000000, 45000000, "$15,000,000+ · Global Flagship"},
	"spotify.com":      {10000000, 30000000, "$10,000,000+ · Global Flagship"},
	"reddit.com":       {10000000, 30000000, "$10,000,000+ · Global Flagship"},
	"tailwindcss.com":  {450000, 1500000, "$450,000+ · Ecosystem Flagship"},
	"cursor.com":       {1800000, 6500000, "$1,800,000+ · Category Flagship"},
	"raycast.com":      {650000, 2200000, "$650,000+ · Category Flagship"},
	"posthog.com":      {650000, 2200000, "$650,000+ · Category Flagship"},
	"sentry.io":        {650000, 2200000, "$650,000+ · Category Flagship"},
	"replit.com":       {1200000, 4500000, "$1,200,000+ · Category Flagship"},
	"perplexity.ai":    {1800000, 6500000, "$1,800,000+ · Category Flagship"},
	"huggingface.co":   {1200000, 4500000, "$1,200,000+ · Category Flagship"},
}

// flagshipBrandRoots recognizes globally established enterprise roots across extensions.
var flagshipBrandRoots = map[string]bool{
	"cloudflare": true, "google": true, "apple": true, "microsoft": true, "amazon": true,
	"nvidia": true, "openai": true, "anthropic": true, "stripe": true, "github": true,
	"vercel": true, "supabase": true, "linear": true, "svelte": true, "golang": true,
	"figma": true, "notion": true, "discord": true, "shopify": true, "datadog": true,
	"snowflake": true, "crowdstrike": true, "palantir": true, "coinbase": true, "airbnb": true,
	"uber": true, "netflix": true, "spotify": true, "reddit": true, "cursor": true,
	"raycast": true, "posthog": true, "sentry": true, "replit": true, "perplexity": true,
	"huggingface": true, "atlassian": true, "gitlab": true, "docker": true, "kubernetes": true,
	"salesforce": true, "oracle": true, "adobe": true, "fastly": true, "akamai": true,
	"digitalocean": true, "planetscale": true, "elevenlabs": true, "midjourney": true,
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
	"cloud": true, "flare": true, "data": true, "flow": true, "sync": true, "core": true,
	"stack": true, "voice": true, "brain": true, "code": true, "mail": true, "bank": true,
	"pay": true, "cash": true, "loan": true, "fund": true, "trade": true, "coin": true,
	"gold": true, "ai": true, "bot": true, "chat": true, "auto": true, "base": true,
	"edge": true, "host": true, "net": true, "web": true, "link": true, "auth": true,
	"key": true, "lock": true, "safe": true, "zero": true, "trust": true, "guard": true,
	"shield": true, "radar": true, "scan": true, "scope": true, "lens": true, "view": true,
	"mind": true, "think": true, "deep": true, "open": true, "fast": true, "swift": true,
	"bold": true, "pure": true, "true": true, "prime": true, "omni": true, "meta": true,
	"hyper": true, "ultra": true, "super": true, "studio": true, "craft": true, "build": true,
	"ship": true, "launch": true, "work": true, "team": true, "loop": true, "wave": true,
	"beam": true, "glow": true, "aura": true, "peak": true, "cursor": true, "linear": true,
	"stripe": true, "apple": true, "oracle": true, "atlas": true, "prism": true, "vertex": true,
	"cipher": true, "matrix": true, "logic": true, "pilot": true, "scout": true, "relay": true,
}

var commercialRoots = []string{
	"cloud", "flare", "data", "pay", "flow", "sync", "pulse", "core", "hq",
	"labs", "base", "stack", "grid", "hub", "scale", "nexus", "mint",
	"nova", "apex", "flux", "orbit", "forge", "spark", "vector", "agent",
	"ai", "super", "studio", "craft", "dev", "sec", "net", "web", "auth",
	"lock", "shield", "guard", "trust", "pilot", "strike", "crowd", "snow",
	"flake", "dog", "mail", "send", "fire", "open", "deep", "mind", "voice",
	"token", "chain", "coin", "bank", "cash", "fund", "trade", "wealth",
	"edge", "host", "node", "mesh", "link", "beam", "wave", "cast", "ray",
}

// EvaluateDomain computes a realistic valuation for a domain name.
// If the domain is a known global enterprise flagship, it automatically evaluates it as a Registered asset.
func EvaluateDomain(fullDomain string) Valuation {
	clean := strings.ToLower(strings.TrimSpace(fullDomain))
	parts := strings.Split(clean, ".")
	root := parts[0]
	if _, isFlagship := flagshipEnterpriseDomains[clean]; isFlagship || flagshipBrandRoots[root] {
		return EvaluateDomainWithTenure(clean, false, 16, true)
	}
	return EvaluateDomainWithTenure(clean, true, 0, false)
}

// EvaluateDomainWithStatus computes a realistic valuation calibrated by whether the domain is currently available or registered.
func EvaluateDomainWithStatus(fullDomain string, isAvailable bool) Valuation {
	clean := strings.ToLower(strings.TrimSpace(fullDomain))
	parts := strings.Split(clean, ".")
	root := parts[0]
	if _, isFlagship := flagshipEnterpriseDomains[clean]; isFlagship || flagshipBrandRoots[root] {
		return EvaluateDomainWithTenure(clean, false, 16, true)
	}
	return EvaluateDomainWithTenure(clean, isAvailable, 0, !isAvailable)
}

// EvaluateDomainWithTenure computes a comprehensive domain valuation factoring in registration status,
// domain age (tenure in years), compound dictionary semantics, and enterprise DNS posture.
func EvaluateDomainWithTenure(fullDomain string, isAvailable bool, ageYears int, hasEnterpriseDNS bool) Valuation {
	fullDomain = strings.ToLower(strings.TrimSpace(fullDomain))
	parts := strings.Split(fullDomain, ".")
	name := parts[0]
	tld := "com"
	if len(parts) > 1 {
		tld = parts[len(parts)-1]
	}

	meta, ok := tldCatalog[tld]
	if !ok {
		meta = tldMeta{weight: 11, regFee: 14}
	}

	// Check if this is an iconic Global Enterprise Flagship (e.g. cloudflare.com, google.com, stripe.com, svelte.dev)
	if flagship, ok := flagshipEnterpriseDomains[fullDomain]; ok {
		return Valuation{
			Score:            99,
			Tier:             "Global Enterprise Flagship",
			RegFeeUSD:        meta.regFee,
			RegFeeDisplay:    fmt.Sprintf("$%d/yr Reg", meta.regFee),
			EstimatedMinUSD:  flagship.minUSD,
			EstimatedMaxUSD:  flagship.maxUSD,
			EstimatedDisplay: flagship.label,
			LengthScore:      30,
			TLDScore:         25,
			PhoneticScore:    24,
			KeywordScore:     20,
			IsDictionaryWord: true,
			Highlights: []string{
				"Global Tier-1 Internet & Enterprise Flagship Brand",
				"Multi-decade institutional equity & authoritative Anycast DNS footprint",
				fmt.Sprintf("Category-defining .%s flagship asset", tld),
			},
		}
	}

	var highlights []string
	isDict := coreDictionaryWords[name]
	isCompound, stemA, stemB := detectCompoundBrand(name)

	if flagshipBrandRoots[name] && !isAvailable {
		isCompound = true
	}

	// 1. Length score (0 - 30)
	length := len(name)
	lengthScore := 0
	switch {
	case length <= 3:
		lengthScore = 30
		highlights = append(highlights, fmt.Sprintf("Ultra-scarce %d-character root", length))
	case length == 4:
		lengthScore = 28
		highlights = append(highlights, "Short 4-letter liquid root scarcity")
	case length <= 6:
		lengthScore = 25
		highlights = append(highlights, fmt.Sprintf("Compact %d-letter brandable length", length))
	case length <= 8:
		lengthScore = 21
		highlights = append(highlights, fmt.Sprintf("Memorable %d-character brand length", length))
	case length <= 10:
		if isCompound || isDict {
			lengthScore = 22
			highlights = append(highlights, fmt.Sprintf("Clean two-stem compound brand (%s + %s)", stemA, stemB))
		} else {
			lengthScore = 16
		}
	case length <= 13:
		if isCompound {
			lengthScore = 18
		} else {
			lengthScore = 11
		}
	default:
		lengthScore = 6
	}

	// 2. TLD score (0 - 25)
	tldScore := meta.weight
	if tld == "com" {
		highlights = append(highlights, fmt.Sprintf("Global .com commercial standard ($%d/yr renewal)", meta.regFee))
	} else if tld == "ai" || tld == "io" || tld == "dev" || tld == "co" || tld == "app" {
		highlights = append(highlights, fmt.Sprintf("High-demand .%s tech extension ($%d/yr renewal)", tld, meta.regFee))
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

	// 4. Dictionary, Compound Brand & Commercial Keyword Intent (0 - 20)
	keywordScore := 7
	if isDict {
		keywordScore = 20
		highlights = append(highlights, "Exact single-word English dictionary root")
	} else if isCompound {
		keywordScore = 19
		if stemA != "" && stemB != "" {
			highlights = append(highlights, fmt.Sprintf("Exact category compound (%s + %s)", stemA, stemB))
		} else {
			highlights = append(highlights, "Recognized global enterprise brand root")
		}
	} else {
		matchedCount := 0
		for _, kw := range commercialRoots {
			if strings.Contains(name, kw) {
				matchedCount++
			}
		}
		if matchedCount >= 2 {
			keywordScore = 17
			highlights = append(highlights, "Multi-stem commercial keyword pairing")
		} else if matchedCount == 1 {
			keywordScore = 14
			highlights = append(highlights, "Contains high-demand commercial stem")
		}
	}

	// 5. Tenure & Active Enterprise Posture Bonus (for Registered Domains)
	tenureBonus := 0
	if !isAvailable {
		if ageYears >= 15 {
			tenureBonus += 10
			highlights = append(highlights, fmt.Sprintf("%d-year legacy registration tenure", ageYears))
		} else if ageYears >= 8 {
			tenureBonus += 6
			highlights = append(highlights, fmt.Sprintf("%d-year established registration history", ageYears))
		} else if ageYears >= 3 {
			tenureBonus += 3
		}
		if hasEnterpriseDNS {
			tenureBonus += 3
		}
	}

	totalScore := lengthScore + tldScore + phoneticScore + keywordScore + tenureBonus
	if !isAvailable && isDict && (tld == "com" || tld == "ai" || tld == "co" || tld == "io") && totalScore < 94 {
		totalScore = 95
	}
	if !isAvailable && (isCompound || flagshipBrandRoots[name]) && tld == "com" && totalScore < 92 {
		totalScore = 94
	}
	if totalScore > 99 {
		totalScore = 99
	}

	tier, minUSD, maxUSD, estDisplay := calibrateRealisticPricing(
		totalScore,
		tld,
		length,
		isDict,
		isCompound,
		isAvailable,
		ageYears,
		meta.regFee,
	)

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
		IsDictionaryWord: isDict || isCompound,
		Highlights:       highlights,
	}
}

func calibrateRealisticPricing(
	score int,
	tld string,
	length int,
	isDict bool,
	isCompound bool,
	isAvailable bool,
	ageYears int,
	regFee int,
) (tier string, minUSD, maxUSD int, display string) {
	// Case A: Domain is AVAILABLE (unregistered right now)
	// Immediate cost is standard registrar fee ($2 - $68/yr), with realistic aftermarket flip range.
	if isAvailable {
		switch {
		case score >= 84 || (isDict && (tld == "com" || tld == "ai" || tld == "io" || tld == "co")):
			minUSD = 450
			maxUSD = 1850
			tier = "Prime Unclaimed Gem"
			display = fmt.Sprintf("$%d/yr · Flip $%s–$%s", regFee, formatInt(minUSD), formatInt(maxUSD))
		case score >= 73:
			minUSD = 120
			maxUSD = 480
			tier = "High-Potential Brandable"
			display = fmt.Sprintf("$%d/yr · Flip $%d–$%d", regFee, minUSD, maxUSD)
		case score >= 62:
			minUSD = 40
			maxUSD = 150
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

	// Case B: Domain is REGISTERED (Taken / Active Brand or Aftermarket Asset)
	// 1. Ultra-short 2-3 letter .com / .ai / .io / .co
	if length <= 3 && tld == "com" {
		return "Ultra-Prime 3L .COM Asset", 250000, 1500000, "$250,000 – $1,500,000+"
	}
	if length <= 3 && (tld == "ai" || tld == "io" || tld == "co") {
		return "Ultra-Prime Short Tech Asset", 45000, 180000, "$45,000 – $180,000+"
	}

	// 2. Exact Single-Word Dictionary .COM (e.g. agent.com, pulse.com, cloud.com)
	if isDict && tld == "com" {
		if length <= 6 || ageYears >= 12 {
			return "Category-Killer .COM Asset", 180000, 850000, "$180,000 – $850,000+"
		}
		return "Institutional Dictionary .COM", 65000, 240000, "$65,000 – $240,000"
	}

	// 3. Exact Single-Word Dictionary on .AI / .CO / .IO / .DEV (e.g. agent.co, voice.ai, forge.io)
	if isDict && (tld == "ai" || tld == "co" || tld == "io" || tld == "dev") {
		if length <= 6 || ageYears >= 10 {
			return fmt.Sprintf("Category-Killer .%s Asset", strings.ToUpper(tld)), 35000, 125000, "$35,000 – $125,000"
		}
		return fmt.Sprintf("Prime Dictionary .%s Asset", strings.ToUpper(tld)), 14000, 48000, "$14,000 – $48,000"
	}

	// 4. Two-Word Exact Compound Brand on .COM (e.g. cloudflare.com, sendgrid.com, databrick.com)
	if isCompound && tld == "com" {
		if ageYears >= 12 {
			return "Institutional Enterprise .COM", 125000, 650000, "$125,000 – $650,000+"
		}
		return "Tier-1 Compound Brand .COM", 24000, 95000, "$24,000 – $95,000"
	}

	// 5. General Registered Domain Calibration with TLD & Tenure Multipliers
	tldMult := 1.0
	switch tld {
	case "com":
		tldMult = 3.2
	case "ai":
		tldMult = 2.4
	case "io", "co", "dev", "app":
		tldMult = 1.8
	default:
		tldMult = 1.0
	}

	tenureMult := 1.0
	if ageYears >= 15 {
		tenureMult = 4.5
	} else if ageYears >= 10 {
		tenureMult = 2.8
	} else if ageYears >= 5 {
		tenureMult = 1.6
	}

	combinedMult := tldMult * tenureMult

	switch {
	case score >= 88:
		minUSD = int(6500 * combinedMult)
		maxUSD = int(24000 * combinedMult)
		tier = "Institutional Brand Asset"
	case score >= 78:
		minUSD = int(2200 * combinedMult)
		maxUSD = int(7800 * combinedMult)
		tier = "Established Brand Domain"
	case score >= 66:
		minUSD = int(650 * combinedMult)
		maxUSD = int(2400 * combinedMult)
		tier = "Registered Brandable"
	default:
		minUSD = int(150 * combinedMult)
		maxUSD = int(550 * combinedMult)
		tier = "Registered Domain"
	}

	return tier, minUSD, maxUSD, fmt.Sprintf("$%s – $%s", formatInt(minUSD), formatInt(maxUSD))
}

// detectCompoundBrand checks if a name is a clean composition of two known dictionary/commercial stems (e.g. "cloud"+"flare").
func detectCompoundBrand(name string) (bool, string, string) {
	if len(name) < 6 || len(name) > 14 {
		return false, "", ""
	}
	for i := 3; i <= len(name)-3; i++ {
		left := name[:i]
		right := name[i:]
		leftMatch := coreDictionaryWords[left] || containsExactStem(left)
		rightMatch := coreDictionaryWords[right] || containsExactStem(right)
		if leftMatch && rightMatch {
			return true, left, right
		}
	}
	return false, "", ""
}

func containsExactStem(part string) bool {
	for _, s := range commercialRoots {
		if s == part {
			return true
		}
	}
	return false
}

// ComputeRegistrationAgeYears extracts the integer age in years from an ISO/YYYY-MM-DD registration date string.
func ComputeRegistrationAgeYears(registeredAt string) int {
	registeredAt = strings.TrimSpace(registeredAt)
	if registeredAt == "" {
		return 0
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, registeredAt)
		if err != nil && len(registeredAt) >= 10 && layout == "2006-01-02" {
			t, err = time.Parse(layout, registeredAt[:10])
		}
		if err == nil {
			years := int(time.Since(t).Hours() / (24 * 365.25))
			if years < 0 {
				return 0
			}
			return years
		}
	}
	return 0
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
