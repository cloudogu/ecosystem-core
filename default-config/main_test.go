package main

import (
	"context"
	"log/slog"
	"os"
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

func Test_readConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer func() {
			os.Setenv("NAMESPACE", "")
			os.Setenv("LOG_LEVEL", "")
			os.Setenv("WAIT_TIMEOUT_MINUTES", "")
		}()
		os.Setenv("NAMESPACE", "ecosystem")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("WAIT_TIMEOUT_MINUTES", "15")
		job := readConfig()
		assert.Equal(t, job.namespace, "ecosystem")
		assert.Equal(t, job.logLevel, "debug")
		assert.Equal(t, time.Duration(15)*time.Minute, job.waitTimeout)
	})
	t.Run("success with default", func(t *testing.T) {
		defer func() {
			os.Setenv("NAMESPACE", "")
			os.Setenv("LOG_LEVEL", "")
			os.Setenv("WAIT_TIMEOUT_MINUTES", "")
		}()
		os.Setenv("NAMESPACE", "ecosystem")
		os.Setenv("LOG_LEVEL", "debug")
		job := readConfig()
		assert.Equal(t, job.namespace, "ecosystem")
		assert.Equal(t, job.logLevel, "debug")
		assert.Equal(t, time.Duration(5)*time.Minute, job.waitTimeout)
	})
}

func Test_configureLogger(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		configureLogger("debug")
		h := slog.Default().Handler()
		ctx := context.Background()

		// Bei Schwellwert "debug" muss alles durchgehen.
		for _, lvl := range []slog.Level{
			slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError,
		} {
			if !h.Enabled(ctx, lvl) {
				t.Fatalf("expected %s to be enabled at debug level", lvl.String())
			}
		}
	})
	t.Run("fallback to default", func(t *testing.T) {
		configureLogger("invalid")
		h := slog.Default().Handler()
		ctx := context.Background()

		// Bei Schwellwert "debug" muss alles durchgehen.
		for _, lvl := range []slog.Level{
			slog.LevelInfo, slog.LevelWarn, slog.LevelError,
		} {
			if !h.Enabled(ctx, lvl) {
				t.Fatalf("expected %s to be enabled at debug level", lvl.String())
			}
		}
	})
}
