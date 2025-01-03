package k8s

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/tools"
	"github.com/cybericebox/agent/pkg/appError"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/apps/v1"
	v13 "k8s.io/client-go/applyconfigurations/core/v1"
	v12 "k8s.io/client-go/applyconfigurations/meta/v1"
	"strings"
)

const labInstanceIDLabel = "lab-instance-id"

func (k *Kubernetes) ApplyDeployment(ctx context.Context, cfg model.ApplyDeploymentConfig) error {

	container := v13.Container()

	volumeMounts := make([]*v13.VolumeMountApplyConfiguration, 0)
	volumes := make([]*v13.VolumeApplyConfiguration, 0)
	for _, v := range cfg.Volumes {
		if v.ConfigMapName != "" {
			volumes = append(volumes, v13.Volume().WithName(v.Name).
				WithConfigMap(v13.ConfigMapVolumeSource().WithName(v.ConfigMapName)))
		}
		volumeMounts = append(volumeMounts, v13.VolumeMount().WithName(v.Name).WithMountPath(v.MountPath))
	}
	if len(volumeMounts) == 0 {
		volumes = nil
		volumeMounts = nil
	}

	envVars := make([]*v13.EnvVarApplyConfiguration, 0)
	for _, env := range cfg.Envs {
		envVars = append(envVars, v13.EnvVar().WithName(env.Name).WithValue(env.Value))
	}
	if len(envVars) > 0 {
		container = container.WithEnv(envVars...)
	}

	if len(cfg.Args) > 0 {
		container = container.WithArgs(cfg.Args...)
	}

	if (cfg.Resources.Limit.CPU != 0 && cfg.Resources.Limit.Memory != 0) || (cfg.Resources.Requests.CPU != 0 && cfg.Resources.Requests.Memory != 0) {
		r := v13.ResourceRequirements()
		if cfg.Resources.Limit.CPU != 0 {
			cpu := resource.NewMilliQuantity(cfg.Resources.Limit.CPU, resource.DecimalExponent)
			memory := resource.NewQuantity(cfg.Resources.Limit.Memory, resource.BinarySI)
			r.WithLimits(coreV1.ResourceList{
				coreV1.ResourceCPU:    *cpu,
				coreV1.ResourceMemory: *memory,
			})
		}
		if cfg.Resources.Requests.CPU != 0 {
			cpu := resource.NewMilliQuantity(cfg.Resources.Requests.CPU, resource.DecimalExponent)
			memory := resource.NewQuantity(cfg.Resources.Requests.Memory, resource.BinarySI)
			r.WithRequests(coreV1.ResourceList{
				coreV1.ResourceCPU:    *cpu,
				coreV1.ResourceMemory: *memory,
			})
		}
		container = container.WithResources(r)
	}

	annotations := make(map[string]string)
	if cfg.IP != "" {
		annotations["cni.projectcalico.org/ipAddrs"] = fmt.Sprintf("[\"%s\"]", strings.Split(cfg.IP, "/")[0])
		annotations["ip"] = strings.Split(cfg.IP, "/")[0]
	}

	capAdds := make([]coreV1.Capability, 0)
	for _, cd := range cfg.CapAdds {
		capAdds = append(capAdds, coreV1.Capability(cd))
	}
	if len(capAdds) == 0 {
		capAdds = nil
	}

	if cfg.ReadinessProbe != nil {
		container = container.WithReadinessProbe(v13.Probe().
			WithPeriodSeconds(cfg.ReadinessProbe.PeriodSeconds).
			WithExec(v13.ExecAction().WithCommand(cfg.ReadinessProbe.Cmd...)))
	}

	dnsConfig := v13.PodDNSConfig()
	dnsPolicy := coreV1.DNSDefault
	//dnsPolicy := coreV1.DNSDefault
	if cfg.DNS != "" {
		dnsConfig = dnsConfig.WithNameservers(strings.Split(cfg.DNS, "/")[0])
		dnsPolicy = coreV1.DNSNone
	}
	if cfg.UsePublicDNS {
		dnsConfig = dnsConfig.WithNameservers("1.1.1.1", "8.8.8.8")
	}

	if _, err := k.kubeClient.AppsV1().Deployments(cfg.LabID).Apply(
		ctx,
		v1.Deployment(cfg.Name, cfg.LabID).WithLabels(cfg.Labels).
			WithSpec(v1.DeploymentSpec().
				WithSelector(v12.LabelSelector().WithMatchLabels(map[string]string{labInstanceIDLabel: tools.GetLabel(cfg.LabID, cfg.Name)})).
				WithReplicas(cfg.ReplicaCount).
				WithTemplate(v13.PodTemplateSpec().
					WithName(cfg.Name).
					WithNamespace(cfg.LabID).
					WithLabels(map[string]string{
						labInstanceIDLabel: tools.GetLabel(cfg.LabID, cfg.Name),
					}).
					WithLabels(cfg.Labels).
					WithAnnotations(annotations).
					WithSpec(v13.PodSpec().
						WithDNSPolicy(dnsPolicy).
						WithRestartPolicy(coreV1.RestartPolicyAlways).
						WithDNSConfig(dnsConfig).
						WithVolumes(volumes...).
						WithContainers(container.
							WithName(cfg.Name).
							WithImage(cfg.Image).
							WithSecurityContext(v13.SecurityContext().
								WithPrivileged(cfg.Privileged).
								WithAllowPrivilegeEscalation(cfg.Privileged).
								WithCapabilities(v13.Capabilities().
									WithAdd(capAdds...))).
							WithVolumeMounts(volumeMounts...))))),
		metaV1.ApplyOptions{FieldManager: "application/apply-patch"}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to apply deployment").Err()
	}

	return nil
}

