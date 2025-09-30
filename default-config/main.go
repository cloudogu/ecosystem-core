package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/cloudogu/ecosystem-core/default-config/config"
	"github.com/cloudogu/ecosystem-core/default-config/fqdn"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const defaultWaitTimeoutMinutes = 5

type configApplier interface {
	ApplyDefaultConfig(ctx context.Context) error
}

type fqdnApplier interface {
	ApplyInitialFQDN(ctx context.Context, timeout time.Duration) error
}

func main() {
	ctx := context.Background()
	cfg := readConfig()

	err := run(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to run default-comfig: %w", err))
	}

	slog.Info("exiting")
}

func run(ctx context.Context, cfg jobConfig) error {
	configureLogger(cfg.logLevel)

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

	ca := config.NewDefaultConfigApplier(globalConfigRepo, doguConfigRepo, sensitiveDoguConfigRepo, k8sSecretClient)
	fa := fqdn.NewApplier(globalConfigRepo, k8sServicesClient)

	if err = applyDefaults(ctx, cfg, ca, fa); err != nil {
		return fmt.Errorf("failed to apply default config: %w", err)
	}

	return nil
}

func applyDefaults(ctx context.Context, cfg jobConfig, configApplier configApplier, fqdnApplier fqdnApplier) error {
	if err := configApplier.ApplyDefaultConfig(ctx); err != nil {
		return fmt.Errorf("failed to apply default config: %w", err)
	}

	if err := fqdnApplier.ApplyInitialFQDN(ctx, cfg.waitTimeout); err != nil {
		return fmt.Errorf("failed to apply intial fqdn: %w", err)
	}

	return nil
}

func configureLogger(logLevel string) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(logLevel))
	if err != nil {
		slog.Error("error parsing log level. Setting log level to INFO.", "err", err)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}))
	slog.SetDefault(logger)

	slog.Info("configured logger", "level", level.String())
}

type jobConfig struct {
	namespace   string
	logLevel    string
	waitTimeout time.Duration
}

func readConfig() jobConfig {
	waitTimeoutMinutes, err := strconv.Atoi(os.Getenv("WAIT_TIMEOUT_MINUTES"))
	if err != nil {
		slog.Warn("failed to parse WAIT_TIMEOUT_MINUTES. Using default value.", "err", err, "defaultWaitTimeoutMinutes", defaultWaitTimeoutMinutes)
		waitTimeoutMinutes = defaultWaitTimeoutMinutes
	}

	return jobConfig{
		namespace:   os.Getenv("NAMESPACE"),
		logLevel:    os.Getenv("LOG_LEVEL"),
		waitTimeout: time.Duration(waitTimeoutMinutes) * time.Minute,
	}
}
