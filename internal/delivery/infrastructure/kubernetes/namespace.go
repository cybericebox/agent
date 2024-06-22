package k8s

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"strings"
)

func (k *Kubernetes) ApplyNamespace(ctx context.Context, name string, ipPoolName *string) error {
	annotations := make(map[string]string)
	if ipPoolName != nil {
		annotations["cni.projectcalico.org/ipv4pools"] = fmt.Sprintf("[\"%s\"]", *ipPoolName)
	}

	_, err := k.kubeClient.CoreV1().Namespaces().Apply(
		ctx,
		v1.Namespace(name).WithAnnotations(annotations).WithLabels(map[string]string{config.PlatformLabel: config.Lab, config.LabIDLabel: name}),
		metaV1.ApplyOptions{FieldManager: "application/apply-patch"})
	return err
}

func (k *Kubernetes) NamespaceExists(ctx context.Context, name string) (bool, error) {
	ns, err := k.kubeClient.CoreV1().Namespaces().Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		} else {
			return false, err
		}
	}

	return ns.GetName() == name, nil
}

func (k *Kubernetes) DeleteNamespace(ctx context.Context, name string) error {
	return k.kubeClient.CoreV1().Namespaces().Delete(ctx, name, metaV1.DeleteOptions{})
}
