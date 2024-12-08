package repository

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/delivery/repository/postgres"
)

type (
	Repository struct {
		*postgres.PostgresRepository
	}

	Dependencies struct {
		Config *config.RepositoryConfig
	}
)

func NewRepository(deps Dependencies) *Repository {
	return &Repository{
		postgres.NewRepository(&deps.Config.Postgres),
	}
}

func (r *Repository) Close() {
	r.PostgresRepository.Close()
}
