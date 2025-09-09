package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/cloudogu/ecosystem-core/default-config/config"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := context.Background()

	namespace := os.Getenv("NAMESPACE")
	slog.Info("starting applying default-configs...", "namespace", namespace)

	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		panic(fmt.Errorf("failed to read kube config: %w", err))
	}

	k8sClientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		slog.Error("failed to create k8s client set", "err", err)
		panic(err)
	}

	k8sConfigMapClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
	k8sSecretClient := k8sClientSet.CoreV1().Secrets(namespace)

	globalConfigRepo := repository.NewGlobalConfigRepository(k8sConfigMapClient)
	doguConfigRepo := repository.NewDoguConfigRepository(k8sConfigMapClient)
	sensitiveDoguConfigRepo := repository.NewSensitiveDoguConfigRepository(k8sSecretClient)

	applier := config.NewDefaultConfigApplier(globalConfigRepo, doguConfigRepo, sensitiveDoguConfigRepo)

	if err := applier.ApplyDefaultConfig(ctx); err != nil {
		slog.Error("failed to apply default config", "err", err)
		panic(err)
	}

	slog.Info("exiting")
}
