package webintel

import (
	"fmt"
	"net/http"
	"strings"
)

// DetectedTech represents a single fingerprinted frontend framework, UI component kit, CSS system, or platform SDK.
type DetectedTech struct {
	Name        string `json:"name"`
	Category    string `json:"category"`   // "FRAMEWORK", "UI_DESIGN", "BUILD_MOTION", "PLATFORM_AUTH", "EDGE_HOSTING"
	Confidence  string `json:"confidence"` // "CERTAIN", "HIGH", "DETECTED"
	MatchedBy   string `json:"matchedBy"`
	Description string `json:"description"`
}

// TechStackTelemetry aggregates all detected technologies and UI design systems for a domain.
type TechStackTelemetry struct {
	PrimaryFramework string         `json:"primaryFramework"`
	UISystemSummary  string         `json:"uiSystemSummary"`
	EdgePlatform     string         `json:"edgePlatform"`
	Summary          string         `json:"summary"`
	TotalDetected    int            `json:"totalDetected"`
	FrameworksCount  int            `json:"frameworksCount"`
	UICount          int            `json:"uiCount"`
	PlatformCount    int            `json:"platformCount"`
	InfraCount       int            `json:"infraCount"`
	Technologies     []DetectedTech `json:"technologies"`
}

type techSignature struct {
	name        string
	category    string
	confidence  string
	description string
	patterns    []string
	headerKeys  []string
	headerVals  []string
}

