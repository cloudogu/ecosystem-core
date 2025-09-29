package fqdn

import (
	"context"
	"testing"
	"time"

	regLibConfig "github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplier_getFQDNFromLoadBalancerService(t *testing.T) {
	t.Run("should get fqdn as IP from load balancer service", func(t *testing.T) {
		sg := newMockServiceGetter(t)
		a := &Applier{serviceGetter: sg}

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: cesLoadBalancerServiceName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "203.0.113.10"},
					},
				},
			},
		}

		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(svc, nil)

		got, err := a.getFQDNFromLoadBalancerService(context.Background(), 5*time.Second)

		require.NoError(t, err)
		assert.Equal(t, "203.0.113.10", got)
	})

	t.Run("should get fqdn as hostname from load balancer service", func(t *testing.T) {
		sg := newMockServiceGetter(t)
		a := &Applier{serviceGetter: sg}

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: cesLoadBalancerServiceName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{Hostname: "my.host"},
					},
				},
			},
		}

		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(svc, nil)

		got, err := a.getFQDNFromLoadBalancerService(context.Background(), 5*time.Second)

		require.NoError(t, err)
		assert.Equal(t, "my.host", got)
	})

	t.Run("should fail to get fqdn if service type is not loadbalancer", func(t *testing.T) {
		sg := newMockServiceGetter(t)
		a := &Applier{serviceGetter: sg}

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: cesLoadBalancerServiceName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "203.0.113.10"},
					},
				},
			},
		}

		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(svc, nil)

		_, err := a.getFQDNFromLoadBalancerService(context.Background(), 5*time.Second)

		require.Error(t, err)
		assert.ErrorContains(t, err, "service \"ces-loadbalancer\" is not of type LoadBalancer")
	})

	t.Run("should get fqdn from load balancer service with multiple tries", func(t *testing.T) {
		sg := newMockServiceGetter(t)
		a := &Applier{serviceGetter: sg}

		originalWaitBetweenTries := waitBetweenTries
		defer func() { waitBetweenTries = originalWaitBetweenTries }()
		waitBetweenTries = 1 * time.Millisecond

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: cesLoadBalancerServiceName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "203.0.113.10"},
					},
				},
			},
		}

		tries := 0
		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).RunAndReturn(func(ctx context.Context, name string, options metav1.GetOptions) (*corev1.Service, error) {
			tries++
			if tries == 2 {
				return svc, nil
			} else if tries >= 3 {
				t.Fatal("should not be called more than 3 times")
			}

			return nil, assert.AnError
		})

		got, err := a.getFQDNFromLoadBalancerService(context.Background(), 5*time.Second)

		require.NoError(t, err)
		assert.Equal(t, "203.0.113.10", got)
	})

	t.Run("should timeout while getting fqdn from load balancer service with multiple tries", func(t *testing.T) {
		sg := newMockServiceGetter(t)
		a := &Applier{serviceGetter: sg}

		originalWaitBetweenTries := waitBetweenTries
		defer func() {
			waitBetweenTries = originalWaitBetweenTries
		}()
		waitBetweenTries = 4 * time.Millisecond

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: cesLoadBalancerServiceName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
		}

		tries := 0
		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).RunAndReturn(func(ctx context.Context, name string, options metav1.GetOptions) (*corev1.Service, error) {
			tries++
			if tries > 3 {
				t.Fatal("should not be called more than 3 times")
			}

			return svc, nil
		})

		_, err := a.getFQDNFromLoadBalancerService(context.Background(), 10*time.Millisecond)

		require.Error(t, err)
		assert.Equal(t, 3, tries, "should be called 2 times")
		assert.ErrorContains(t, err, "timed out after 10ms waiting for external address on service \"ces-loadbalancer\"")
	})
}

