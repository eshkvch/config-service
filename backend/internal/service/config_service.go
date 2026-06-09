package service

import (
	"config-service/backend/internal/model"
	"config-service/backend/internal/repository"
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

	if err := s.repo.Create(config); err != nil {
		if errors.Is(err, repository.ErrConfigAlreadyExists) {
			return ErrConfigExists
		}
		return err
	}

	return nil
}

func (s *configService) GetConfig(environment, key string) (*model.Config, error) {
	config, err := s.repo.Get(environment, key)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}
	return config, nil
}

func (s *configService) GetAllConfigs(environment string) ([]*model.Config, error) {
	return s.repo.GetAll(environment)
}

func (s *configService) UpdateConfig(environment, key, value string) error {
	config, err := s.repo.Get(environment, key)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return ErrConfigNotFound
		}
		return err
	}

	if err := config.UpdateValue(value); err != nil {
		return err
	}

	if err := s.repo.Update(config); err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return ErrConfigNotFound
		}
		return err
	}

	return nil
}

func (s *configService) DeleteConfig(environment, key string) error {
	exists, err := s.repo.Exists(environment, key)
	if err != nil {
		return err
	}
	if !exists {
		return ErrConfigNotFound
	}

	if err := s.repo.Delete(environment, key); err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return ErrConfigNotFound
		}
		return err
	}

	return nil
}
