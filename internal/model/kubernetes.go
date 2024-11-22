package model

type (
	ApplyDeploymentConfig struct {
		Name           string
		LabID          string
		Labels         map[string]string
		ReplicaCount   int32
		Image          string
		IP             string
		DNS            string
		Resources      ResourcesConfig
		Envs           []EnvConfig
		Args           []string
		Volumes        []Volume
		Privileged     bool
		UsePublicDNS   bool
		CapAdds        []string
		ReadinessProbe *Probe
	}

	Volume struct {
		Name          string
		ConfigMapName string
		MountPath     string
	}

	Probe struct {
		Cmd           []string
		PeriodSeconds int32
	}

	DeploymentStatus struct {
		Name     string
		Labels   map[string]string
		IP       string
		Replicas Replicas
	}

	Replicas struct {
		Total       int32
		Ready       int32
		Available   int32
		Unavailable int32
		Updated     int32
	}
)
