package config

import (
	"context"
	"fmt"
	"log/slog"

	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
)

type globalConfigRepo interface {
	Get(ctx context.Context) (regLibConfig.GlobalConfig, error)
	Create(ctx context.Context, globalConfig regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error)
	SaveOrMerge(ctx context.Context, globalConfig regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error)
}

type cesGlobalConfigWriter struct {
	globalConfigRepo globalConfigRepo
}

func (gcw *cesGlobalConfigWriter) applyDefaultGlobalConfig(ctx context.Context, defaultGlobalConfig map[string]string) error {
	slog.Info("Applying default global config...")

	globalConfig, err := gcw.globalConfigRepo.Get(ctx)
	if err != nil {
		if !cesLibErr.IsNotFoundError(err) {
			return fmt.Errorf("error reading global config: %w", err)
		}

		globalConfig, err = gcw.globalConfigRepo.Create(ctx, regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries)))
		if err != nil {
			return fmt.Errorf("error creating new global config: %w", err)
		}
	}

	for key, value := range defaultGlobalConfig {
		cKey := regLibConfig.Key(key)
		cValue := regLibConfig.Value(value)

		_, exists := globalConfig.Get(cKey)
		if exists {
			slog.Info("Global config key already exists. Skipping...", "key", cKey.String())
			continue
		}

		slog.Info("Setting global config key", "key", cKey.String())
		newGlobalConfig, err := globalConfig.Set(cKey, cValue)
		if err != nil {
			return fmt.Errorf("failed to set global config key %s: %w", cKey, err)
		}

		globalConfig = regLibConfig.GlobalConfig{Config: newGlobalConfig}

	}

	_, err = gcw.globalConfigRepo.SaveOrMerge(ctx, globalConfig)
	if err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	slog.Info("...Successfully applied default-values to global config.")

	return nil
}
