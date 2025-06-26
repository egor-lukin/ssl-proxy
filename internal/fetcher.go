package internal

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	// "log"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const namespace = "website-builder"

type ServersFetcher interface {
	Read() []Server
}

type KubeServersFetcher struct {
	clientset *kubernetes.Clientset
	certsPath string
}

type NginxLocalServersFetcher struct {
	serverSettingsPath string
	certsPath          string
}

func NewNginxLocalServersFetcher(certsPath, serverSettingsPath string) *NginxLocalServersFetcher {
	return &NginxLocalServersFetcher{
		serverSettingsPath: serverSettingsPath,
		certsPath:          certsPath,
	}
}

func (fetcher *NginxLocalServersFetcher) Read() []Server {
	var servers []Server

	files, err := os.ReadDir(fetcher.serverSettingsPath)
	if err != nil {
		fmt.Printf("Failed to read server settings dir: %v\n", err)
		return servers
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		domain := file.Name()
		snippetPath := filepath.Join(fetcher.serverSettingsPath, domain)
		snippetBytes, err := os.ReadFile(snippetPath)
		if err != nil {
			fmt.Printf("Failed to read snippet for %s: %v\n", domain, err)
			continue
		}

		crtPath := filepath.Join(fetcher.certsPath, fmt.Sprintf("%s.crt", domain))
		keyPath := filepath.Join(fetcher.certsPath, fmt.Sprintf("%s.key", domain))
		crtBytes, err := os.ReadFile(crtPath)
		if err != nil {
			fmt.Printf("Failed to read cert for %s: %v\n", domain, err)
			continue
		}
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			fmt.Printf("Failed to read key for %s: %v\n", domain, err)
			continue
		}

		server := Server{
			Domain:   domain,
			Snippets: string(snippetBytes),
			Cert: Cert{
				Domain:      domain,
				Certificate: string(crtBytes),
				PrivateKey:  string(keyBytes),
			},
		}
		servers = append(servers, server)
	}

	return servers
}

func NewKubeServersFetcher(certsPath string) *KubeServersFetcher {
	home := os.Getenv("HOME")
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &KubeServersFetcher{clientset: clientset, certsPath: certsPath}
}

func (fetcher *KubeServersFetcher) Read() []Server {
	ingresses, err := fetcher.clientset.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list ingresses: %v\n", err)
		return nil
	}

	secrets, err := fetcher.clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list secrets: %v\n", err)
		return nil
	}

	var servers []Server
	for _, ingress := range ingresses.Items {
		annotations := ingress.Annotations
		domain, hasDomain := annotations["external-proxy/domain"]
		serverSnippets, hasServerSnippets := annotations["external-proxy/server-snippets"]
		if !hasDomain || !hasServerSnippets {
			continue
		}

		tmpl, err := template.New("nginx").Parse(serverSnippets)
		if err != nil {
			continue
			// return nil
			// return err
		}

		data := struct{ CertsPath string }{CertsPath: fetcher.certsPath}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			continue
			// return err
			// return nil
		}

		server := Server{
			Domain:   domain,
			Snippets: string(buf.Bytes()),
		}

		for _, secret := range secrets.Items {
			if secret.Type == "kubernetes.io/tls" {
				secretDomain, ok := secret.Annotations["external-proxy/domain"]
				if !ok || secretDomain != domain {
					continue
				}
				crt, crtOk := secret.Data["tls.crt"]
				key, keyOk := secret.Data["tls.key"]
				if !crtOk || !keyOk {
					fmt.Printf("Secret %s missing tls.crt or tls.key\n", secretDomain)
					continue
				}

				cert := Cert{
					Domain:      secretDomain,
					Certificate: string(crt),
					PrivateKey:  string(key),
				}

				server.Cert = cert
			}
		}

		servers = append(servers, server)
	}

	return servers
}
