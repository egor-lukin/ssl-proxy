package internal

import (
	"github.com/spf13/afero"
	"path/filepath"
	"testing"
)

func TestUpdateNginxConfig_Afero(t *testing.T) {
	fs := afero.NewMemMapFs()
	certsPath := "/certs"
	nginxPath := "/nginx"
	fs.MkdirAll(certsPath, 0755)
	fs.MkdirAll(nginxPath, 0755)

	servers := []Server{
		{
			Domain:   "example.com",
			Snippets: "server { listen 80; }",
			Cert: Cert{
				Domain:      "example.com",
				Certificate: "dummy-cert",
				PrivateKey:  "dummy-key",
			},
		},
	}

	UpdateNginxConfig(fs, servers, certsPath, nginxPath)

	// Check cert files
	crtFile := filepath.Join(certsPath, "example.com.crt")
	keyFile := filepath.Join(certsPath, "example.com.key")
	crtData, err := afero.ReadFile(fs, crtFile)
	if err != nil || string(crtData) != "dummy-cert" {
		t.Errorf("Cert file not written correctly: %v, %s", err, string(crtData))
	}
	keyData, err := afero.ReadFile(fs, keyFile)
	if err != nil || string(keyData) != "dummy-key" {
		t.Errorf("Key file not written correctly: %v, %s", err, string(keyData))
	}

	// Check snippet file
	snippetFile := filepath.Join(nginxPath, "example.com")
	snippetData, err := afero.ReadFile(fs, snippetFile)
	if err != nil {
		t.Errorf("Snippet file not created: %v", err)
	}
	expectedSnippet := "server { listen 80; }"
	if string(snippetData) != expectedSnippet {
		t.Errorf("Snippet file content mismatch: got %q, want %q", string(snippetData), expectedSnippet)
	}
}
