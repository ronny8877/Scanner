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

var inspectCmd = &cobra.Command{
	Use:     "inspect <domain>",
	Aliases: []string{"whois", "inquire", "info"},
	Short:   "Inquire about a domain's registration date, age, registrar, DNS records, and valuation",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		fmt.Println(RenderBanner("Mode: Deep Domain Inquiry (RDAP + DNS + Valuation)  •  Target: " + target))

		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		info := domain.InspectDomain(ctx, target)

		statusPill := BadgeAvailable.Render(" AVAILABLE FOR REGISTRATION ")
		if !info.Available {
			statusPill = BadgeTaken.Render(" REGISTERED / ACTIVE DOMAIN ")
		}

		var rows []string
		rows = append(rows, fmt.Sprintf("%s %s  %s", LabelStyle.Render("Domain:"), ValueStyle.Render(info.Domain), statusPill))
		rows = append(rows, fmt.Sprintf("%s %s (%d/100)  •  Est. Value: %s",
			LabelStyle.Render("Valuation:"),
			RenderTierBadge(info.Valuation.Tier),
			info.Valuation.Score,
			lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render(info.Valuation.EstimatedDisplay),
		))

		if !info.Available {
			regDate := info.RegisteredAt
			if regDate == "" {
				regDate = "Not disclosed by registry RDAP"
			}
			expDate := info.ExpiresAt
			if expDate == "" {
				expDate = "Not disclosed by registry RDAP"
			}
			ageStr := info.DomainAge
			if ageStr == "" {
				ageStr = "Active"
			}
			registrar := info.Registrar
			if registrar == "" {
				registrar = "Private / Registry Protected"
			}

			rows = append(rows, "")
			rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("── Registration & Lifecycle (RDAP) ──"))
			rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Registered On:"), ValueStyle.Render(regDate)))
			rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Domain Age:"), lipgloss.NewStyle().Bold(true).Foreground(ColorAmber).Render(ageStr)))
			rows = append(rows, fmt.Sprintf("%s %s (%d days remaining)", LabelStyle.Render("Expires On:"), ValueStyle.Render(expDate), info.DaysToExpiry))
			if info.UpdatedAt != "" {
				rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Last Updated:"), SubtleStyle.Render(info.UpdatedAt)))
			}
			rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Registrar:"), ValueStyle.Render(registrar)))
			if len(info.StatusFlags) > 0 {
				rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("EPP Status:"), SubtleStyle.Render(strings.Join(info.StatusFlags, ", "))))
			}
		} else {
			rows = append(rows, "")
			rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Render("★ This domain is currently UNREGISTERED and available to claim!"))
			for _, h := range info.Valuation.Highlights {
				rows = append(rows, "  • "+h)
			}
		}

		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("── DNS Infrastructure ──"))
		rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Nameservers:"), formatSliceOrEmpty(info.Nameservers)))
		rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("IPv4 (A):"), formatSliceOrEmpty(info.DNS.A)))
		rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("IPv6 (AAAA):"), formatSliceOrEmpty(info.DNS.AAAA)))
		rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("Mail (MX):"), formatSliceOrEmpty(info.DNS.MX)))
		if len(info.DNS.TXT) > 0 {
			rows = append(rows, fmt.Sprintf("%s %s", LabelStyle.Render("TXT Records:"), SubtleStyle.Render(fmt.Sprintf("%d record(s) found", len(info.DNS.TXT)))))
		}

		fmt.Println(CardStyle.Render(strings.Join(rows, "\n")))
		return nil
	},
}

func formatSliceOrEmpty(items []string) string {
	if len(items) == 0 {
		return SubtleStyle.Render("None")
	}
	return ValueStyle.Render(strings.Join(items, ", "))
}
