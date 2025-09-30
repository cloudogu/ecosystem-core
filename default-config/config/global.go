package config

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"slices"

	cesLibErr "github.com/cloudogu/ces-commons-lib/errors"
	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ecosystemCertificateName    = "ecosystem-certificate"
	ecosystemCertificateDataKey = "tls.crt"
	localIssuer                 = "ces.local"

	certificateConfigTypeKey   = "certificate/type"
	certificateSelfSignedValue = "selfsigned"
	certificateExternalValue   = "external"
)

type globalConfigRepo interface {
	Get(ctx context.Context) (regLibConfig.GlobalConfig, error)
	Create(ctx context.Context, globalConfig regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error)
	SaveOrMerge(ctx context.Context, globalConfig regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error)
}

type cesGlobalConfigWriter struct {
	globalConfigRepo globalConfigRepo
	secretClient     secretClient
	parseCertificate func(der []byte) (*x509.Certificate, error)
	pemDecode        func(data []byte) (p *pem.Block, rest []byte)
}

func newCesGlobalConfigWriter(globalConfigRepo globalConfigRepo, secretClient secretClient) *cesGlobalConfigWriter {
	return &cesGlobalConfigWriter{
		globalConfigRepo: globalConfigRepo,
		secretClient:     secretClient,
		parseCertificate: x509.ParseCertificate,
		pemDecode:        pem.Decode,
	}
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

		if cKey.String() == certificateConfigTypeKey {
			certType, sErr := gcw.getCertificateType(ctx)
			if sErr != nil {
				return fmt.Errorf("failed to get default value for %s: %w", certificateConfigTypeKey, sErr)
			}

			cValue = regLibConfig.Value(certType)
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

func (gcw *cesGlobalConfigWriter) getCertificateType(ctx context.Context) (string, error) {
	external, cErr := gcw.isExternalCertificate(ctx)
	if cErr != nil {
		return "", fmt.Errorf("failed to verify external certificate: %w", cErr)
	}

	if external {
		return certificateExternalValue, nil
	}

	return certificateSelfSignedValue, nil
}

func (gcw *cesGlobalConfigWriter) isExternalCertificate(ctx context.Context) (bool, error) {
	certSecret, err := gcw.secretClient.Get(ctx, ecosystemCertificateName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return false, fmt.Errorf("failed to get secret for ecosystem certificate: %w", err)
	}

	if apierrors.IsNotFound(err) {
		return false, nil
	}

	certBytes, ok := certSecret.Data[ecosystemCertificateDataKey]
	if !ok {
		return false, nil
	}

	leaf, _ := gcw.pemDecode(certBytes)
	if leaf == nil || leaf.Type != "CERTIFICATE" {
		return false, nil
	}

	cert, err := gcw.parseCertificate(leaf.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse ecosystem certificate: %w", err)
	}

	if slices.Contains(cert.Issuer.Organization, localIssuer) {
		return false, nil
	}

	return true, nil
}
