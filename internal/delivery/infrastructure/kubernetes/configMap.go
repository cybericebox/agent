package k8s

import (
	"context"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/pkg/appError"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func (k *Kubernetes) ApplyConfigMap(ctx context.Context, name, labID string, data map[string]string) error {
	if _, err := k.kubeClient.CoreV1().ConfigMaps(labID).Apply(ctx,
		v1.ConfigMap(name, labID).WithLabels(map[string]string{
			config.PlatformLabel: config.LabDNSConfig,
			config.LabIDLabel:    labID,
		}).WithData(data), metaV1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to apply config map").Err()
	}

	return nil
}

func (k *Kubernetes) GetConfigMapData(ctx context.Context, name, labID string) (map[string]string, error) {
	get, err := k.kubeClient.CoreV1().ConfigMaps(labID).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, appError.ErrKubernetes.WithError(err).WithMessage("Failed to get config map data").Err()
	}

	return get.Data, nil
}

func (k *Kubernetes) DeleteConfigMap(ctx context.Context, name, labID string) error {
	if err := k.kubeClient.CoreV1().ConfigMaps(labID).Delete(ctx, name, metaV1.DeleteOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to delete config map").Err()
	}

	return nil
}
