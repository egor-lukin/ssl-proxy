package certs_fetcher

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

func Exec(artifactsPath string) {
	FetchCerts(artifactsPath)
	FetchChallenges(artifactsPath)
}

func FetchCerts(artifactsPath string) {
	outputDir := filepath.Join(artifactsPath, "certs")
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

	secrets, err := clientset.CoreV1().Secrets("website-builder").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list secrets: %v\n", err)
		return
	}

	for _, secret := range secrets.Items {
		if secret.Type == "kubernetes.io/tls" {
			externalDomain, ok := secret.Annotations["website-builder/external-domain"];
			if !ok || externalDomain == "" {
				continue
			}
			crt, crtOk := secret.Data["tls.crt"]
			key, keyOk := secret.Data["tls.key"]
			if !crtOk || !keyOk {
				fmt.Printf("Secret %s missing tls.crt or tls.key\n", externalDomain)
				continue
			}

			crtPath := filepath.Join(outputDir, fmt.Sprintf("%s.crt", externalDomain))
			keyPath := filepath.Join(outputDir, fmt.Sprintf("%s.key", externalDomain))
			if err := os.WriteFile(crtPath, crt, 0600); err != nil {
				fmt.Printf("Failed to write %s: %v\n", crtPath, err)
			}
			if err := os.WriteFile(keyPath, key, 0600); err != nil {
				fmt.Printf("Failed to write %s: %v\n", keyPath, err)
			}
			fmt.Printf("Saved %s and %s\n", crtPath, keyPath)
		}
	}
}

func FetchChallenges(artifactsPath string) {
	outputDir := filepath.Join(artifactsPath, "challenges")
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

	ingresses, err := clientset.NetworkingV1().Ingresses("website-builder").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list ingresses: %v\n", err)
		return
	}

	for _, ingress := range ingresses.Items {
		annotations := ingress.Annotations
		challengePath, hasPath := annotations["website-builder/challenge-path"]
		challengeContent, hasContent := annotations["website-builder/challenge-content"]
		if !hasPath || !hasContent {
			continue
		}

		domain := ingress.Name
		if len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != "" {
			domain = ingress.Spec.Rules[0].Host
		}

		pathFile := filepath.Join(outputDir, domain+".path")
		contentFile := filepath.Join(outputDir, domain+".content")

		if err := os.WriteFile(pathFile, []byte(challengePath), 0600); err != nil {
			fmt.Printf("Failed to write %s: %v\n", pathFile, err)
		}
		if err := os.WriteFile(contentFile, []byte(challengeContent), 0600); err != nil {
			fmt.Printf("Failed to write %s: %v\n", contentFile, err)
		}
		fmt.Printf("Saved challenge for %s\n", domain)
	}
}
