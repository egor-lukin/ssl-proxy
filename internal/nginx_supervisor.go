package internal

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
	"os/exec"
)

func UpdateNginxConfig(servers []Server, certsPath, nginxServersPath string) {
	for _, server := range servers {
		saveCert(server.Domain, server.Cert, certsPath)
		saveSnippet(server.Domain, server.Snippet, certsPath, nginxServersPath)
	}

	cmd := exec.Command("systemctl", "restart", "nginx")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to restart nginx: %v, output: %s", err, string(output))
	} else {
		log.Printf("nginx restarted successfully: %s", string(output))
	}
}

func saveCert(domain string, cert Cert, certsPath string) {
	crtPath := filepath.Join(certsPath, fmt.Sprintf("%s.crt", domain))
	keyPath := filepath.Join(certsPath, fmt.Sprintf("%s.key", domain))

	if err := os.WriteFile(crtPath, cert.Certificate, 0600); err != nil {
		fmt.Printf("Failed to write %s: %v\n", crtPath, err)
	}
	if err := os.WriteFile(keyPath, cert.PrivateKey, 0600); err != nil {
		fmt.Printf("Failed to write %s: %v\n", keyPath, err)
	}
	fmt.Printf("Saved %s and %s\n", crtPath, keyPath)
}

func saveSnippet(domain, snippet, certsPath, nginxServersPath string) error {
	path := filepath.Join(nginxServersPath, domain)
	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		fmt.Printf("Failed to create nginx challenge config for %s: %v\n", domain, err)
		return err
	}

	tmpl, err := template.New("nginx").Parse(snippet)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		return err
	}

	data = {
		CertPath: certsPath,
	}

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Failed to render template for %s: %v\n", domain, err)
		return err
	}

	fmt.Printf("Generated nginx config: %s\n", path)
}
