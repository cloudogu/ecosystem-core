package config

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"

	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	t.Run("should set certificate type to self signed", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 3)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			val, exists = cfg.Get(certificateConfigTypeKey)
			assert.True(t, exists)
			assert.Equal(t, certificateSelfSignedValue, val.String())

			return cfg, nil
		})

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(&corev1.Secret{
			Data: map[string][]byte{
				ecosystemCertificateDataKey: []byte("superSecret"),
			},
		}, nil)

		parseMock := func(der []byte) (*x509.Certificate, error) {
			return &x509.Certificate{
				Issuer: pkix.Name{
					Organization: []string{localIssuer},
				},
			}, nil
		}

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: parseMock,
			pemDecode: func(data []byte) (p *pem.Block, rest []byte) {
				return &pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   []byte{},
				}, nil
			},
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should set certificate type to self signed, when certificate not found", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 3)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			val, exists = cfg.Get(certificateConfigTypeKey)
			assert.True(t, exists)
			assert.Equal(t, certificateSelfSignedValue, val.String())

			return cfg, nil
		})

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(nil, errors.NewNotFound(schema.GroupResource{}, "error"))

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: nil,
			pemDecode:        nil,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should set certificate type to self signed, when data key is not found", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 3)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			val, exists = cfg.Get(certificateConfigTypeKey)
			assert.True(t, exists)
			assert.Equal(t, certificateSelfSignedValue, val.String())

			return cfg, nil
		})

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(&corev1.Secret{
			Data: map[string][]byte{
				"invalid": []byte("superSecret"),
			},
		}, nil)

		parseMock := func(der []byte) (*x509.Certificate, error) {
			return &x509.Certificate{
				Issuer: pkix.Name{
					Organization: []string{localIssuer},
				},
			}, nil
		}

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: parseMock,
			pemDecode: func(data []byte) (p *pem.Block, rest []byte) {
				return &pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   []byte{},
				}, nil
			},
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should set certificate type to self signed, when leaf is no certificate", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 3)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			val, exists = cfg.Get(certificateConfigTypeKey)
			assert.True(t, exists)
			assert.Equal(t, certificateSelfSignedValue, val.String())

			return cfg, nil
		})

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(&corev1.Secret{
			Data: map[string][]byte{
				ecosystemCertificateDataKey: []byte("superSecret"),
			},
		}, nil)

		parseMock := func(der []byte) (*x509.Certificate, error) {
			return &x509.Certificate{
				Issuer: pkix.Name{
					Organization: []string{localIssuer},
				},
			}, nil
		}

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: parseMock,
			pemDecode: func(data []byte) (p *pem.Block, rest []byte) {
				return &pem.Block{
					Type:    "PRIVATE KEY",
					Headers: nil,
					Bytes:   []byte{},
				}, nil
			},
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should set certificate type to external", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		mockRepo.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 3)

			val, exists := cfg.Get("key")
			assert.True(t, exists)
			assert.Equal(t, "value", val.String())

			val, exists = cfg.Get("foo")
			assert.True(t, exists)
			assert.Equal(t, "bar", val.String())

			val, exists = cfg.Get(certificateConfigTypeKey)
			assert.True(t, exists)
			assert.Equal(t, certificateExternalValue, val.String())

			return cfg, nil
		})

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(&corev1.Secret{
			Data: map[string][]byte{
				ecosystemCertificateDataKey: []byte("superSecret"),
			},
		}, nil)

		parseMock := func(der []byte) (*x509.Certificate, error) {
			return &x509.Certificate{
				Issuer: pkix.Name{
					Organization: []string{"otherOrga"},
				},
			}, nil
		}

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: parseMock,
			pemDecode: func(data []byte) (p *pem.Block, rest []byte) {
				return &pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   []byte{},
				}, nil
			},
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.NoError(t, err)
	})

	t.Run("should fail when certificate cannot be received", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(nil, assert.AnError)

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: nil,
			pemDecode:        nil,
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get secret for ecosystem certificate")
	})

	t.Run("should fail when parsing certificate returns error", func(t *testing.T) {
		defaultConfig := map[string]string{
			certificateConfigTypeKey: "",
			"key":                    "value",
			"foo":                    "bar",
		}

		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		mockRepo := newMockGlobalConfigRepo(t)
		mockRepo.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		secretClientMock := newMockSecretClient(t)
		secretClientMock.EXPECT().Get(testCtx, ecosystemCertificateName, mock.Anything).Return(&corev1.Secret{
			Data: map[string][]byte{
				ecosystemCertificateDataKey: []byte("superSecret"),
			},
		}, nil)

		parseMock := func(der []byte) (*x509.Certificate, error) {
			return nil, assert.AnError
		}

		gcw := cesGlobalConfigWriter{
			globalConfigRepo: mockRepo,
			secretClient:     secretClientMock,
			parseCertificate: parseMock,
			pemDecode: func(data []byte) (p *pem.Block, rest []byte) {
				return &pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   []byte{},
				}, nil
			},
		}

		err := gcw.applyDefaultGlobalConfig(testCtx, defaultConfig)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to parse ecosystem certificate")
	})
}