func TestApplier_ApplyInitialFQDN(t *testing.T) {
	testCtx := context.Background()

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: cesLoadBalancerServiceName,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "203.0.113.10"},
				},
			},
		},
	}

	t.Run("should apply initial fqdn", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		cr := newMockGlobalConfigRepo(t)
		cr.EXPECT().Get(testCtx).Return(emptyConfig, nil)
		cr.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 1)

			val, exists := cfg.Get("fqdn")
			assert.True(t, exists)
			assert.Equal(t, "203.0.113.10", val.String())

			return cfg, nil
		})

		sg := newMockServiceGetter(t)
		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(svc, nil)

		a := &Applier{
			globalConfigRepo: cr,
			serviceGetter:    sg,
		}

		err := a.ApplyInitialFQDN(testCtx, time.Second)

		require.NoError(t, err)
	})

	t.Run("should fail to apply initial fqdn on error getting global-config", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		cr := newMockGlobalConfigRepo(t)
		cr.EXPECT().Get(testCtx).Return(emptyConfig, assert.AnError)

		sg := newMockServiceGetter(t)

		a := &Applier{
			globalConfigRepo: cr,
			serviceGetter:    sg,
		}

		err := a.ApplyInitialFQDN(testCtx, time.Second)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "error reading global config while checking for fqdn:")
	})

	t.Run("should fail to apply initial fqdn on error saving global-config", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		cr := newMockGlobalConfigRepo(t)
		cr.EXPECT().Get(testCtx).Return(emptyConfig, nil)
		cr.EXPECT().SaveOrMerge(testCtx, mock.Anything).RunAndReturn(func(ctx context.Context, cfg regLibConfig.GlobalConfig) (regLibConfig.GlobalConfig, error) {
			assert.Len(t, cfg.GetAll(), 1)

			val, exists := cfg.Get("fqdn")
			assert.True(t, exists)
			assert.Equal(t, "203.0.113.10", val.String())

			return cfg, assert.AnError
		})

		sg := newMockServiceGetter(t)
		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(svc, nil)

		a := &Applier{
			globalConfigRepo: cr,
			serviceGetter:    sg,
		}

		err := a.ApplyInitialFQDN(testCtx, time.Second)

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to save global config while setting fqdn")
	})

	t.Run("should fail to apply initial fqdn on timeout getting fqdn", func(t *testing.T) {
		emptyConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))

		cr := newMockGlobalConfigRepo(t)
		cr.EXPECT().Get(testCtx).Return(emptyConfig, nil)

		sg := newMockServiceGetter(t)
		sg.EXPECT().Get(mock.Anything, cesLoadBalancerServiceName, metav1.GetOptions{}).Return(nil, assert.AnError)

		a := &Applier{
			globalConfigRepo: cr,
			serviceGetter:    sg,
		}

		err := a.ApplyInitialFQDN(testCtx, time.Millisecond*10)

		require.Error(t, err)
		assert.ErrorContains(t, err, "error getting fqdn from load balancer service: timed out after 10ms waiting for external address on service \"ces-loadbalancer\"")
	})

	t.Run("should not apply initial fqdn if already set", func(t *testing.T) {
		existingConfig := regLibConfig.CreateGlobalConfig(make(regLibConfig.Entries))
		newExisting, err := existingConfig.Set("fqdn", "1.2.3.4")
		require.NoError(t, err)
		existingConfig = regLibConfig.GlobalConfig{Config: newExisting}

		cr := newMockGlobalConfigRepo(t)
		cr.EXPECT().Get(testCtx).Return(existingConfig, nil)

		sg := newMockServiceGetter(t)

		a := &Applier{
			globalConfigRepo: cr,
			serviceGetter:    sg,
		}

		err = a.ApplyInitialFQDN(testCtx, time.Second)

		require.NoError(t, err)
	})
}

func TestNewApplier(t *testing.T) {
	t.Run("should create new applier", func(t *testing.T) {
		cr := newMockGlobalConfigRepo(t)
		sg := newMockServiceGetter(t)

		applier := NewApplier(cr, sg)

		require.NotNil(t, applier)
		assert.Equal(t, cr, applier.globalConfigRepo)
		assert.Equal(t, sg, applier.serviceGetter)
	})
}
