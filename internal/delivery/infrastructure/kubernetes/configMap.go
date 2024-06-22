package k8s

import (
	"context"
	"github.com/cybericebox/agent/internal/config"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func (k *Kubernetes) ApplyConfigMap(ctx context.Context, name, labId string, data map[string]string) error {
	_, err := k.kubeClient.CoreV1().ConfigMaps(labId).Apply(ctx,
		v1.ConfigMap(name, labId).WithLabels(map[string]string{
			config.PlatformLabel: config.LabDNSConfig,
			config.LabIDLabel:    labId,
		}).WithData(data), metaV1.ApplyOptions{FieldManager: "application/apply-patch"})
	return err
}

func (k *Kubernetes) GetConfigMapData(ctx context.Context, name, labId string) (map[string]string, error) {
	get, err := k.kubeClient.CoreV1().ConfigMaps(labId).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return get.Data, nil
}

func (k *Kubernetes) DeleteConfigMap(ctx context.Context, name, labId string) error {
	return k.kubeClient.CoreV1().ConfigMaps(labId).Delete(ctx, name, metaV1.DeleteOptions{})
}
