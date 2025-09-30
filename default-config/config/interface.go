package config

import corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

type secretClient interface {
	corev1client.SecretInterface
}
