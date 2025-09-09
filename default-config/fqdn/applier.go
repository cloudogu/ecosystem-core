package fqdn

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const fqdnKey = "fqdn"
const cesLoadBalancerServiceName = "ces-loadbalancer"
const defaultWaitBetweenTries = 2 * time.Second

var waitBetweenTries = defaultWaitBetweenTries

type globalConfigRepo interface {
	Get(ctx context.Context) (regLibConfig.GlobalConfig, error)
	SaveOrMerge(ctx context.Context, globalConfig regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error)
}

type serviceGetter interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Service, error)
}

type Applier struct {
	globalConfigRepo globalConfigRepo
	serviceGetter    serviceGetter
}

func NewApplier(globalConfigRepo globalConfigRepo, serviceGetter serviceGetter) *Applier {
	return &Applier{
		globalConfigRepo: globalConfigRepo,
		serviceGetter:    serviceGetter,
	}
}

func (a *Applier) ApplyInitialFQDN(ctx context.Context, timeout time.Duration) error {
	globalConfig, err := a.globalConfigRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("error reading global config while checking for fqdn: %w", err)
	}

	fqdn, exists := globalConfig.Get(fqdnKey)
	if exists && fqdn != "" {
		slog.Info("fqdn already set. Skipping...")
		return nil
	}

	slog.Info("fqdn not set. Retrieving fqdn from load balancer service...")

	loadBalancerFqdn, err := a.getFQDNFromLoadBalancerService(ctx, timeout)
	if err != nil {
		return fmt.Errorf("error getting fqdn from load balancer service: %w", err)
	}

	slog.Info("fqdn retrieved from load balancer service", "fqdn", loadBalancerFqdn)

	newGlobalConfig, err := globalConfig.Set(fqdnKey, regLibConfig.Value(loadBalancerFqdn))
	if err != nil {
		return fmt.Errorf("failed to set fqdn in global config: %w", err)
	}

	globalConfig = regLibConfig.GlobalConfig{Config: newGlobalConfig}

	_, err = a.globalConfigRepo.SaveOrMerge(ctx, globalConfig)
	if err != nil {
		return fmt.Errorf("failed to save global config while setting fqdn: %w", err)
	}

	slog.Info("...Successfully applied fqdn from load balancer service to global config.")

	return nil
}

func (a *Applier) getFQDNFromLoadBalancerService(ctx context.Context, timeout time.Duration) (string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(waitBetweenTries)
	defer ticker.Stop()

	for {
		// Refresh the service to get the latest status
		loadBalancerService, err := a.serviceGetter.Get(ctxWithTimeout, cesLoadBalancerServiceName, metav1.GetOptions{})
		if err != nil {
			slog.Debug("error getting load balancer service", "err", err)
		} else {
			if loadBalancerService.Spec.Type != corev1.ServiceTypeLoadBalancer {
				return "", fmt.Errorf("service %q is not of type LoadBalancer", loadBalancerService.Name)
			}

			ingresses := loadBalancerService.Status.LoadBalancer.Ingress
			if len(ingresses) > 0 {
				ing := ingresses[0]
				if ip := ing.IP; ip != "" {
					return ip, nil
				}
				if host := ing.Hostname; host != "" {
					return host, nil
				}
			}
		}

		select {
		case <-ctxWithTimeout.Done():
			return "", fmt.Errorf("timed out after %s waiting for external address on service %q", timeout, cesLoadBalancerServiceName)
		case <-ticker.C:
			// retry
		}
	}
}
