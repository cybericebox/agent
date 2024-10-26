package k8s

import (
	"context"
	apinetworkingv1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
	networkingv1 "k8s.io/client-go/applyconfigurations/networking/v1"
)

func (k *Kubernetes) ApplyNetworkPolicy(ctx context.Context, labID string) error {
	if _, err := k.kubeClient.NetworkingV1().NetworkPolicies(labID).Apply(ctx, &networkingv1.NetworkPolicyApplyConfiguration{
		TypeMetaApplyConfiguration:   *v1.TypeMeta().WithKind("NetworkPolicy").WithAPIVersion("networking.k8s.io/v1"),
		ObjectMetaApplyConfiguration: v1.ObjectMeta().WithName("default").WithNamespace(labID),
		Spec: networkingv1.NetworkPolicySpec().
			WithPolicyTypes(apinetworkingv1.PolicyTypeIngress, apinetworkingv1.PolicyTypeEgress).
			WithPodSelector(v1.LabelSelector()).
			WithIngress(
				networkingv1.NetworkPolicyIngressRule().
					WithFrom(networkingv1.NetworkPolicyPeer().WithPodSelector(v1.LabelSelector())),
				networkingv1.NetworkPolicyIngressRule().
					WithFrom(networkingv1.NetworkPolicyPeer().WithNamespaceSelector(v1.LabelSelector())),
			).
			WithEgress(
				networkingv1.NetworkPolicyEgressRule().
					WithTo(networkingv1.NetworkPolicyPeer().WithPodSelector(v1.LabelSelector())),
				networkingv1.NetworkPolicyEgressRule().
					WithTo(networkingv1.NetworkPolicyPeer().WithIPBlock(networkingv1.IPBlock().WithCIDR("0.0.0.0/0").WithExcept(k.podCIDR))),
			),
	}, metaV1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
		return err
	}
	return nil
}

func (k *Kubernetes) DeleteNetworkPolicy(ctx context.Context, labID string) error {
	return k.kubeClient.NetworkingV1().NetworkPolicies(labID).Delete(ctx, "default", metaV1.DeleteOptions{})
}
