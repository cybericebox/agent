package k8s

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/helper"
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
		if v.HostPath != "" {
			volumes = append(volumes, v13.Volume().WithName(v.Name).
				WithHostPath(v13.HostPathVolumeSource().WithType(coreV1.HostPathDirectory).WithPath(v.HostPath)))
			continue
		}
		if v.ConfigMapName != "" {
			volumes = append(volumes, v13.Volume().WithName(v.Name).
				WithConfigMap(v13.ConfigMapVolumeSource().WithName(v.ConfigMapName)))
		}

		for _, vm := range v.Mounts {
			nVM := v13.VolumeMount().
				WithName(v.Name).
				WithMountPath(vm.MountPath)
			if vm.SubPath != "" {
				nVM = nVM.WithSubPath(vm.SubPath)
			}
			volumeMounts = append(volumeMounts, nVM)
		}
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

	if (cfg.Resources.Limit.CPU != "" && cfg.Resources.Limit.Memory != "") || (cfg.Resources.Requests.CPU != "" && cfg.Resources.Requests.Memory != "") {
		r := v13.ResourceRequirements()
		if cfg.Resources.Limit.CPU != "" {
			cpu, err := resource.ParseQuantity(cfg.Resources.Limit.CPU)
			if err != nil {
				return err
			}
			memory, err := resource.ParseQuantity(cfg.Resources.Limit.Memory)
			if err != nil {
				return err
			}
			r.WithLimits(coreV1.ResourceList{
				coreV1.ResourceCPU:    cpu,
				coreV1.ResourceMemory: memory,
			})
		}
		if cfg.Resources.Requests.CPU != "" {
			cpu, err := resource.ParseQuantity(cfg.Resources.Requests.CPU)
			if err != nil {
				return err
			}
			memory, err := resource.ParseQuantity(cfg.Resources.Requests.Memory)
			if err != nil {
				return err
			}
			r.WithRequests(coreV1.ResourceList{
				coreV1.ResourceCPU:    cpu,
				coreV1.ResourceMemory: memory,
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

	if cfg.ReplicaCount == 0 {
		cfg.ReplicaCount = 1
	}

	if cfg.ReadinessProbe != nil {
		container = container.WithReadinessProbe(v13.Probe().
			WithPeriodSeconds(cfg.ReadinessProbe.PeriodSeconds).
			WithExec(v13.ExecAction().WithCommand(cfg.ReadinessProbe.Cmd...)))
	}

	_, err := k.kubeClient.AppsV1().Deployments(cfg.LabID).Apply(
		ctx,
		v1.Deployment(cfg.Name, cfg.LabID).WithLabels(cfg.Labels).
			WithSpec(v1.DeploymentSpec().
				WithSelector(v12.LabelSelector().WithMatchLabels(map[string]string{labInstanceIDLabel: helper.GetLabel(cfg.LabID, cfg.Name)})).
				WithReplicas(cfg.ReplicaCount).
				WithTemplate(v13.PodTemplateSpec().
					WithName(cfg.Name).
					WithNamespace(cfg.LabID).
					WithLabels(map[string]string{
						labInstanceIDLabel: helper.GetLabel(cfg.LabID, cfg.Name),
					}).
					WithLabels(cfg.Labels).
					WithAnnotations(annotations).
					WithSpec(v13.PodSpec().
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
		metaV1.ApplyOptions{FieldManager: "application/apply-patch"})
	return err
}

func (k *Kubernetes) GetDeploymentsInNamespaceBySelector(ctx context.Context, labId string, selector ...string) ([]model.DeploymentStatus, error) {
	labelSelector := fmt.Sprintf("%s=%s", config.PlatformLabel, config.Lab)

	if len(selector) > 0 {
		labelSelector = selector[0]
	}

	dps, err := k.kubeClient.AppsV1().Deployments(labId).List(ctx, metaV1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	dpsStatus := make([]model.DeploymentStatus, 0)

	for _, dp := range dps.Items {
		dpsStatus = append(dpsStatus, model.DeploymentStatus{
			Name:          dp.GetName(),
			IP:            dp.Spec.Template.Annotations["ip"],
			AllReplicas:   dp.Status.Replicas,
			ReadyReplicas: dp.Status.ReadyReplicas,
			Labels:        dp.GetLabels(),
		})
	}

	return dpsStatus, nil
}

func (k *Kubernetes) DeploymentExists(ctx context.Context, name, labId string) (bool, error) {
	dp, err := k.kubeClient.AppsV1().Deployments(labId).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return dp.GetName() == name && dp.GetNamespace() == labId, nil
}

func (k *Kubernetes) ResetDeployment(ctx context.Context, name, labId string) error {
	if err := k.ScaleDeployment(ctx, name, labId, 0); err != nil {
		return err

	}
	if err := k.ScaleDeployment(ctx, name, labId, 1); err != nil {
		return err
	}
	return nil
}

func (k *Kubernetes) ScaleDeployment(ctx context.Context, name, labId string, scale int32) error {
	if _, err := k.kubeClient.AppsV1().Deployments(labId).UpdateScale(ctx, name, &autoscalingv1.Scale{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Scale",
			APIVersion: "autoscaling/v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: labId,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: scale,
		},
	}, metaV1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (k *Kubernetes) DeleteDeployment(ctx context.Context, name, labId string) error {
	return k.kubeClient.AppsV1().Deployments(labId).Delete(ctx, name, metaV1.DeleteOptions{})
}
