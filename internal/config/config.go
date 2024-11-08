package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var MigrationPath string

type (
	Config struct {
		Debug      bool             `yaml:"debug" env:"AGENT_DEBUG" env-default:"false" env-description:"Debug mode"`
		Controller ControllerConfig `yaml:"controller"`
		Service    ServiceConfig    `yaml:"service"`
		Repository RepositoryConfig `yaml:"repository"`
	}

	ControllerConfig struct {
		GRPC GRPCConfig `yaml:"grpc"`
	}

	GRPCConfig struct {
		Endpoint string     `yaml:"host" env:"AGENT_GRPC_LISTEN_ENDPOINT" env-default:"0.0.0.0:5454" env-description:"Listen endpoint of GRPC server"`
		TLS      TLSConfig  `yaml:"tls"`
		Auth     AuthConfig `yaml:"auth"`
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

func GetConfig() *Config {
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

	MigrationPath = "migrations"
	if instance.Debug {
		MigrationPath = "internal/delivery/repository/postgres/migrations"
	}

	// set log mode
	if !instance.Debug {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return instance
}
