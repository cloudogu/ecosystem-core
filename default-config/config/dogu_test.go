package config

import (
	"context"
	"testing"

	cesLibDogu "github.com/cloudogu/ces-commons-lib/dogu"
	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_applyDefaultsForRepo(t *testing.T) {
	testCtx := context.Background()

	defaultDoguConfig := map[string]map[string]string{
		"ldap": {
			"key": "value",
			"foo": "bar",
		},
		"cas": {
			"other": "thing",
		},
	}

	t.Run("should apply default global config", func(t *testing.T) {
		emptyLdapConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))
		emptyCasConfig := regLibConfig.CreateDoguConfig("cas", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("ldap")).Return(emptyLdapConfig, nil)
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error) {
			if cfg.DoguName.String() == "ldap" {
				assert.Len(t, cfg.GetAll(), 2)
				val, exists := cfg.Get("key")
				assert.True(t, exists)
				assert.Equal(t, "value", val.String())

				val, exists = cfg.Get("foo")
				assert.True(t, exists)
				assert.Equal(t, "bar", val.String())
			} else if cfg.DoguName.String() == "cas" {
				assert.Len(t, cfg.GetAll(), 1)
				val, exists := cfg.Get("other")
				assert.True(t, exists)
				assert.Equal(t, "thing", val.String())
			} else {
				t.Errorf("unexpected dogu name: %s", cfg.DoguName.String())
			}

			return cfg, nil
		})

		err := applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.NoError(t, err)
	})

	t.Run("should not apply config key if already exists", func(t *testing.T) {
		existingLdapConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))
		newExisting, err := existingLdapConfig.Set("foo", "alreadyExists")
		require.NoError(t, err)
		existingLdapConfig = regLibConfig.DoguConfig{Config: newExisting}

		emptyCasConfig := regLibConfig.CreateDoguConfig("cas", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("ldap")).Return(existingLdapConfig, nil)
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error) {
			if cfg.DoguName.String() == "ldap" {
				assert.Len(t, cfg.GetAll(), 2)
				val, exists := cfg.Get("key")
				assert.True(t, exists)
				assert.Equal(t, "value", val.String())

				val, exists = cfg.Get("foo")
				assert.True(t, exists)
				assert.Equal(t, "alreadyExists", val.String())
			} else if cfg.DoguName.String() == "cas" {
				assert.Len(t, cfg.GetAll(), 1)
				val, exists := cfg.Get("other")
				assert.True(t, exists)
				assert.Equal(t, "thing", val.String())
			} else {
				t.Errorf("unexpected dogu name: %s", cfg.DoguName.String())
			}

			return cfg, nil
		})

		err = applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.NoError(t, err)
	})

	t.Run("should create new global config if not exists", func(t *testing.T) {
		emptyLdapConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))
		emptyCasConfig := regLibConfig.CreateDoguConfig("cas", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("ldap")).Return(emptyLdapConfig, cesLibErr.NewNotFoundError(assert.AnError))
		mockRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, nil)
		mockRepo.EXPECT().Create(testCtx, emptyLdapConfig).Return(emptyLdapConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error) {
			if cfg.DoguName.String() == "ldap" {
				assert.Len(t, cfg.GetAll(), 2)
				val, exists := cfg.Get("key")
				assert.True(t, exists)
				assert.Equal(t, "value", val.String())

				val, exists = cfg.Get("foo")
				assert.True(t, exists)
				assert.Equal(t, "bar", val.String())
			} else if cfg.DoguName.String() == "cas" {
				assert.Len(t, cfg.GetAll(), 1)
				val, exists := cfg.Get("other")
				assert.True(t, exists)
				assert.Equal(t, "thing", val.String())
			} else {
				t.Errorf("unexpected dogu name: %s", cfg.DoguName.String())
			}

			return cfg, nil
		})

		err := applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.NoError(t, err)
	})

	t.Run("should fail to apply default global config on error getting config", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, mock.Anything).Return(emptyConfig, assert.AnError)

		err := applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "error reading dogu config")
	})

	t.Run("should fail to apply default global config on error creating config", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, mock.Anything).Return(emptyConfig, cesLibErr.NewNotFoundError(assert.AnError))
		mockRepo.EXPECT().Create(testCtx, mock.Anything).Return(emptyConfig, assert.AnError)

		err := applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "error creating new dogu config for dogu")
	})

	t.Run("should fail to apply default global config on error saving config", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))

		mockRepo := newMockDoguConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx, mock.Anything).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.DoguConfig) (regLibConfig.DoguConfig, error) {
			if cfg.DoguName.String() == "ldap" {
				assert.Len(t, cfg.GetAll(), 2)
				val, exists := cfg.Get("key")
				assert.True(t, exists)
				assert.Equal(t, "value", val.String())

				val, exists = cfg.Get("foo")
				assert.True(t, exists)
				assert.Equal(t, "bar", val.String())
			} else if cfg.DoguName.String() == "cas" {
				assert.Len(t, cfg.GetAll(), 1)
				val, exists := cfg.Get("other")
				assert.True(t, exists)
				assert.Equal(t, "thing", val.String())
			} else {
				t.Errorf("unexpected dogu name: %s", cfg.DoguName.String())
			}

			return cfg, assert.AnError
		})

		err := applyDefaultsForRepo(testCtx, defaultDoguConfig, mockRepo)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to save new dogu config for dogu")
	})
}

