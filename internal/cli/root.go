package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scanner [keywords...]",
	Short: "All-in-one Domain Availability Scanner, RDAP Inspector & Site Structure Crawler",
	Long: `Scanner is an all-in-one domain intelligence CLI and API server.
By default, running 'scanner [keywords...]' scans for valuable available (unregistered) domains.

Available Modes:
  • scanner [keywords...]          Scan for available domains & score market value (default)
  • scanner inspect <domain>       Deep RDAP inquiry (registration date, age, registrar, DNS)
  • scanner crawl <domain/url>     Crawl website structure & render hierarchical site tree
  • scanner serve                  Launch the REST API server (:8080) for the Svelte UI`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeScan(args)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&flagTLDs, "tlds", "t", "com,ai,io,dev,co,app", "Comma-separated list of TLDs to scan")
	rootCmd.Flags().BoolVarP(&flagMutations, "mutations", "m", true, "Generate brandable prefix/suffix variations")
	rootCmd.Flags().BoolVarP(&flagOnlyAvailable, "available", "a", false, "Show only available (empty) domains")
	rootCmd.Flags().IntVarP(&flagMinScore, "min-score", "s", 0, "Minimum valuation score (0-100)")
	rootCmd.Flags().IntVarP(&flagMaxResults, "limit", "l", 36, "Maximum number of domain candidates to check")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(reconCmd)
	rootCmd.AddCommand(crawlCmd)
	rootCmd.AddCommand(serveCmd)
}

// Execute runs the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
