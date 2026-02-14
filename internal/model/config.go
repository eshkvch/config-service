package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidEnvironment = errors.New("invalid environment name")
	ErrInvalidKey         = errors.New("invalid key")
	ErrInvalidValue       = errors.New("invalid value")
)

type Config struct {
	Environment string    `json:"env"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func NewConfig(environment, key, value string) (*Config, error) {
	if err := validateEnvironment(environment); err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if err := validateValue(value); err != nil {
		return nil, err
	}

	return &Config{
		Environment: environment,
		Key:         key,
		Value:       value,
		UpdatedAt:   time.Now(),
	}, nil
}

func (c *Config) UpdateValue(value string) error {
	if err := validateValue(value); err != nil {
		return err
	}
	c.Value = value
	c.UpdatedAt = time.Now()
	return nil
}

func validateEnvironment(env string) error {
	if env == "" {
		return ErrInvalidEnvironment
	}
	if len(env) > 100 {
		return ErrInvalidEnvironment
	}
	return nil
}

func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	if len(key) > 255 {
		return ErrInvalidKey
	}
	return nil
}

func validateValue(value string) error {
	if len(value) > 10000 {
		return ErrInvalidValue
	}
	return nil
}
