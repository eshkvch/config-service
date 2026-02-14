package service

import (
	"config-service/internal/model"
	"config-service/internal/repository"
	"errors"
)

var (
	ErrConfigNotFound = errors.New("config not found")
	ErrConfigExists   = errors.New("config already exists")
)

type ConfigService interface {
	CreateConfig(environment, key, value string) error
	GetConfig(environment, key string) (*model.Config, error)
	GetAllConfigs(environment string) ([]*model.Config, error)
	UpdateConfig(environment, key, value string) error
	DeleteConfig(environment, key string) error
}

type configService struct {
	repo repository.ConfigRepository
}

func NewConfigService(repo repository.ConfigRepository) ConfigService {
	return &configService{repo: repo}
}

func (s *configService) CreateConfig(environment, key, value string) error {
	exists, err := s.repo.Exists(environment, key)
	if err != nil {
		return err
	}
	if exists {
		return ErrConfigExists
	}

	config, err := model.NewConfig(environment, key, value)
	if err != nil {
		return err
	}

	return s.repo.Create(config)
}

func (s *configService) GetConfig(environment, key string) (*model.Config, error) {
	config, err := s.repo.Get(environment, key)
	if err != nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}

func (s *configService) GetAllConfigs(environment string) ([]*model.Config, error) {
	return s.repo.GetAll(environment)
}

func (s *configService) UpdateConfig(environment, key, value string) error {
	config, err := s.repo.Get(environment, key)
	if err != nil {
		return ErrConfigNotFound
	}

	if err := config.UpdateValue(value); err != nil {
		return err
	}

	return s.repo.Update(config)
}

func (s *configService) DeleteConfig(environment, key string) error {
	exists, err := s.repo.Exists(environment, key)
	if err != nil {
		return err
	}
	if !exists {
		return ErrConfigNotFound
	}

	return s.repo.Delete(environment, key)
}
