package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rny/scanner/internal/crawler"
	"github.com/spf13/cobra"
)

var (
	flagMaxPages int
	flagMaxDepth int
)

var crawlCmd = &cobra.Command{
	Use:     "crawl <domain-or-url>",
	Aliases: []string{"tree", "structure"},
	Short:   "Crawl a website to map and visualize its full URL hierarchy and page structure",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		fmt.Println(RenderBanner("Mode: Site Structure Crawler & Hierarchy Mapper  •  Target: " + target))

		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()

		report := crawler.CrawlSite(ctx, crawler.CrawlOptions{
			TargetURL: target,
			MaxPages:  flagMaxPages,
			MaxDepth:  flagMaxDepth,
		})

		techStr := "Standard HTTP"
		if len(report.TechHeaders) > 0 {
			techStr = strings.Join(report.TechHeaders, "  •  ")
		}

		summary := fmt.Sprintf(
			"%s %s    %s %s    %s %s    %s %s\n%s %s",
			SubtleStyle.Render("Host:"),
			ValueStyle.Render(report.Host),
			SubtleStyle.Render("Pages Crawled:"),
			lipgloss.NewStyle().Foreground(ColorEmerald).Bold(true).Render(fmt.Sprintf("%d", report.PagesCrawled)),
			SubtleStyle.Render("Internal Routes:"),
			lipgloss.NewStyle().Foreground(ColorCyan).Bold(true).Render(fmt.Sprintf("%d", report.TotalLinks)),
			SubtleStyle.Render("Elapsed:"),
			ValueStyle.Render(fmt.Sprintf("%dms", report.DurationMs)),
			SubtleStyle.Render("Detected Stack:"),
			ValueStyle.Render(techStr),
		)
		fmt.Println(CardStyle.Render(summary))

		fmt.Println(SectionHeaderStyle.Render(" HIERARCHICAL SITE STRUCTURE TREE "))
		fmt.Println(RenderSiteTree(report.Tree))

		fmt.Println(SectionHeaderStyle.Render(" CRAWLED PAGE INVENTORY "))
		fmt.Println(SubtleStyle.Render(fmt.Sprintf("  %-28s %-8s %-10s %-10s %s", "PATH", "STATUS", "LATENCY", "LINKS", "PAGE TITLE")))
		fmt.Println(SubtleStyle.Render("  " + strings.Repeat("─", 92)))

		for _, p := range report.Pages {
			pathCol := lipgloss.NewStyle().Foreground(ColorCyan).Width(28).Render(truncateStr(p.Path, 26))
			statusCol := fmt.Sprintf("%-8d", p.StatusCode)
			latCol := SubtleStyle.Render(fmt.Sprintf("%-10s", fmt.Sprintf("%dms", p.LatencyMs)))
			linksCol := ValueStyle.Width(10).Render(fmt.Sprintf("%d int", p.InternalLinks))
			titleCol := p.Title
			if titleCol == "" {
				titleCol = SubtleStyle.Render("(untitled)")
			}
			fmt.Printf("  %s %s %s %s %s\n", pathCol, statusCol, latCol, linksCol, titleCol)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	crawlCmd.Flags().IntVarP(&flagMaxPages, "pages", "p", 20, "Maximum number of pages to crawl")
	crawlCmd.Flags().IntVarP(&flagMaxDepth, "depth", "d", 2, "Maximum link depth to follow")
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
