package cmd

import (
	"github.com/egor-lukin/ssl-proxy/internal/certs_fetcher"
	"github.com/egor-lukin/ssl-proxy/internal/proxy_supervisor"
	"github.com/spf13/cobra"
	"time"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "ssl-proxy",
	Short: "SSL Proxy CLI",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the proxy server",
}

var artifactsPath string
var proxySettingsPath string

var runProxySupervisorCmd = &cobra.Command{
	Use:   "proxy-supervisor",
	Short: "Run the proxy supervisor watcher",
	Run: func(cmd *cobra.Command, args []string) {
		proxy_supervisor.Watch(artifactsPath, proxySettingsPath)
	},
}

var fetcherArtifactsPath string
var fetcherInterval string

var runFetcherCmd = &cobra.Command{
	Use:   "fetcher",
	Short: "Run the certs and challenges fetcher",
	Run: func(cmd *cobra.Command, args []string) {
		if fetcherInterval == "" {
			certs_fetcher.Exec(fetcherArtifactsPath)
			return
		}
		dur, err := time.ParseDuration(fetcherInterval)
		if err != nil {
			log.Fatalf("Invalid interval: %v", err)
		}
		for {
			certs_fetcher.Exec(fetcherArtifactsPath)
			time.Sleep(dur)
		}
	},
}

func init() {
	runProxySupervisorCmd.Flags().StringVar(&artifactsPath, "artifacts", "./artifacts", "Path to artifacts directory")
	runProxySupervisorCmd.Flags().StringVar(&proxySettingsPath, "proxy-settings", "./proxy-settings", "Path to proxy settings directory")
	runFetcherCmd.Flags().StringVar(&fetcherArtifactsPath, "artifacts", "./artifacts", "Path to artifacts directory")
	runFetcherCmd.Flags().StringVar(&fetcherInterval, "interval", "", "Interval to repeat fetching (e.g. 10m, 1h). If empty, runs once.")
	runCmd.AddCommand(runProxySupervisorCmd)
	runCmd.AddCommand(runFetcherCmd)
}

func Execute() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
