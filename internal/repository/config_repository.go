package repository

import "config-service/internal/model"

type ConfigRepository interface {
	Create(config *model.Config) error
	Get(environment, key string) (*model.Config, error)
	GetAll(environment string) ([]*model.Config, error)
	Update(config *model.Config) error
	Delete(environment, key string) error
	Exists(environment, key string) (bool, error)
}
