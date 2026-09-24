package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rny/scanner/internal/domain"
	"github.com/rny/scanner/internal/recon"
	"github.com/spf13/cobra"
)

var reconCmd = &cobra.Command{
	Use:     "recon <domain>",
	Aliases: []string{"ports", "surface"},
	Short:   "Run parallel TCP port scanning, TLS handshake inspection, subdomain discovery & security audit",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		fmt.Println(RenderBanner("Mode: Parallel Port, TLS & Security Recon  •  Target: " + target))

		ctx, cancel := context.WithTimeout(context.Background(), 16*time.Second)
		defer cancel()

		rep := recon.RunRecon(ctx, target)

		summary := fmt.Sprintf(
			"%s %s    %s %s    %s %s    %s %s    %s %s",
			SubtleStyle.Render("Domain:"),
			ValueStyle.Render(rep.Domain),
			SubtleStyle.Render("Primary IP:"),
			ValueStyle.Render(rep.TargetIP),
			SubtleStyle.Render("Open Ports:"),
			lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Render(fmt.Sprintf("%d/%d", rep.OpenPortsCount, rep.PortsScanned)),
			SubtleStyle.Render("Security Grade:"),
			lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render(fmt.Sprintf("%s (%d/100)", rep.SecurityGrade, rep.SecurityScore)),
			SubtleStyle.Render("Duration:"),
			ValueStyle.Render(fmt.Sprintf("%dms", rep.DurationMs)),
		)
		fmt.Println(CardStyle.Render(summary))

		fmt.Println(SectionHeaderStyle.Render(" PARALLEL TCP PORT MATRIX "))
		for _, p := range rep.Ports {
			status := SubtleStyle.Render("CLOSED")
			if p.Open {
				status = BadgeAvailable.Render(" OPEN ")
			}
			fmt.Printf("  %-8s %-14s %-18s %-10s %s\n",
				ValueStyle.Render(fmt.Sprintf(":%d", p.Port)),
				status,
				p.Service,
				SubtleStyle.Render(fmt.Sprintf("%dms", p.LatencyMs)),
				SubtleStyle.Render(p.Category),
			)
		}

		if rep.TLS.Supported {
			fmt.Println(SectionHeaderStyle.Render(" TLS HANDSHAKE & CERTIFICATE "))
			fmt.Printf("  %s %s (%s)\n", LabelStyle.Render("Protocol:"), ValueStyle.Render(rep.TLS.Version), rep.TLS.CipherSuite)
			fmt.Printf("  %s %s\n", LabelStyle.Render("Issuer:"), ValueStyle.Render(rep.TLS.Issuer))
			fmt.Printf("  %s %s → %s (%d days left)\n", LabelStyle.Render("Validity:"), rep.TLS.ValidFrom, rep.TLS.ValidUntil, rep.TLS.DaysRemaining)
		}

		if len(rep.Subdomains) > 0 {
			fmt.Println(SectionHeaderStyle.Render(" ACTIVE SUBDOMAINS DISCOVERED "))
			for _, sub := range rep.Subdomains {
				fmt.Printf("  • %-28s → %s\n", lipgloss.NewStyle().Foreground(ColorCyan).Render(sub.Subdomain), strings.Join(sub.IPs, ", "))
			}
		}
		fmt.Println()
		return nil
	},
}

var historyCmd = &cobra.Command{
	Use:     "history <domain>",
	Aliases: []string{"wayback", "past"},
	Short:   "Check if a domain was previously registered in the past via Wayback CDX & Certificate Transparency logs",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		fmt.Println(RenderBanner("Mode: Past Registration & Archive History (Wayback + CT)  •  Target: " + target))

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		hist := domain.CheckDomainHistory(ctx, target)

		verdictBadge := BadgeAvailable.Render(" CLEAN VIRGIN DOMAIN ")
		if hist.PreviouslyRegistered {
			verdictBadge = BadgeBrandable.Render(" PREVIOUSLY REGISTERED IN PAST ")
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("%s %s  %s", LabelStyle.Render("Domain:"), ValueStyle.Render(hist.Domain), verdictBadge))
		lines = append(lines, fmt.Sprintf("%s %s", LabelStyle.Render("Verdict:"), ValueStyle.Render(hist.HistoryVerdict)))
		lines = append(lines, fmt.Sprintf("%s %s", LabelStyle.Render("Summary:"), SubtleStyle.Render(hist.SummaryNote)))
		if hist.PreviouslyRegistered {
			lines = append(lines, fmt.Sprintf("%s %s → %s (%d active years)", LabelStyle.Render("Archive Span:"), ValueStyle.Render(hist.FirstSeenAt), ValueStyle.Render(hist.LastSeenAt), len(hist.ActiveYears)))
			lines = append(lines, fmt.Sprintf("%s %d Wayback snapshots  •  %d historical TLS certs", LabelStyle.Render("Evidence:"), hist.WaybackSnapshots, hist.CertCount))
			if len(hist.PastSubdomains) > 0 {
				lines = append(lines, fmt.Sprintf("%s %s", LabelStyle.Render("Past Subdomains:"), SubtleStyle.Render(strings.Join(hist.PastSubdomains, ", "))))
			}
		}
		fmt.Println(CardStyle.Render(strings.Join(lines, "\n")))
		return nil
	},
}
