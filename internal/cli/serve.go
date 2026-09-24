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
		fmt.Println(RenderBanner(fmt.Sprintf("Mode: API Server for Svelte UI  •  Listening on http://localhost%s", addr)))

		infoBox := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n%s %s",
			LabelStyle.Render("API Endpoint:"),
			lipgloss.NewStyle().Bold(true).Foreground(ColorEmerald).Render(fmt.Sprintf("http://localhost%s/api", addr)),
			LabelStyle.Render("Scan Route:"),
			ValueStyle.Render("POST /api/scan     (Available domain discovery & valuation)"),
			LabelStyle.Render("Inspect Route:"),
			ValueStyle.Render("GET  /api/inspect  (RDAP registration dates, age & DNS)"),
			LabelStyle.Render("Crawl Route:"),
			ValueStyle.Render("POST /api/crawl    (Site hierarchy & structure tree)"),
		)
		fmt.Println(CardStyle.Render(infoBox))

		srv := server.New(addr)
		return srv.Start()
	},
}

func init() {
	serveCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Port for the HTTP API server")
}
