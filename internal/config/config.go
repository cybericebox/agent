package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

var MigrationPath string

type (
	Config struct {
		Environment    string               `yaml:"environment" env:"ENV" env-default:"production" env-description:"Environment"`
		Controller     ControllerConfig     `yaml:"controller"`
		Service        ServiceConfig        `yaml:"service"`
		Repository     RepositoryConfig     `yaml:"repository"`
		Infrastructure InfrastructureConfig `yaml:"infrastructure"`
		Worker         WorkerConfig         `yaml:"worker"`
	}

	WorkerConfig struct {
		MaxWorkers int           `yaml:"maxWorkers" env:"AGENT_MAX_WORKERS" env-default:"10" env-description:"Max workers for the worker pool"`
		Throttle   time.Duration `yaml:"throttle" env:"AGENT_WORKERS_THROTTLE" env-default:"10ms" env-description:"Throttle for the worker pool"`
	}

	ControllerConfig struct {
		GRPC GRPCConfig `yaml:"grpc"`
	}

	GRPCConfig struct {
		Host string     `yaml:"host" env:"AGENT_GRPC_HOST" env-default:"0.0.0.0" env-description:"Host of GRPC server"`
		Port string     `yaml:"port" env:"AGENT_GRPC_PORT" env-default:"5454" env-description:"Port of GRPC server"`
		TLS  TLSConfig  `yaml:"tls"`
		Auth AuthConfig `yaml:"auth"`
	}

	TLSConfig struct {
		Enabled  bool   `yaml:"enabled" env:"AGENT_GRPC_TLS_ENABLED" env-default:"false" env-description:"Enabled TLS of GRPC server"`
		CertFile string `yaml:"certFile" env:"AGENT_GRPC_TLS_CERT" env-default:"" env-description:"CertFile of GRPC server"`
		CertKey  string `yaml:"certKey" env:"AGENT_GRPC_TLS_KEY" env-default:"" env-description:"CertKey of GRPC server"`
		CAFile   string `yaml:"caFile" env:"AGENT_GRPC_TLS_CA" env-default:"" env-description:"CaFile of GRPC server"`
	}

	AuthConfig struct {
		AuthKey string `yaml:"authKey" env:"AGENT_GRPC_AUTH_KEY" env-description:"Auth key of GRPC server"`
		SignKey string `yaml:"signKey" env:"AGENT_GRPC_SIGN_KEY" env-description:"Sign key of GRPC server"`
	}

	ServiceConfig struct {
		LabsCIDR string `yaml:"labsCIDR" env:"LABS_CIDR" env-default:"128.0.0.0/8" env-description:"Labs subnet"`
	}

	RepositoryConfig struct {
		Postgres PostgresConfig `yaml:"postgres"`
	}

	InfrastructureConfig struct {
		Kubernetes KubernetesConfig `yaml:"kubernetes"`
	}

	KubernetesConfig struct {
		KubeConfigPath string `yaml:"kubeConfigPath" env:"KUBE_CONFIG_PATH" env-default:"" env-description:"Path to kubeconfig file"`
		PodsCIDR       string // PodsCIDR is equal to LabsCIDR
	}

	// PostgresConfig is the configuration for the Postgres database
	PostgresConfig struct {
		Host     string `yaml:"host" env:"POSTGRES_HOSTNAME" env-description:"Host of Postgres"`
		Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432" env-description:"Port of Postgres"`
		Username string `yaml:"username" env:"POSTGRES_USER" env-description:"Username of Postgres"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-description:"Password of Postgres"`
		Database string `yaml:"database" env:"POSTGRES_DB" env-description:"Database of Postgres"`
		SSLMode  string `yaml:"sslMode" env:"POSTGRES_SSL_MODE" env-default:"verify-full" env-description:"SSL mode of Postgres"`
	}
)

func MustGetConfig() *Config {
	path := flag.String("config", "", "Path to config file")
	flag.Parse()

	log.Info().Msg("Reading agent configuration")

	instance := &Config{}
	header := "Config variables:"
	help, _ := cleanenv.GetDescription(instance, &header)

	var err error

	if path != nil && *path != "" {
		err = cleanenv.ReadConfig(*path, instance)
	} else {
		err = cleanenv.ReadEnv(instance)
	}

	if err != nil {
		fmt.Println(help)
		log.Fatal().Err(err).Msg("Failed to read config")
		return nil
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// set log mode
	if instance.Environment != Production {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	MigrationPath = "migrations"
	if instance.Environment == Local {
		MigrationPath = "internal/delivery/repository/postgres/migrations"
	}

	instance.populateForAllConfig()

	return instance
}

func (c *Config) populateForAllConfig() {
	c.Infrastructure.Kubernetes.PodsCIDR = c.Service.LabsCIDR
}
