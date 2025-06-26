package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// networkingv1 "k8s.io/api/networking/v1"
)

struct Cert {
	Domain  string
	PrivateKey string
	Certificate string
}

struct Server {
	Domain   string
	Snippets string
	Cert     Cert
}

const namespace = 'website-builder'

type ServersFetcher interface {
	Read() []Server
}

type KubeServersFetcher struct {

	clientset *kubernetes.Clientset
}

func NewKubeServersFetcher(clientset *kubernetes.Clientset) *KubeServersFetcher {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Failed to load kubeconfig: %v\n", err)
		return
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %v\n", err)
		return
	}

	return &KubeServersFetcher{clientset: clientset}
}

func (fetcher *KubeServersFetcher) Read() []Server {
	ingresses, err := fetcher.clientset.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list ingresses: %v\n", err)
		return nil
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list secrets: %v\n", err)
		return
	}

	var servers []Server
	for _, ingress := range ingresses.Items {
		annotations := ingress.Annotations
		domain, hasDomain := annotations["external-proxy/domain"]
		serverSnippets, hasServerSnippets := annotations["external-proxy/server-snippets"]
		if !hasDomain || !hasServerSnippets {
			continue
		}

		server := Server{
			Domain:   domain,
			Snippets: serverSnippets,
		}

		for _, secret := range secrets.Items {
			if secret.Type == "kubernetes.io/tls" {
				domain, ok := secret.Annotations["external-proxy/domain"];
				if !ok || externalDomain == "" {
					continue
				}
				crt, crtOk := secret.Data["tls.crt"]
				key, keyOk := secret.Data["tls.key"]
				if !crtOk || !keyOk {
					fmt.Printf("Secret %s missing tls.crt or tls.key\n", externalDomain)
					continue
				}

				cert = Cert{
					Domain: domain,
					Certificate: crt,
					PrivateKey: key,
				}

				server.Cert = cert
			}
		}

		servers = append(servers, server)
	}

	return servers
}
