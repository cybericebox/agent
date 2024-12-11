package model

type (
	ChallengeConfig struct {
		ID        string
		Instances []InstanceConfig
	}

	InstanceConfig struct {
		ID        string
		Image     string
		Resources ResourcesConfig
		Envs      []EnvConfig
		Records   []DNSRecordConfig
	}

	ResourcesConfig struct {
		Requests ResourceConfig
		Limit    ResourceConfig
	}

	ResourceConfig struct {
		Memory int64
		CPU    int64
	}

	EnvConfig struct {
		Name  string
		Value string
	}

	DNSRecordConfig struct {
		Type string
		Name string
		Data string
	}
)
