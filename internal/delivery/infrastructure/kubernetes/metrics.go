package k8s

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func (k *Kubernetes) GetPodsMetrics(ctx context.Context, namespace string, selectors ...string) ([]model.PodMetrics, error) {
	labelSelector := strings.Join(selectors, ",")
	podList, err := k.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metaV1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, appError.ErrKubernetes.WithError(err).WithMessage("Failed to get pods metrics").Err()
	}

	podMetrics := make([]model.PodMetrics, 0, len(podList.Items))
	for _, pod := range podList.Items {
		podMetrics = append(podMetrics, model.PodMetrics{
			Labels: pod.GetLabels(),
			Resources: model.ResourceConfig{
				Memory: pod.Containers[0].Usage.Memory().Value(),
				CPU:    pod.Containers[0].Usage.Cpu().MilliValue(),
			},
		})
	}

	return podMetrics, nil
}
