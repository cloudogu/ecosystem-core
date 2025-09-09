package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_applyDefaults(t *testing.T) {
	testCtx := context.Background()

	cfg := jobConfig{waitTimeout: defaultWaitTimeoutMinutes * time.Minute}

	t.Run("should apply defaults", func(t *testing.T) {
		ca := newMockConfigApplier(t)
		ca.EXPECT().ApplyDefaultConfig(testCtx).Return(nil)

		fa := newMockFqdnApplier(t)
		fa.EXPECT().ApplyInitialFQDN(testCtx, cfg.waitTimeout).Return(nil)

		err := applyDefaults(testCtx, cfg, ca, fa)

		require.NoError(t, err)
	})

	t.Run("should fail to apply config defaults", func(t *testing.T) {
		ca := newMockConfigApplier(t)
		ca.EXPECT().ApplyDefaultConfig(testCtx).Return(assert.AnError)

		fa := newMockFqdnApplier(t)

		err := applyDefaults(testCtx, cfg, ca, fa)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply default config:")
	})

	t.Run("should fail to apply fqdn defaults", func(t *testing.T) {
		ca := newMockConfigApplier(t)
		ca.EXPECT().ApplyDefaultConfig(testCtx).Return(nil)

		fa := newMockFqdnApplier(t)
		fa.EXPECT().ApplyInitialFQDN(testCtx, cfg.waitTimeout).Return(assert.AnError)

		err := applyDefaults(testCtx, cfg, ca, fa)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to apply intial fqdn:")
	})
}

func Test_run(t *testing.T) {
	t.Run("should run default-config job", func(t *testing.T) {

	})
}
