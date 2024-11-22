package k8s

import (
	"context"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/pkg/appError"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) ApplyNetwork(ctx context.Context, name, cidr string, blockSize int) error {
	if k.networkExists(ctx, name) {
		return nil
	}
	return k.createNetwork(ctx, name, cidr, blockSize)
}

func (k *Kubernetes) DeleteNetwork(ctx context.Context, name string) error {
	if err := k.calicoClient.ProjectcalicoV3().IPPools().Delete(ctx, name, metaV1.DeleteOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to delete network").Err()
	}

	return nil
}

func (k *Kubernetes) networkExists(ctx context.Context, name string) bool {
	if _, err := k.calicoClient.ProjectcalicoV3().IPPools().Get(ctx, name, metaV1.GetOptions{}); err != nil {
		return false
	}
	return true
}

func (k *Kubernetes) GetNetworkCIDR(ctx context.Context, name string) (string, error) {
	get, err := k.calicoClient.ProjectcalicoV3().IPPools().Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return "", appError.ErrKubernetes.WithError(err).WithMessage("Failed to get network CIDR").Err()
	}
	return get.Spec.CIDR, nil
}

func (k *Kubernetes) createNetwork(ctx context.Context, name, cidr string, blockSize int) error {
	if _, err := k.calicoClient.ProjectcalicoV3().IPPools().Create(ctx,
		&v3.IPPool{
			TypeMeta: metaV1.TypeMeta{},
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					config.PlatformLabel: config.LabNetwork,
					config.LabIDLabel:    name,
				},
			},
			Spec: v3.IPPoolSpec{
				CIDR:         cidr,
				IPIPMode:     "Always",
				NATOutgoing:  true,
				BlockSize:    blockSize,
				NodeSelector: "!all()",
			},
		}, metaV1.CreateOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to create network").Err()
	}

	return nil
}
