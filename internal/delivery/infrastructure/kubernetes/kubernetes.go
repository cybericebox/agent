package k8s

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/lib/pkg/worker"
	calico "github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type (
	Kubernetes struct {
		kubeClient   *kubernetes.Clientset
		calicoClient *calico.Clientset
		worker       worker.Worker
		podCIDR      string
	}

	Dependencies struct {
		Config *config.KubernetesConfig
		Worker worker.Worker
	}
)

func NewKubernetes(deps Dependencies) *Kubernetes {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if deps.Config.KubeConfigPath != "" {
		kubeconfig = deps.Config.KubeConfigPath
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get kubernetes config")
		}
	}

	k := &Kubernetes{
		podCIDR: deps.Config.PodsCIDR,
		worker:  deps.Worker,
	}
	k.kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create kubernetes client")
	}
	k.calicoClient, err = calico.NewForConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create calico client")
	}

	return k
}

func (k *Kubernetes) GetKubeClient() *kubernetes.Clientset {
	return k.kubeClient
}

func (k *Kubernetes) GetCalicoClient() *calico.Clientset {
	return k.calicoClient
}
