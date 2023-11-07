package main

import (
	"context"
	"k8s-secret-replicator/internal/helpers"
	"sync"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchapi "k8s.io/apimachinery/pkg/watch"
)

var logger, _ = zap.NewProduction()

// Create secret in new namespaces if it doesn't exist
func watchNamespaces(client *kubernetes.Clientset, watchApi watchapi.Interface, sourceSecret *v1.Secret, sourceSecretName string, wg *sync.WaitGroup) {
	wg.Add(1)
	for event := range watchApi.ResultChan() {
		ns := event.Object.DeepCopyObject().(*v1.Namespace)
		if event.Type == watchapi.Added {
			_, err := client.CoreV1().Secrets(ns.Name).Get(context.TODO(), sourceSecretName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					logger.Info("Secret in namespace doesn't exist yet",
						zap.String("name", sourceSecretName),
						zap.String("namespace", ns.Name),
						zap.Error(err),
					)
					err := helpers.CreateSecretFromExisting(client, logger, sourceSecret, sourceSecretName, ns.Name)
					if err != nil {
						logger.Fatal("", zap.Error(err))
					}
				} else {
					logger.Error("", zap.Error(err))
				}
			}
		}
	}
	wg.Done()
}

// Watch for additions/modifications of source secret
func watchSourceSecret(client *kubernetes.Clientset, watchApi watchapi.Interface, sourceSecretName, sourceSecretNamespace string, wg *sync.WaitGroup) {
	wg.Add(1)
	for event := range watchApi.ResultChan() {
		if event.Type == watchapi.Added || event.Type == watchapi.Modified {
			namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				logger.Error("", zap.Error(err))
			} else {
				for _, ns := range namespaces.Items {
					if ns.Name == sourceSecretNamespace {
						logger.Info("Skip updating secret in source namespace")
						continue
					}
					err = helpers.UpdateExistingSecret(client, logger, event.Object.(*v1.Secret), sourceSecretName, ns.Name)
					if err != nil {
						logger.Error("failed to update secret",
							zap.String("name", sourceSecretName),
							zap.String("namespace", ns.Name),
							zap.Error(err),
						)
					}
				}
			}
		}
	}
	wg.Done()
}

func main() {
	defer logger.Sync() // flushes buffer, if any

	var wg sync.WaitGroup
	sourceSecretName := helpers.GetEnv("SOURCE_SECRET_NAME", "my-secret-to-replicate")
	sourceSecretNamespace := helpers.GetEnv("SOURCE_SECRET_NAMESPACE", "kube-system")
	logger.Info("[Config]", zap.String("source-secret-name", sourceSecretName))
	logger.Info("[Config]", zap.String("source-secret-namespace", sourceSecretNamespace))

	client, err := helpers.Authenticate(logger)
	if err != nil {
		logger.Fatal("[Auth] cannot create config from home or KUBECONFIG", zap.Error(err))
	}
	ctx := context.TODO()

	// Get source secret
	sourceSecret, err := client.CoreV1().Secrets(sourceSecretNamespace).Get(ctx, sourceSecretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Fatal("source secret doesn't exist",
				zap.String("name", sourceSecretName),
				zap.String("namespace", sourceSecretNamespace),
				zap.Error(err),
			)
		}
		logger.Fatal("error getting source secret",
			zap.String("name", sourceSecretName),
			zap.String("namespace", sourceSecretNamespace),
			zap.Error(err),
		)
	}

	// Watch for namespace events
	watchNamespacesApi, err := client.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Fatal("cannot watch namespaces", zap.Error(err))
	}
	defer watchNamespacesApi.Stop()
	go watchNamespaces(client, watchNamespacesApi, sourceSecret, sourceSecretName, &wg)

	// Watch for source secret events
	sourceSecretFs := fields.OneTermEqualSelector("metadata.name", sourceSecretName)
	watchSourceSecretApi, err := client.CoreV1().Secrets(sourceSecretNamespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: sourceSecretFs.String(),
	})
	if err != nil {
		logger.Fatal("cannot watch source secret",
			zap.String("name", sourceSecretName),
			zap.String("namespace", sourceSecretNamespace),
			zap.Error(err),
		)
	}
	defer watchSourceSecretApi.Stop()
	go watchSourceSecret(client, watchSourceSecretApi, sourceSecretName, sourceSecretNamespace, &wg)

	wg.Wait()
}
