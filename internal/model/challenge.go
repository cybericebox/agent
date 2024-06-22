package model

type (
	ChallengeConfig struct {
		Id        string
		Instances []InstanceConfig
	}

	InstanceConfig struct {
		Id        string
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
		Memory string
		CPU    string
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
