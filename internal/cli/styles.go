package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rny/scanner/internal/crawler"
)

var (
	ColorCyan    = lipgloss.Color("#22D3EE")
	ColorEmerald = lipgloss.Color("#10B981")
	ColorAmber   = lipgloss.Color("#F59E0B")
	ColorRose    = lipgloss.Color("#F43F5E")
	ColorViolet  = lipgloss.Color("#A78BFA")
	ColorSlate   = lipgloss.Color("#94A3B8")
	ColorMuted   = lipgloss.Color("#64748B")
	ColorWhite   = lipgloss.Color("#F8FAFC")

	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0E7490")).
			Padding(0, 2).
			MarginBottom(1)

	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(lipgloss.Color("#1E293B")).
				Padding(0, 1).
				MarginTop(1).
				MarginBottom(1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(1, 2).
			MarginBottom(1)

	BadgeAvailable = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#022C22")).
			Background(ColorEmerald).
			Padding(0, 1)

	BadgeTaken = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4C0519")).
			Background(ColorRose).
			Padding(0, 1)

	BadgeUltra = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1E1B4B")).
			Background(ColorViolet).
			Padding(0, 1)

	BadgeHigh = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#083344")).
			Background(ColorCyan).
			Padding(0, 1)

	BadgeBrandable = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#451A03")).
			Background(ColorAmber).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSlate).
			Width(18)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// RenderBanner prints the Scanner CLI header.
func RenderBanner(subtitle string) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("◈ SCANNER // DOMAIN INTELLIGENCE & SITE STRUCTURE SUITE")
	sub := SubtleStyle.Render(subtitle)
	return BannerStyle.Render(fmt.Sprintf("%s\n%s", title, sub))
}

// RenderTierBadge returns a styled Lipgloss pill for the domain valuation tier.
func RenderTierBadge(tier string) string {
	switch tier {
	case "Ultra Premium":
		return BadgeUltra.Render("★ ULTRA PREMIUM")
	case "High Value":
		return BadgeHigh.Render("◆ HIGH VALUE")
	case "Brandable":
		return BadgeBrandable.Render("● BRANDABLE")
	default:
		return SubtleStyle.Render("○ STANDARD")
	}
}

// RenderSiteTree renders a hierarchical crawler.SiteNode tree with box-drawing connectors.
func RenderSiteTree(node *crawler.SiteNode) string {
	if node == nil {
		return SubtleStyle.Render("  (no nodes discovered)")
	}
	var sb strings.Builder
	rootBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("🌐 " + node.Segment)
	statusPill := renderStatusCode(node.StatusCode)
	titleStr := ""
	if node.Title != "" {
		titleStr = SubtleStyle.Render(" — " + node.Title)
	}
	sb.WriteString(fmt.Sprintf("  %s %s%s\n", rootBadge, statusPill, titleStr))

	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		renderTreeNode(&sb, child, "  ", isLast)
	}
	return sb.String()
}

func renderTreeNode(sb *strings.Builder, node *crawler.SiteNode, prefix string, isLast bool) {
	connector := "├── "
	nextPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		nextPrefix = prefix + "    "
	}

	segStyle := lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	if len(node.Children) > 0 {
		segStyle = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
	}

	status := renderStatusCode(node.StatusCode)
	meta := ""
	if node.Title != "" {
		meta = SubtleStyle.Render(fmt.Sprintf(" (%s)", node.Title))
	}
	latency := ""
	if node.LatencyMs > 0 {
		latency = SubtleStyle.Render(fmt.Sprintf(" [%dms]", node.LatencyMs))
	}

	sb.WriteString(fmt.Sprintf("%s%s%s %s%s%s\n",
		SubtleStyle.Render(prefix+connector),
		segStyle.Render("/"+node.Segment),
		status,
		latency,
		meta,
		"",
	))

	for i, child := range node.Children {
		childLast := i == len(node.Children)-1
		renderTreeNode(sb, child, nextPrefix, childLast)
	}
}

func renderStatusCode(code int) string {
	if code == 0 {
		return SubtleStyle.Render("[discovered]")
	}
	switch {
	case code >= 200 && code < 300:
		return lipgloss.NewStyle().Foreground(ColorEmerald).Render(fmt.Sprintf("[%d]", code))
	case code >= 300 && code < 400:
		return lipgloss.NewStyle().Foreground(ColorAmber).Render(fmt.Sprintf("[%d]", code))
	default:
		return lipgloss.NewStyle().Foreground(ColorRose).Render(fmt.Sprintf("[%d]", code))
	}
}