func Test_cesDoguConfigWriter_applyDefaultDoguConfig(t *testing.T) {
	testCtx := context.Background()

	defaultDoguConfig := map[string]map[string]string{
		"cas": {
			"other": "thing",
		},
	}

	defaultSensitiveDoguConfig := map[string]map[string]string{
		"ldap": {
			"secret": "value",
		},
	}

	emptyCasConfig := regLibConfig.CreateDoguConfig("cas", make(regLibConfig.Entries))
	emptyLdapConfig := regLibConfig.CreateDoguConfig("ldap", make(regLibConfig.Entries))

	t.Run("should apply default dogu & sensitive config", func(t *testing.T) {
		mockDoguRepo := newMockDoguConfigRepo(t)
		mockDoguRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, nil)
		mockDoguRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).Return(emptyCasConfig, nil)

		mockSensitiveDoguRepo := newMockDoguConfigRepo(t)
		mockSensitiveDoguRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("ldap")).Return(emptyLdapConfig, nil)
		mockSensitiveDoguRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).Return(emptyLdapConfig, nil)

		dcw := cesDoguConfigWriter{
			doguConfigRepo:          mockDoguRepo,
			sensitiveDoguConfigRepo: mockSensitiveDoguRepo,
		}

		err := dcw.applyDefaultDoguConfig(testCtx, defaultDoguConfig, defaultSensitiveDoguConfig)

		require.NoError(t, err)
	})

	t.Run("should fail to apply default dogu & sensitive config on error in dogu config", func(t *testing.T) {
		mockDoguRepo := newMockDoguConfigRepo(t)
		mockDoguRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, assert.AnError)

		mockSensitiveDoguRepo := newMockDoguConfigRepo(t)

		dcw := cesDoguConfigWriter{
			doguConfigRepo:          mockDoguRepo,
			sensitiveDoguConfigRepo: mockSensitiveDoguRepo,
		}

		err := dcw.applyDefaultDoguConfig(testCtx, defaultDoguConfig, defaultSensitiveDoguConfig)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply default dogu config")
	})

	t.Run("should fail to apply default dogu & sensitive config on error in sensitive dogu config", func(t *testing.T) {
		mockDoguRepo := newMockDoguConfigRepo(t)
		mockDoguRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("cas")).Return(emptyCasConfig, nil)
		mockDoguRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).Return(emptyCasConfig, nil)

		mockSensitiveDoguRepo := newMockDoguConfigRepo(t)
		mockSensitiveDoguRepo.EXPECT().Get(testCtx, cesLibDogu.SimpleName("ldap")).Return(emptyLdapConfig, assert.AnError)

		dcw := cesDoguConfigWriter{
			doguConfigRepo:          mockDoguRepo,
			sensitiveDoguConfigRepo: mockSensitiveDoguRepo,
		}

		err := dcw.applyDefaultDoguConfig(testCtx, defaultDoguConfig, defaultSensitiveDoguConfig)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply default sensitive dogu config")
	})
}
