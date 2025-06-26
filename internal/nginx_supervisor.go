package internal

import (
	"path/filepath"
	"text/template"
	"bytes"
	// "os/exec"
	"github.com/spf13/afero"
)

func UpdateNginxConfig(fs afero.Fs, servers []Server, certsPath, nginxServersPath string) {
	for _, server := range servers {
		saveCertFs(fs, server.Domain, server.Cert, certsPath)
		saveSnippetFs(fs, server.Domain, server.Snippets, certsPath, nginxServersPath)
	}
	// Omit systemctl restart for testability
}

func saveCertFs(fs afero.Fs, domain string, cert Cert, certsPath string) error {
	crtPath := filepath.Join(certsPath, domain+".crt")
	keyPath := filepath.Join(certsPath, domain+".key")
	if err := afero.WriteFile(fs, crtPath, []byte(cert.Certificate), 0600); err != nil {
		return err
	}
	if err := afero.WriteFile(fs, keyPath, []byte(cert.PrivateKey), 0600); err != nil {
		return err
	}
	return nil
}

func saveSnippetFs(fs afero.Fs, domain, snippet, certsPath, nginxServersPath string) error {
	path := filepath.Join(nginxServersPath, domain)
	tmpl, err := template.New("nginx").Parse(snippet)
	if err != nil {
		return err
	}
	data := struct{ CertPath string }{CertPath: certsPath}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return afero.WriteFile(fs, path, buf.Bytes(), 0644)
}
