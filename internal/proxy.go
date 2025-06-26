package internal

import (
	"time"
)

func RunProxy(duration, certsPath, nginxSettingsPath string, fetcher ServersFetcher) {
	servers = fetcher.Read()
	UpdateNginxConfig(servers, certsPath, proxySettingsPath)

	for {
		servers = fetcher.Read()
		UpdateNginxConfig(servers, certsPath, proxySettingsPath)

		time.Sleep(duration)
	}
}
