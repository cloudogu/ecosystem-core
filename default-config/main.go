package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cloudogu/ecosystem-core/default-config/config"
	"github.com/cloudogu/ecosystem-core/default-config/fqdn"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultWaitTimeout = time.Minute * 5

var waitTimeout = defaultWaitTimeout

type configApplier interface {
	ApplyDefaultConfig(ctx context.Context) error
}

type fqdnApplier interface {
	ApplyInitialFQDN(ctx context.Context, timeout time.Duration) error
}

func main() {
	ctx := context.Background()

	err := run(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to run default-comfig: %w", err))
	}

	slog.Info("exiting")
}

func run(ctx context.Context) error {
	namespace := os.Getenv("NAMESPACE")
	slog.Info("starting applying default-configs...", "namespace", namespace)

	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to read kube config: %w", err)
	}

	k8sClientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to create k8s client set: %w", err)
	}

	k8sConfigMapClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
	k8sSecretClient := k8sClientSet.CoreV1().Secrets(namespace)
	k8sServicesClient := k8sClientSet.CoreV1().Services(namespace)

	globalConfigRepo := repository.NewGlobalConfigRepository(k8sConfigMapClient)
	doguConfigRepo := repository.NewDoguConfigRepository(k8sConfigMapClient)
	sensitiveDoguConfigRepo := repository.NewSensitiveDoguConfigRepository(k8sSecretClient)

	ca := config.NewDefaultConfigApplier(globalConfigRepo, doguConfigRepo, sensitiveDoguConfigRepo)
	fa := fqdn.NewApplier(globalConfigRepo, k8sServicesClient)

	if err = applyDefaults(ctx, ca, fa); err != nil {
		return fmt.Errorf("failed to apply default config: %w", err)
	}

	return nil
}

func applyDefaults(ctx context.Context, configApplier configApplier, fqdnApplier fqdnApplier) error {
	if err := configApplier.ApplyDefaultConfig(ctx); err != nil {
		return fmt.Errorf("failed to apply default config: %w", err)
	}

	if err := fqdnApplier.ApplyInitialFQDN(ctx, waitTimeout); err != nil {
		return fmt.Errorf("failed to apply intial fqdn: %w", err)
	}

	return nil
}
