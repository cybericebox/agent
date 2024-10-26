package k8s

import (
	calico "github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type Kubernetes struct {
	kubeClient   *kubernetes.Clientset
	calicoClient *calico.Clientset
	podCIDR      string
}

func NewKubernetes(podCIDR string) *Kubernetes {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get kubernetes config")
		}
	}

	k := &Kubernetes{
		podCIDR: podCIDR,
	}
	k.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create kubernetes client")
	}
	k.calicoClient, err = calico.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create calico client")
	}
	return k
}