var knownTechSignatures = []techSignature{
	// 1. FRONTEND & FULL-STACK FRAMEWORKS
	{
		name:        "Next.js",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "React full-stack framework with App/Pages Router & SSR/ISR",
		patterns:    []string{"/_next/static/", "__next_data__", "self.__next_f.push", "next-route-announcer", "data-nextjs-scroll-focus-boundary"},
		headerKeys:  []string{"x-nextjs-cache", "x-nextjs-matched-path"},
		headerVals:  []string{"next.js"},
	},
	{
		name:        "React",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Component-driven JavaScript UI library by Meta",
		patterns:    []string{"/_next/static/", "__next_data__", "data-reactroot", "data-reactid", "react-dom", "react.production.min.js", "__react_devtools_global_hook__", "self.__next_f.push", "data-remix"},
	},
	{
		name:        "Svelte / SvelteKit",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Compiler-first reactive UI framework & SvelteKit app router",
		patterns:    []string{"__sveltekit", "/_app/immutable/", "data-sveltekit-preload-data", "svelte-", "__svelte"},
	},
	{
		name:        "Vue.js",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Progressive reactive JavaScript frontend framework",
		patterns:    []string{"data-v-", "vue.runtime", "vue.global", "__vue__", "/_nuxt/", "id=\"__nuxt\""},
	},
	{
		name:        "Nuxt",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Full-stack Vue.js meta-framework with nitro server engine",
		patterns:    []string{"/__nuxt", "/_nuxt/", "window.__nuxt__", "id=\"__nuxt\""},
	},
	{
		name:        "Astro",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Islands-architecture web framework for content & interactive apps",
		patterns:    []string{"astro-island", "astro-slot", "data-astro-cid-", "/_astro/"},
	},
	{
		name:        "Remix / React Router",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Web-standards full-stack React router and data loader framework",
		patterns:    []string{"window.__remixcontext", "__remixmanifest", "data-remix"},
	},
	{
		name:        "Angular",
		category:    "FRAMEWORK",
		confidence:  "CERTAIN",
		description: "Enterprise TypeScript web application platform",
		patterns:    []string{"ng-version=", "_ngcontent-", "_nghost-"},
	},
	{
		name:        "Alpine.js",
		category:    "FRAMEWORK",
		confidence:  "HIGH",
		description: "Lightweight declarative reactive DOM directives (x-data, x-bind)",
		patterns:    []string{"x-data=", "alpinejs", "x-init=", "x-show="},
	},
	{
		name:        "HTMX",
		category:    "FRAMEWORK",
		confidence:  "HIGH",
		description: "Hypermedia-driven HTML extensions for AJAX & server fragments",
		patterns:    []string{"hx-get=", "hx-post=", "hx-target=", "htmx.org", "htmx.min.js"},
	},

	// 2. UI COMPONENT KITS & CSS DESIGN SYSTEMS (Tailwind, shadcn/ui, DaisyUI, Radix, MUI, Chakra, etc.)
	{
		name:        "Tailwind CSS",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Utility-first CSS framework with responsive & design-token utilities",
		patterns:    []string{"--tw-", "--tw-ring-offset-shadow", "--tw-backdrop-blur", "backdrop-blur-", "tracking-tight", "rounded-2xl", "grid-cols-", "sm:flex", "lg:grid-cols-", "focus-visible:outline", "bg-background", "text-foreground"},
	},
	{
		name:        "shadcn/ui",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Radix Primitives + Tailwind CSS design-token component architecture",
		patterns:    []string{"--muted-foreground", "--accent-foreground", "--popover-foreground", "--card-foreground", "--destructive-foreground", "bg-background text-foreground", "data-slot=\"", "focus-visible:ring-ring", "border-input bg-background"},
	},
	{
		name:        "DaisyUI",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Semantic Tailwind CSS component library (btn, card, badge, data-theme)",
		patterns:    []string{"daisyui", "--b1:", "--b2:", "--bc:", "btn-primary", "btn-ghost", "card-body", "card-actions", "modal-box", "drawer-toggle", "navbar-start"},
	},
	{
		name:        "Radix UI Primitives",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Unstyled accessible UI primitives powering modern React/shadcn design systems",
		patterns:    []string{"data-radix-", "--radix-", "data-radix-popper-content-wrapper", "data-radix-collection-item", "@radix-ui"},
	},
	{
		name:        "Headless UI",
		category:    "UI_DESIGN",
		confidence:  "HIGH",
		description: "Unstyled accessible UI components by Tailwind Labs",
		patterns:    []string{"data-headlessui-state", "headlessui-"},
	},
	{
		name:        "Material UI (MUI)",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "React Material Design component library with Emotion CSS-in-JS",
		patterns:    []string{"muibutton-root", "muitypography-", "muibox-root", "muisvgicon-root", "data-emotion=\"css"},
	},
	{
		name:        "Chakra UI",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Accessible modular React component library with CSS custom variables",
		patterns:    []string{"--chakra-colors-", "chakra-button", "chakra-ui", "chakra-stack"},
	},
	{
		name:        "Ant Design (antd)",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Enterprise-class React UI design language and component suite",
		patterns:    []string{"ant-btn", "ant-layout", "ant-col", "ant-select", "--ant-"},
	},
	{
		name:        "Bootstrap",
		category:    "UI_DESIGN",
		confidence:  "CERTAIN",
		description: "Classic responsive CSS grid and component toolkit",
		patterns:    []string{"bootstrap.min.css", "bootstrap.bundle", "--bs-primary", "navbar-expand-", "container-fluid"},
	},

	// 3. MOTION, 3D, BUNDLERS & TYPOGRAPHY
	{
		name:        "Framer Motion / Motion",
		category:    "BUILD_MOTION",
		confidence:  "HIGH",
		description: "Production declarative layout & spring animation engine",
		patterns:    []string{"framer-motion", "data-projection-id", "motion-"},
	},
	{
		name:        "GSAP (GreenSock)",
		category:    "BUILD_MOTION",
		confidence:  "HIGH",
		description: "High-performance timeline and scroll-driven web animation toolkit",
		patterns:    []string{"gsap.min.js", "scrolltrigger", "tweenmax", "gsap"},
	},
	{
		name:        "Three.js / WebGL",
		category:    "BUILD_MOTION",
		confidence:  "HIGH",
		description: "3D WebGL shader and scene rendering engine",
		patterns:    []string{"three.min.js", "three.module", "react-three-fiber", "__r3f"},
	},
	{
		name:        "Lucide / Feather Vector Icons",
		category:    "BUILD_MOTION",
		confidence:  "HIGH",
		description: "Clean stroke-based SVG icon library (lucide-*)",
		patterns:    []string{"lucide ", "lucide-", "feather-"},
	},
	{
		name:        "Vite Bundler",
		category:    "BUILD_MOTION",
		confidence:  "HIGH",
		description: "Next-generation ES module frontend build toolchain",
		patterns:    []string{"/@vite/", "/assets/index-", "modulepreload"},
	},

	// 4. CMS, NO-CODE, AUTH & COMMERCE PLATFORMS
	{
		name:        "Framer Site Builder",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Visual interactive web design & publishing platform",
		patterns:    []string{"framerusercontent.com", "data-framer-appear-id", "data-framer-name", "framer-body"},
	},
	{
		name:        "Webflow",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Visual web design & CMS hosting platform",
		patterns:    []string{"data-wf-site", "data-wf-page", "webflow.js", "w-nav", "w-inline-block"},
	},
	{
		name:        "Shopify Commerce",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Hosted e-commerce storefront and checkout platform",
		patterns:    []string{"cdn.shopify.com", "shopify.theme", "shopify-section"},
	},
	{
		name:        "WordPress",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Open-source content management system",
		patterns:    []string{"/wp-content/", "/wp-includes/", "wp-json"},
	},
	{
		name:        "Supabase",
		category:    "PLATFORM_AUTH",
		confidence:  "HIGH",
		description: "Open-source Postgres backend-as-a-service, Auth & Realtime",
		patterns:    []string{"supabase.co", "supabase-js"},
	},
	{
		name:        "Clerk Auth",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Modern user authentication and session management SDK",
		patterns:    []string{"clerk.browser.js", "data-clerk-", "clerk."},
	},
	{
		name:        "Stripe Billing / Checkout",
		category:    "PLATFORM_AUTH",
		confidence:  "CERTAIN",
		description: "Payment processing and checkout infrastructure",
		patterns:    []string{"js.stripe.com", "checkout.stripe.com", "stripe-js"},
	},

	// 5. EDGE INFRASTRUCTURE & CLOUD HOSTING
	{
		name:        "Vercel Edge Network",
		category:    "EDGE_HOSTING",
		confidence:  "CERTAIN",
		description: "Serverless & Edge deployment cloud platform",
		patterns:    []string{"/_vercel/insights", "/_vercel/speed-insights", "x-vercel-challenge"},
		headerKeys:  []string{"x-vercel-id", "x-vercel-cache", "x-vercel-mitigated"},
		headerVals:  []string{"vercel"},
	},
	{
		name:        "Cloudflare Edge & CDN",
		category:    "EDGE_HOSTING",
		confidence:  "CERTAIN",
		description: "Global Anycast CDN, DNS, WAF & Workers edge compute",
		patterns:    []string{"cloudflareinsights.com", "/cdn-cgi/"},
		headerKeys:  []string{"cf-ray", "cf-cache-status"},
		headerVals:  []string{"cloudflare"},
	},
	{
		name:        "Netlify Edge",
		category:    "EDGE_HOSTING",
		confidence:  "CERTAIN",
		description: "Composable web platform and edge CDN",
		patterns:    []string{"netlify"},
		headerKeys:  []string{"x-nf-request-id"},
		headerVals:  []string{"netlify"},
	},
	{
		name:        "AWS CloudFront / S3",
		category:    "EDGE_HOSTING",
		confidence:  "CERTAIN",
		description: "Amazon Web Services global content delivery network",
		headerKeys:  []string{"x-amz-cf-id", "x-amz-cf-pop"},
		headerVals:  []string{"cloudfront", "amazonS3"},
	},
	{
		name:        "Fly.io Edge",
		category:    "EDGE_HOSTING",
		confidence:  "CERTAIN",
		description: "Global application micro-VM edge platform",
		headerKeys:  []string{"fly-request-id"},
		headerVals:  []string{"fly.io"},
	},
}

