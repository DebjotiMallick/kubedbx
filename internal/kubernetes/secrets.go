package kubernetes

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewClient initializes and returns a Kubernetes client
func NewClient() (*kubernetes.Clientset, error) {
	// In-cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetSecret retrieves a Kubernetes secret from a given namespace
func GetSecret(clientset *kubernetes.Clientset, namespace, secretName string) (*v1.Secret, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s in namespace %s: %w", secretName, namespace, err)
	}
	return secret, nil
}

// ListSecrets lists all secrets in a given namespace
func ListSecrets(clientset *kubernetes.Clientset, namespace string) (*v1.SecretList, error) {
	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in namespace %s: %w", namespace, err)
	}
	return secrets, nil
}

// GetServiceAccountToken retrieves a service account token from a given namespace
func GetServiceAccountToken(clientset *kubernetes.Clientset, namespace, serviceAccountName string) (string, error) {
	sa, err := clientset.CoreV1().ServiceAccounts(namespace).Get(context.Background(), serviceAccountName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service account %s in namespace %s: %w", serviceAccountName, namespace, err)
	}

	if len(sa.Secrets) == 0 {
		return "", fmt.Errorf("service account %s in namespace %s has no secrets", serviceAccountName, namespace)
	}

	secretName := sa.Secrets[0].Name // Assumes the first secret is the token secret.
	secret, err := GetSecret(clientset, namespace, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s in namespace %s: %w", secretName, namespace, err)
	}

	token, ok := secret.Data["token"]
	if !ok {
		return "", fmt.Errorf("token not found in secret %s", secretName)
	}

	return string(token), nil
}

// EnsureNamespaceExists checks if a namespace exists and creates it if it doesn't.
func EnsureNamespaceExists(clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get namespace %s: %w", namespace, err)
		}

		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create namespace %s: %w", namespace, err)
		}
		log.Printf("Namespace %s created", namespace)
	}
	return nil
}
