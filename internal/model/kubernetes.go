package model

type ApplyDeploymentConfig struct {
	Name           string
	LabID          string
	ReplicaCount   int32
	Image          string
	IP             string
	Labels         map[string]string
	Resources      ResourcesConfig
	Envs           []EnvConfig
	Args           []string
	Volumes        []Volume
	Privileged     bool
	CapAdds        []string
	ReadinessProbe *Probe
}

type Volume struct {
	Name          string
	ConfigMapName string
	HostPath      string
	Mounts        []Mount
}

type Mount struct {
	MountPath string
	SubPath   string
}

type Probe struct {
	Cmd           []string
	PeriodSeconds int32
}

type DeploymentStatus struct {
	Name          string
	Labels        map[string]string
	IP            string
	AllReplicas   int32
	ReadyReplicas int32
}
