// Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pollsweb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)
import "gopkg.in/yaml.v3"

type ConfigError struct {
	PollWebError
	ErrMessage   string
	WrappedError error
}

func NewConfigError(message string, wrapped error) ConfigError {
	return ConfigError{
		ErrMessage:   message,
		WrappedError: wrapped,
	}
}

func (err ConfigError) Error() string {
	message := "config error: " + err.ErrMessage
	if err.WrappedError != nil {
		message += ". caused by: " + err.WrappedError.Error()
	}
	return message
}

func (err ConfigError) Unwrap() error {
	return err.WrappedError
}

type TimeZone string

func (tz *TimeZone) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.ShortTag() != "!!str" {
		return errors.New("can't unmarshal timezone from yaml")
	}
	_, locationErr := time.LoadLocation(value.Value)
	if locationErr != nil {
		return fmt.Errorf("invalid timezone \"%s\": %w", value.Value, locationErr)
	}
	buff := make([]byte, len(value.Value))
	copy(buff, value.Value)
	*tz = TimeZone(buff)
	return nil
}

func (tz TimeZone) String() string {
	return string(tz)
}

type PostgresConfig struct {
	Host     string
	Port     int32
	User     string
	Password string
	Database string
	SSLMode  string
	Timeout  time.Duration
}

func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Database: "gopolls",
		SSLMode:  "disable",
		Timeout:  time.Duration(30 * time.Second),
	}
}

func (config *PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode)
}

type I18nConfig struct {
	Language string
	Timezone TimeZone
}

func DefaultI18nConfig() *I18nConfig {
	return &I18nConfig{
		Language: "en",
		Timezone: "UTC",
	}
}

type AssetsConfig struct {
	MigrationsDirectory string `yaml:"migrations"`
}

type Config struct {
	Postgres *PostgresConfig
	I18n     *I18nConfig
	Assets   *AssetsConfig
}

func DefaultConfig() *Config {
	res := &Config{
		Postgres: DefaultPostgresConfig(),
		I18n:     DefaultI18nConfig(),
	}

	return res
}

func ReadConfig(config *Config, r io.Reader) (*Config, error) {
	if config == nil {
		config = DefaultConfig()
	}
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	err := decoder.Decode(config)
	if err != nil {
		return nil, NewConfigError("unable to read config", err)
	}
	return config, nil
}

func ReadConfigFile(config *Config, fileName string) (*Config, error) {
	f, err := os.Open(fileName)
	if err != nil {
		// return err directly
		return nil, err
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()
	return ReadConfig(config, f)
}

type AppContext struct {
	*I18nConfig
	Generator *SlugGenerator
}

func NewAppContext(i18n *I18nConfig) *AppContext {
	return &AppContext{
		I18nConfig: i18n,
		Generator:  NewSlugGenerator(i18n.Language),
	}
}
