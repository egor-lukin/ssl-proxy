package cmd

import (
	"github.com/egor-lukin/ssl-proxy/internal"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var certsPath string
var nginxSettingsPath string
var interval string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		duration, err := time.ParseDuration(interval)
		if err != nil {
			log.Fatalf("Invalid interval: %v", err)
			return
		}

		internal.RunProxy(duration, certsPath, nginxSettingsPath)
	},
}

func init() {
	runCmd.Flags().StringVar(&certsPath, "certsPath", "./certs", "Path to certs directory")
	runCmd.Flags().StringVar(&nginxSettingsPath, "nginxSettingsPath", "/etc/nginx/sites-enabled", "Path to nginx settings directory")
	runCmd.Flags().StringVar(&interval, "interval", "5s", "Interval to repeat fetching (e.g. 10s, 1m). Default: 5s.")
}

func Execute() {
	if err := runCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
