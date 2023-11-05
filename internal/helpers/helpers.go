package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetEnv gets value from environment variable or fallbacks to default value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Authenticate is helper to perform authentication to Kubernetes
// using in-cluster config or KUBECONFIG or ~/.kube/config
func Authenticate(logger *zap.Logger) (*kubernetes.Clientset, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		logger.Info("[Auth] Using KUBECONFIG to authenticate", zap.String("kubeconfig", kubeconfigEnv))
		kubeconfig = kubeconfigEnv
	}

	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Warn("[Auth] Cannot create incluster config", zap.Error(err))
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		} else {
			logger.Info("[Auth] Client created from home or KUBECONFIG")
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// CreateSecretFromExisting creates new secret based on sourceSecret
func CreateSecretFromExisting(client *kubernetes.Clientset, logger *zap.Logger, sourceSecret *v1.Secret, name, ns string) error {
	s := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data:      sourceSecret.DeepCopy().Data,
		Type:      sourceSecret.Type,
		Immutable: sourceSecret.Immutable,
	}
	_, err := client.CoreV1().Secrets(ns).Create(context.TODO(), &s, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret %s/%s: %w", s.Namespace, s.Name, err)
	}
	logger.Info("[Auth] New secret was created",
		zap.String("name", s.Name),
		zap.String("namespace", s.Namespace),
	)
	return nil
}

// UpdateExistingSecret updates a secret that already exists
func UpdateExistingSecret(client *kubernetes.Clientset, logger *zap.Logger, sourceSecret *v1.Secret, name, ns string) error {
	s := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data:      sourceSecret.DeepCopy().Data,
		Type:      sourceSecret.Type,
		Immutable: sourceSecret.Immutable,
	}
	_, err := client.CoreV1().Secrets(s.Namespace).Update(context.TODO(), &s, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Warn("cannot update secret because it doesn't exist yet",
				zap.String("name", name),
				zap.String("namespace", ns),
				zap.Error(err),
			)
		} else {
			return fmt.Errorf("failed to update secret %s/%s: %w", ns, name, err)
		}
	}
	logger.Info("Secret was updated",
		zap.String("name", name),
		zap.String("namespace", ns),
	)
	return nil
}
