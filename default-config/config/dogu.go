package config

import (
	"context"
	"fmt"
	"log/slog"

	cesLibDogu "github.com/cloudogu/ces-commons-lib/dogu"
	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
)

type doguConfigRepo interface {
	Get(ctx context.Context, name cesLibDogu.SimpleName) (regLibConfig.DoguConfig, error)
	Create(ctx context.Context, doguConfig regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error)
	SaveOrMerge(ctx context.Context, doguConfig regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error)
}

type doguConfigWriter struct {
	doguConfigRepo          doguConfigRepo
	sensitiveDoguConfigRepo doguConfigRepo
}

func (dcw *doguConfigWriter) applyDefaultDoguConfig(ctx context.Context, defaultDoguConfig map[string]map[string]string, sensitiveDefaultDoguConfig map[string]map[string]string) error {
	slog.Info("Applying default dogu config...")
	if err := applyDefaultsForRepo(ctx, defaultDoguConfig, dcw.doguConfigRepo); err != nil {
		return fmt.Errorf("failed to apply default dogu config: %w", err)
	}

	slog.Info("Applying default sensitive dogu config...")
	if err := applyDefaultsForRepo(ctx, sensitiveDefaultDoguConfig, dcw.sensitiveDoguConfigRepo); err != nil {
		return fmt.Errorf("failed to apply default sensitive dogu config: %w", err)
	}

	return nil
}

func applyDefaultsForRepo(ctx context.Context, defaultDoguConfig map[string]map[string]string, repo doguConfigRepo) error {
	for dogu, doguDefaultConfig := range defaultDoguConfig {
		slog.Info("Applying default dogu config...", "dogu", dogu)

		doguName := cesLibDogu.SimpleName(dogu)
		doguConfig, err := repo.Get(ctx, doguName)
		if err != nil {
			if !cesLibErr.IsNotFoundError(err) {
				return fmt.Errorf("error reading dogu config for dogu %q: %w", dogu, err)
			}

			doguConfig, err = repo.Create(ctx, regLibConfig.CreateDoguConfig(doguName, make(regLibConfig.Entries)))
			if err != nil {
				return fmt.Errorf("error creating new dogu config for dogu %q: %w", dogu, err)
			}
		}

		for key, value := range doguDefaultConfig {
			cKey := regLibConfig.Key(key)
			cValue := regLibConfig.Value(value)

			_, exists := doguConfig.Get(cKey)
			if exists {
				slog.Debug("Dogu config key already exists. Skipping...", "dogu", dogu, "key", cKey.String())
				continue
			}

			slog.Debug("Setting dogu config key", "dogu", dogu, "key", cKey.String())
			newDoguConfig, err := doguConfig.Set(cKey, cValue)
			if err != nil {
				return fmt.Errorf("failed to set dogu config key %q for dogu %q: %w", cKey, dogu, err)
			}

			doguConfig = regLibConfig.DoguConfig{
				DoguName: doguName,
				Config:   newDoguConfig,
			}

		}

		_, err = repo.SaveOrMerge(ctx, doguConfig)
		if err != nil {
			return fmt.Errorf("failed to save new dogu config for dogu %q: %w", dogu, err)
		}

		slog.Info("...Successfully applied default-values to dogu config.", "dogu", dogu)
	}
	return nil
}
