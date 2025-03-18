package application

import (
	"service/internal/domains/person"

	"service/internal/domains/api"
)

type Service struct {
	Person *person.Service
	Api    *api.Service
}

func NewService(repo *Repository) *Service {
	return &Service{
		Person: person.NewService(repo.Person),
		Api:    api.NewService(repo.Api),
	}
}
