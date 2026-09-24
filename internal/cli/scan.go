package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rny/scanner/internal/domain"
	"github.com/spf13/cobra"
)

var (
	flagTLDs          string
	flagMutations     bool
	flagOnlyAvailable bool
	flagMinScore      int
	flagMaxResults    int
)

var scanCmd = &cobra.Command{
	Use:   "scan [keywords...]",
	Short: "Scan for available (empty) domains and score their market value (Default Mode)",
	Long:  "Generates brandable domain candidates across top TLDs, checks registration availability in parallel, and evaluates aftermarket brand value.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeScan(args)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&flagTLDs, "tlds", "t", "com,ai,io,dev,co,app", "Comma-separated list of TLDs to scan")
	scanCmd.Flags().BoolVarP(&flagMutations, "mutations", "m", true, "Generate brandable prefix/suffix variations (e.g. hq, labs, flow, get)")
	scanCmd.Flags().BoolVarP(&flagOnlyAvailable, "available", "a", false, "Show only available (unregistered) domains")
	scanCmd.Flags().IntVarP(&flagMinScore, "min-score", "s", 0, "Minimum valuation score (0-100)")
	scanCmd.Flags().IntVarP(&flagMaxResults, "limit", "l", 36, "Maximum number of domain candidates to check")
}

func executeScan(args []string) error {
	keywords := args
	if len(keywords) == 0 {
		keywords = []string{"nova", "pulse"}
	}

	var tlds []string
	for _, t := range strings.Split(flagTLDs, ",") {
		if trimmed := strings.TrimSpace(t); trimmed != "" {
			tlds = append(tlds, trimmed)
		}
	}

	fmt.Println(RenderBanner("Mode: Valuable Available Domain Scanner  •  Seeds: " + strings.Join(keywords, ", ")))

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	report := domain.ScanDomains(ctx, domain.ScanOptions{
		Keywords:      keywords,
		TLDs:          tlds,
		Mutations:     flagMutations,
		OnlyAvailable: flagOnlyAvailable,
		MinScore:      flagMinScore,
		Concurrency:   14,
		MaxResults:    flagMaxResults,
	})

	summaryBox := fmt.Sprintf(
		"%s %s    %s %s    %s %s    %s %s    %s %s",
		SubtleStyle.Render("Checked:"),
		ValueStyle.Render(fmt.Sprintf("%d", report.TotalChecked)),
		SubtleStyle.Render("Available (Empty):"),
		lipgloss.NewStyle().Foreground(ColorEmerald).Bold(true).Render(fmt.Sprintf("%d", report.AvailableCount)),
		SubtleStyle.Render("High-Value Empty:"),
		lipgloss.NewStyle().Foreground(ColorCyan).Bold(true).Render(fmt.Sprintf("%d", report.HighValueCount)),
		SubtleStyle.Render("Registered:"),
		lipgloss.NewStyle().Foreground(ColorRose).Bold(true).Render(fmt.Sprintf("%d", report.TakenCount)),
		SubtleStyle.Render("Duration:"),
		ValueStyle.Render(fmt.Sprintf("%dms", report.DurationMs)),
	)
	fmt.Println(CardStyle.Render(summaryBox))

	fmt.Println(SectionHeaderStyle.Render(" DOMAIN CANDIDATES & VALUATION MATRIX "))

	header := fmt.Sprintf("  %-26s %-14s %-8s %-18s %-18s %s",
		"DOMAIN", "STATUS", "SCORE", "TIER", "EST. VALUE", "KEY VALUE DRIVER")
	fmt.Println(SubtleStyle.Render(header))
	fmt.Println(SubtleStyle.Render("  " + strings.Repeat("─", 98)))

	shown := 0
	for _, item := range report.Items {
		if shown >= 28 {
			break
		}
		shown++

		statusBadge := BadgeAvailable.Render(" AVAILABLE ")
		domainRender := lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Width(26).Render(item.Domain)
		if !item.Available {
			statusBadge = BadgeTaken.Render(" REGISTERED")
			domainRender = lipgloss.NewStyle().Foreground(ColorSlate).Width(26).Render(item.Domain)
		}

		scoreColor := ColorSlate
		if item.Valuation.Score >= 85 {
			scoreColor = ColorViolet
		} else if item.Valuation.Score >= 74 {
			scoreColor = ColorCyan
		} else if item.Valuation.Score >= 62 {
			scoreColor = ColorAmber
		}
		scoreStr := lipgloss.NewStyle().Bold(true).Foreground(scoreColor).Width(8).Render(fmt.Sprintf("%d/100", item.Valuation.Score))

		tierStr := fmt.Sprintf("%-18s", item.Valuation.Tier)
		estStr := lipgloss.NewStyle().Foreground(ColorWhite).Width(18).Render(item.Valuation.EstimatedDisplay)

		highlight := ""
		if len(item.Valuation.Highlights) > 0 {
			highlight = SubtleStyle.Render(item.Valuation.Highlights[0])
		}

		fmt.Printf("  %s %-14s %s %s %s %s\n",
			domainRender,
			statusBadge,
			scoreStr,
			tierStr,
			estStr,
			highlight,
		)
	}

	fmt.Println()
	fmt.Println(SubtleStyle.Render("  Tip: Run `scanner inspect <domain>` for registration dates & WHOIS/RDAP history, or `scanner crawl <domain>` for site structure."))
	return nil
}
