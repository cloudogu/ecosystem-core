package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigApplier_ApplyDefaultConfig(t *testing.T) {
	testCtx := context.Background()
	t.Run("should apply default config", func(t *testing.T) {
		mockPg := newMockPasswordGenerator(t)
		mockPg.EXPECT().generatePassword(passwordLength).Return("password")

		mockGcw := newMockGlobalConfigWriter(t)
		mockGcw.EXPECT().applyDefaultGlobalConfig(testCtx, globalDefaults).Return(nil)

		expectedSensitiveConfig := map[string]map[string]string{
			"ldap": {
				"admin_password": "password",
			},
		}

		mockDcw := newMockDoguConfigWriter(t)
		mockDcw.EXPECT().applyDefaultDoguConfig(testCtx, doguDefaults, expectedSensitiveConfig).Return(nil)

		dca := &DefaultConfigApplier{
			passwordGenerator:  mockPg,
			globalConfigWriter: mockGcw,
			doguConfigWriter:   mockDcw,
		}

		err := dca.ApplyDefaultConfig(testCtx)

		require.NoError(t, err)
	})

	t.Run("should fail to apply default global config", func(t *testing.T) {
		mockPg := newMockPasswordGenerator(t)

		mockGcw := newMockGlobalConfigWriter(t)
		mockGcw.EXPECT().applyDefaultGlobalConfig(testCtx, globalDefaults).Return(assert.AnError)

		mockDcw := newMockDoguConfigWriter(t)

		dca := &DefaultConfigApplier{
			passwordGenerator:  mockPg,
			globalConfigWriter: mockGcw,
			doguConfigWriter:   mockDcw,
		}

		err := dca.ApplyDefaultConfig(testCtx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply default global config:")
	})

	t.Run("should fail to apply default dogu config", func(t *testing.T) {
		mockPg := newMockPasswordGenerator(t)
		mockPg.EXPECT().generatePassword(passwordLength).Return("password")

		mockGcw := newMockGlobalConfigWriter(t)
		mockGcw.EXPECT().applyDefaultGlobalConfig(testCtx, globalDefaults).Return(nil)

		expectedSensitiveConfig := map[string]map[string]string{
			"ldap": {
				"admin_password": "password",
			},
		}

		mockDcw := newMockDoguConfigWriter(t)
		mockDcw.EXPECT().applyDefaultDoguConfig(testCtx, doguDefaults, expectedSensitiveConfig).Return(assert.AnError)

		dca := &DefaultConfigApplier{
			passwordGenerator:  mockPg,
			globalConfigWriter: mockGcw,
			doguConfigWriter:   mockDcw,
		}

		err := dca.ApplyDefaultConfig(testCtx)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply default dogu config:")
	})
}

func TestNewDefaultConfigApplier(t *testing.T) {
	mockGlobalRepo := newMockGlobalConfigRepo(t)
	mockDoguRepo := newMockDoguConfigRepo(t)
	mockSensitiveDoguRepo := newMockDoguConfigRepo(t)
	mockSecClient := newMockSecretClient(t)

	applier := NewDefaultConfigApplier(mockGlobalRepo, mockDoguRepo, mockSensitiveDoguRepo, mockSecClient)

	require.NotNil(t, applier)
	assert.NotNil(t, applier.passwordGenerator)
	assert.IsType(t, &adminPasswordGenerator{}, applier.passwordGenerator)
	assert.NotNil(t, applier.globalConfigWriter)
	assert.IsType(t, &cesGlobalConfigWriter{}, applier.globalConfigWriter)
	assert.Equal(t, mockGlobalRepo, applier.globalConfigWriter.(*cesGlobalConfigWriter).globalConfigRepo)
	assert.NotNil(t, applier.doguConfigWriter)
	assert.IsType(t, &cesDoguConfigWriter{}, applier.doguConfigWriter)
	assert.Equal(t, mockDoguRepo, applier.doguConfigWriter.(*cesDoguConfigWriter).doguConfigRepo)
	assert.Equal(t, mockSensitiveDoguRepo, applier.doguConfigWriter.(*cesDoguConfigWriter).sensitiveDoguConfigRepo)
}
