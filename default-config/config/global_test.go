package config

import (
	"context"
	"testing"

	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_cesGlobalConfigWriter_applyDefaultGlobalConfig(t *testing.T) {
	testCtx := context.Background()
	t.Run("should apply default global config", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 2)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			return cfg, nil
		})

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should not apply config key if already exists", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		existingConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))
		newExisting, err := existingConfig.Set("foo", "alreadyExists")
		require.NoError(t, err)
		existingConfig = regLibConfig.GlobalConfig{Config: newExisting}

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(existingConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 2)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "alreadyExists", val.String())

			return cfg, nil
		})

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err = gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should create new global config if not exists", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, cesLibErr.NewNotFoundError(assert.AnError))
		mockRepo.EXPECT().Create(testCtx, emptyConfig).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 2)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			return cfg, nil
		})

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should fail to apply default global config on error getting config", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, assert.AnError)

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "error reading global config")
	})

	t.Run("should fail to apply default global config on error creating config", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, cesLibErr.NewNotFoundError(assert.AnError))
		mockRepo.EXPECT().Create(testCtx, emptyConfig).Return(emptyConfig, assert.AnError)

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "error creating new global config")
	})

	t.Run("should fail to apply default global config on error saving config", func(t *testing.T) {
		defaultConfig := map[string]string{
			"key": "value",
			"foo": "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 2)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			return cfg, assert.AnError
		})

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to save global config")
	})
}
