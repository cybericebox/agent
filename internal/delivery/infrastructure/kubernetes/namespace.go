package k8s

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/pkg/appError"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func (k *Kubernetes) ApplyNamespace(ctx context.Context, name string, ipPoolName *string) error {
	annotations := make(map[string]string)
	if ipPoolName != nil {
		annotations["cni.projectcalico.org/ipv4pools"] = fmt.Sprintf("[\"%s\"]", *ipPoolName)
	}

	if _, err := k.kubeClient.CoreV1().Namespaces().Apply(
		ctx,
		v1.Namespace(name).WithAnnotations(annotations).WithLabels(map[string]string{config.PlatformLabel: config.Lab, config.LabIDLabel: name}),
		metaV1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to apply namespace").Err()
	}

	return nil
}

func (k *Kubernetes) NamespaceExists(ctx context.Context, name string) (bool, error) {
	ns, err := k.kubeClient.CoreV1().Namespaces().Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, appError.ErrKubernetes.WithError(err).WithMessage("Failed to get namespace").Err()
		}
	}

	return ns.GetName() == name, nil
}

func (k *Kubernetes) DeleteNamespace(ctx context.Context, name string) error {
	if err := k.kubeClient.CoreV1().Namespaces().Delete(ctx, name, metaV1.DeleteOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to delete namespace").Err()
	}

	return nil
}
