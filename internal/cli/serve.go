package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rny/scanner/internal/server"
	"github.com/spf13/cobra"
)

var flagPort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server powering the Svelte Domain Intelligence UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf(":%d", flagPort)
		fmt.Println(RenderBanner(fmt.Sprintf("Mode: Single-Binary Web Studio + API Server  •  http://localhost%s", addr)))

		infoBox := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
			LabelStyle.Render("Web Studio UI:"),
			lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Render(fmt.Sprintf("http://localhost%s      (Full Svelte 5 Studio Embedded)", addr)),
			LabelStyle.Render("API Endpoint: "),
			lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Render(fmt.Sprintf("http://localhost%s/api  (Go Parallel Reconnaissance Engine)", addr)),
			LabelStyle.Render("Scan Route:   "),
			ValueStyle.Render("POST /api/scan           (24-TLD discovery & appraisal)"),
			LabelStyle.Render("Inspect Route:"),
			ValueStyle.Render("GET  /api/inspect        (RDAP, Port-43 WHOIS, DNS & age)"),
			LabelStyle.Render("Dossier Suite:"),
			ValueStyle.Render("POST /api/parallel-suite (360° Master Domain Report)"),
		)
		fmt.Println(CardStyle.Render(infoBox))

		srv := server.New(addr)
		return srv.Start()
	},
}

func init() {
	serveCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Port for the HTTP API server")
}