// DetectTechStackFromHTML inspects HTML documents, linked CSS stylesheets, and HTTP headers
// to fingerprint Frontend Frameworks (React, Next.js, SvelteKit, Vue, Astro), UI Systems (Tailwind CSS, shadcn/ui, DaisyUI, Radix),
// Build/Motion tools, and Cloud Edge platforms.
func DetectTechStackFromHTML(htmlChunks []string, cssChunks []string, headers http.Header) TechStackTelemetry {
	var combinedBuilder strings.Builder
	for _, h := range htmlChunks {
		combinedBuilder.WriteString(strings.ToLower(h))
		combinedBuilder.WriteByte('\n')
	}
	for _, c := range cssChunks {
		combinedBuilder.WriteString(strings.ToLower(c))
		combinedBuilder.WriteByte('\n')
	}
	corpus := combinedBuilder.String()

	headerMap := make(map[string]string)
	var allHeaderValuesBuilder strings.Builder
	for k, vals := range headers {
		lk := strings.ToLower(k)
		joined := strings.ToLower(strings.Join(vals, " "))
		headerMap[lk] = joined
		allHeaderValuesBuilder.WriteString(joined)
		allHeaderValuesBuilder.WriteByte(' ')
	}
	allHeaderVals := allHeaderValuesBuilder.String()

	var detected []DetectedTech
	seen := make(map[string]bool)

	for _, sig := range knownTechSignatures {
		if seen[sig.name] {
			continue
		}
		var matchedRule string

		// Check HTTP header keys first
		for _, hk := range sig.headerKeys {
			if val, ok := headerMap[strings.ToLower(hk)]; ok && val != "" {
				matchedRule = fmt.Sprintf("Header %s", hk)
				break
			}
		}

		// Check HTTP header values (e.g. Server: Vercel, X-Powered-By: Next.js)
		if matchedRule == "" {
			for _, hv := range sig.headerVals {
				if strings.Contains(allHeaderVals, strings.ToLower(hv)) {
					matchedRule = fmt.Sprintf("HTTP Server / Powered-By (%s)", hv)
					break
				}
			}
		}

		// Check HTML + CSS corpus patterns
		if matchedRule == "" && corpus != "" {
			for _, pat := range sig.patterns {
				if strings.Contains(corpus, strings.ToLower(pat)) {
					matchedRule = fmt.Sprintf("Signature `%s`", pat)
					break
				}
			}
		}

		if matchedRule != "" {
			seen[sig.name] = true
			detected = append(detected, DetectedTech{
				Name:        sig.name,
				Category:    sig.category,
				Confidence:  sig.confidence,
				MatchedBy:   matchedRule,
				Description: sig.description,
			})
		}
	}

	// Smart inference: if shadcn/ui or DaisyUI is detected, Tailwind CSS is inherently present
	if (seen["shadcn/ui"] || seen["DaisyUI"] || seen["Headless UI"]) && !seen["Tailwind CSS"] {
		seen["Tailwind CSS"] = true
		detected = append(detected, DetectedTech{
			Name:        "Tailwind CSS",
			Category:    "UI_DESIGN",
			Confidence:  "CERTAIN",
			MatchedBy:   "Inferred from Tailwind component layer (shadcn/ui / DaisyUI)",
			Description: "Utility-first CSS framework powering the active component tokens",
		})
	}

	// Smart inference: if shadcn/ui is detected and React isn't yet marked, infer React
	if seen["shadcn/ui"] && !seen["React"] && !seen["Svelte / SvelteKit"] && !seen["Vue.js"] {
		seen["React"] = true
		detected = append(detected, DetectedTech{
			Name:        "React",
			Category:    "FRAMEWORK",
			Confidence:  "HIGH",
			MatchedBy:   "shadcn/ui + Radix UI React component tree",
			Description: "Component-driven JavaScript UI library",
		})
	}

	var frameworks []string
	var uiSystems []string
	var edgePlatforms []string
	var frameworksCount, uiCount, platformCount, infraCount int

	for _, t := range detected {
		switch t.Category {
		case "FRAMEWORK":
			frameworksCount++
			frameworks = append(frameworks, t.Name)
		case "UI_DESIGN":
			uiCount++
			uiSystems = append(uiSystems, t.Name)
		case "BUILD_MOTION":
			uiCount++
			uiSystems = append(uiSystems, t.Name)
		case "PLATFORM_AUTH":
			platformCount++
		case "EDGE_HOSTING":
			infraCount++
			edgePlatforms = append(edgePlatforms, t.Name)
		}
	}

	primaryFramework := "Custom HTML5 / Server-Rendered"
	if len(frameworks) > 0 {
		primaryFramework = strings.Join(frameworks, " + ")
	}

	uiSystemSummary := "Custom CSS Stylesheet"
	if len(uiSystems) > 0 {
		uiSystemSummary = strings.Join(uiSystems, " · ")
	}

	edgePlatform := "Direct Origin Server"
	if len(edgePlatforms) > 0 {
		edgePlatform = strings.Join(edgePlatforms, " + ")
	}

	summary := fmt.Sprintf("Powered by %s with %s on %s (%d technologies fingerprinted).", primaryFramework, uiSystemSummary, edgePlatform, len(detected))
	if len(detected) == 0 {
		summary = "Standard HTML5/CSS3 web document with zero third-party framework bundles exposed."
	}

	return TechStackTelemetry{
		PrimaryFramework: primaryFramework,
		UISystemSummary:  uiSystemSummary,
		EdgePlatform:     edgePlatform,
		Summary:          summary,
		TotalDetected:    len(detected),
		FrameworksCount:  frameworksCount,
		UICount:          uiCount,
		PlatformCount:    platformCount,
		InfraCount:       infraCount,
		Technologies:     detected,
	}
}
