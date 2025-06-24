package cmd

import (
	"database/sql"
	"github.com/egor-lukin/ssl-proxy/internal"
	"github.com/egor-lukin/ssl-proxy/internal/letsencrypt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"log"
)

const dbFilePath = "./settings.db"

var rootCmd = &cobra.Command{
	Use:   "ssl-proxy",
	Short: "SSL Proxy CLI",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Prepare the proxy",
	Run: func(cmd *cobra.Command, args []string) {
		proxy, _ := cmd.Flags().GetString("proxy")
		destination, _ := cmd.Flags().GetString("destination")
		email, _ := cmd.Flags().GetString("email")
		sslGenerator := letsencrypt.LetsEncryptGenerator{}
		_, err := internal.PrepareProxy(dbFilePath, proxy, destination, email, sslGenerator)
		if err != nil {
			log.Fatalf("failed to prepare proxy: %v", err)
		}
		log.Println("Proxy prepared successfully.")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", dbFilePath)
		if err != nil {
			log.Fatalf("failed to open db: %v", err)
		}
		srv := internal.NewServer(db)
		log.Println("Listening on 443 port")
		log.Fatal(srv.ListenAndServe())
	},
}

func init() {
	initCmd.Flags().String("proxy", "", "Proxy domain")
	initCmd.Flags().String("destination", "", "Destination domain")
	initCmd.Flags().String("email", "", "Email address")
}

func Execute() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
