package domain

// DictionaryPack represents a curated collection of real English dictionary words & roots for domain discovery.
type DictionaryPack struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Words       []string `json:"words"`
}

var BuiltinDictionaryPacks = []DictionaryPack{
	{
		ID:          "short_english",
		Name:        "Short 4–5 Letter English Words",
		Description: "Crisp, single-syllable & two-syllable English dictionary nouns and verbs",
		Words: []string{
			"loom", "cove", "kiln", "helm", "arch", "dune", "vibe", "echo", "tide", "zeal",
			"folk", "wren", "kith", "veld", "brim", "glint", "spire", "flint", "crest", "verve",
		},
	},
	{
		ID:          "ai_neural",
		Name:        "AI, Agents & Compute Dictionary",
		Description: "High-velocity artificial intelligence, model inference, and autonomous agent roots",
		Words: []string{
			"agent", "neural", "cortex", "tensor", "infer", "prompt", "token", "synth",
			"vector", "latent", "epoch", "cipher", "quorum", "matrix", "axon", "synapse",
		},
	},
	{
		ID:          "devtools_infra",
		Name:        "Systems, DevTools & Cloud Infra",
		Description: "Developer tooling, distributed systems, and edge networking vocabulary",
		Words: []string{
			"deploy", "socket", "kernel", "runtime", "trace", "beacon", "proxy", "vault",
			"cache", "mesh", "anvil", "daemon", "relay", "fabric", "ledger", "shard",
		},
	},
	{
		ID:          "fintech_capital",
		Name:        "Fintech, Treasury & Commerce",
		Description: "Financial infrastructure, settlement, and institutional capital terminology",
		Words: []string{
			"treasury", "yield", "settle", "clearing", "float", "escrow", "capital",
			"payout", "basis", "mint", "quota", "margin", "credit", "liquid", "tender",
		},
	},
	{
		ID:          "design_atelier",
		Name:        "Design Studio, Editorial & Craft",
		Description: "Editorial publishing, industrial design, and creative studio nouns",
		Words: []string{
			"atelier", "foundry", "canvas", "press", "edition", "folio", "forma",
			"motif", "serif", "monolith", "patina", "studio", "draft", "index", "type",
		},
	},
}

// GetDictionaryWords returns words for a given pack ID (or all packs if "all").
func GetDictionaryWords(packID string) []string {
	for _, p := range BuiltinDictionaryPacks {
		if p.ID == packID {
			return p.Words
		}
	}
	return nil
}
