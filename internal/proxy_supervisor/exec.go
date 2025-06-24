package proxy_supervisor

import (
	"log"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/fsnotify/fsnotify"
	"time"
)

func Watch(artifactsPath, proxySettingsPath string) {
	log.Println("Running started sync...")
	Exec(artifactsPath, proxySettingsPath)
	log.Println("Sync Done...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Failed to create watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(artifactsPath)
	if err != nil {
		fmt.Printf("Failed to watch directory %s: %v\n", artifactsPath, err)
		return
	}

	var debounceTimer *time.Timer
	const debounceDuration = 2 * time.Second

	triggerExec := func() {
		log.Println("Artifacts directory was changed")

		Exec(artifactsPath, proxySettingsPath)

		log.Println("Proxy Settings was updated")
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDuration, triggerExec)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

func Exec(artifactsPath, proxySettingsPath string) {
	SyncChallenges(artifactsPath, proxySettingsPath)
	SyncCerts(artifactsPath, proxySettingsPath)
	// restart nginx
}

func SyncChallenges(artifactsPath, proxySettingsPath string) {
	challangesDir := filepath.Join(artifactsPath, "challanges")
	templatePath := "templates/nginx-http-challenge.tmpl"
	outputDir := proxySettingsPath

	files, err := ioutil.ReadDir(challangesDir)
	if err != nil {
		fmt.Printf("Failed to read challenges directory: %v\n", err)
		return
	}

	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Failed to read template: %v\n", err)
		return
	}
	tmpl, err := template.New("nginx").Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".content") {
			continue
		}
		domain := strings.TrimSuffix(file.Name(), ".content")

		challengeContent, err := os.ReadFile(filepath.Join(challangesDir, file.Name()))
		if err != nil {
			fmt.Printf("Failed to read challenge file for %s: %v\n", domain, err)
			continue
		}

		challengePath, err := os.ReadFile(filepath.Join(challangesDir, domain + ".path"))
		if err != nil {
			fmt.Printf("Failed to read challenge path file for %s: %v\n", domain, err)
			continue
		}

		data := map[string]string{
			"Domain":           domain,
			"ChallengePath":    string(challengePath),
			"ChallengeContent": string(challengeContent),
		}

		outputPath := filepath.Join(outputDir, domain + ".challange" + ".conf")
		outFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Failed to create nginx challenge config for %s: %v\n", domain, err)
			continue
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			fmt.Printf("Failed to render template for %s: %v\n", domain, err)
		} else {
			fmt.Printf("Generated nginx config: %s\n", outputPath)
		}
		outFile.Close()
	}
}

func SyncCerts(artifactsPath, proxySettingsPath string) {
	certsDir := filepath.Join(artifactsPath, "certs")
	templatePath := "templates/nginx-reverse-proxy.tmpl"
	outputDir := proxySettingsPath

	files, err := ioutil.ReadDir(certsDir)
	if err != nil {
		fmt.Printf("Failed to read certs directory: %v\n", err)
		return
	}

	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Failed to read template: %v\n", err)
		return
	}
	tmpl, err := template.New("nginx").Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".crt") {
			continue
		}
		domain := strings.TrimSuffix(file.Name(), ".crt")
		crtPath := filepath.Join(certsDir, domain+".crt")
		keyPath := filepath.Join(certsDir, domain+".key")

		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			fmt.Printf("Key file missing for domain %s, skipping\n", domain)
			continue
		}

		data := map[string]string{
			"Domain":        domain,
			"CertPath":      crtPath,
			"KeyPath":       keyPath,
			"BackendServer": "turbositetest.com:443", // You may want to make this configurable
		}

		outputPath := filepath.Join(outputDir, domain+".conf")
		outFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Failed to create nginx config for %s: %v\n", domain, err)
			continue
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			fmt.Printf("Failed to render template for %s: %v\n", domain, err)
		} else {
			fmt.Printf("Generated nginx config: %s\n", outputPath)
		}
		outFile.Close()
	}
}