func (k *Kubernetes) GetDeploymentsInNamespaceBySelector(ctx context.Context, labID string, selector ...string) ([]model.DeploymentStatus, error) {
	labelSelector := strings.Join(selector, ",")

	dps, err := k.kubeClient.AppsV1().Deployments(labID).List(ctx, metaV1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, appError.ErrKubernetes.WithError(err).WithMessage("Failed to get deployments").Err()
	}

	dpsStatus := make([]model.DeploymentStatus, 0)

	for _, dp := range dps.Items {
		dpsStatus = append(dpsStatus, model.DeploymentStatus{
			Name:   dp.GetName(),
			IP:     dp.Spec.Template.Annotations["ip"],
			Status: StatusFromReplicas(dp.Status.Replicas, dp.Status.ReadyReplicas, dp.Status.AvailableReplicas, dp.Status.UnavailableReplicas),
			Labels: dp.GetLabels(),
		})
	}

	return dpsStatus, nil
}

func StatusFromReplicas(total, ready, available, unavailable int32) model.Status {
	if total > ready || total > available {
		return model.StatusStarting
	}
	if total == ready && available == ready && unavailable == 0 {
		return model.StatusRunning
	}
	if total < ready || total < available {
		return model.StatusStopping
	}
	if total == 0 && ready == 0 && available == 0 && unavailable == 0 {
		return model.StatusStopped
	}
	if total != ready && unavailable > 0 {
		return model.StatusError
	}
	return model.StatusUnknown
}

func (k *Kubernetes) DeploymentExists(ctx context.Context, name, labID string) (bool, error) {
	dp, err := k.kubeClient.AppsV1().Deployments(labID).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, appError.ErrKubernetes.WithError(err).WithMessage("Failed to get deployment").Err()
		}
	}
	return dp.GetName() == name && dp.GetNamespace() == labID, nil
}

func (k *Kubernetes) ResetDeployment(ctx context.Context, name, labID string) error {
	if err := k.ScaleDeployment(ctx, name, labID, 0); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to scale deployment down").Err()

	}
	if err := k.ScaleDeployment(ctx, name, labID, 1); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to scale deployment up").Err()
	}
	return nil
}

func (k *Kubernetes) DeleteDeployment(ctx context.Context, name, labID string) error {
	if err := k.kubeClient.AppsV1().Deployments(labID).Delete(ctx, name, metaV1.DeleteOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to delete deployment").Err()
	}

	return nil
}

func (k *Kubernetes) ScaleDeployment(ctx context.Context, name, labID string, scale int32) error {
	if _, err := k.kubeClient.AppsV1().Deployments(labID).UpdateScale(ctx, name, &autoscalingv1.Scale{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Scale",
			APIVersion: "autoscaling/v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: labID,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: scale,
		},
	}, metaV1.UpdateOptions{}); err != nil {
		return appError.ErrKubernetes.WithError(err).WithMessage("Failed to scale deployment").Err()
	}

	return nil
}
