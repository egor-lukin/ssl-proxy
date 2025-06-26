package internal

import (
	"github.com/spf13/afero"
	"log"
	"path/filepath"
	"time"
)

func RunProxy(duration time.Duration, certsPath, nginxSettingsPath string) {
	appFs := afero.NewOsFs()
	localServersFetcher := NewNginxLocalServersFetcher(certsPath, nginxSettingsPath)
	remoteServersFetcher := NewKubeServersFetcher(certsPath)

	for {
		localServers := localServersFetcher.Read()
		remoteServers := remoteServersFetcher.Read()

		changedServers := SelectChangedServers(localServers, remoteServers)
		removedServers := SelectRemovedServers(localServers, remoteServers)

		if len(removedServers) > 0 {
			for _, s := range removedServers {
				path := filepath.Join(nginxSettingsPath, s.Domain)
				if err := appFs.Remove(path); err != nil {
					log.Printf("Failed to remove server config for domain %s: %v", s.Domain, err)
				} else {
					log.Printf("Removed server config for domain %s", s.Domain)
				}
			}
		}

		if len(changedServers) > 0 {
			log.Println("Has new or updated server settings")
			for _, s := range changedServers {
				certInfo := " (without certs)"
				if s.Cert.Certificate != "" && s.Cert.PrivateKey != "" {
					certInfo = " (with certs)"
				}
				log.Printf("Changed server: domain=%s, snippet=%s, cert=%s\n", s.Domain, s.Snippets, certInfo)
			}

			err := UpdateNginxConfig(appFs, changedServers, certsPath, nginxSettingsPath)
			if err != nil {
				log.Printf("Failed update nginx: %v", err)
			} else {
				localServers = remoteServers
			}
		}

		time.Sleep(duration)
	}
}
