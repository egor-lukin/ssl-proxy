package internal

import (
	"github.com/spf13/afero"
	"os/exec"
	"path/filepath"
)

func UpdateNginxConfig(fs afero.Fs, servers []Server, certsPath, nginxServersPath string) error {
	var err error
	for _, server := range servers {
		if err := saveCertFs(fs, server.Domain, server.Cert, certsPath); err != nil {
			return err
		}
		if err = saveSnippetFs(fs, server.Domain, server.Snippets, certsPath, nginxServersPath); err != nil {
			return err
		}
	}

	cmd := exec.Command("nginx", "-s", "reload")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	} else {
		return nil
	}
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
	return afero.WriteFile(fs, path, []byte(snippet), 0644)
}
