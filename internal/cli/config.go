package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "config.json"

type accountConfig struct {
	CredentialKind string `json:"credentialKind"`
}

type hostConfig struct {
	Server         string                   `json:"server"`
	ActiveUsername string                   `json:"activeUsername"`
	Accounts       map[string]accountConfig `json:"accounts"`
}

type config struct {
	ActiveHostname string                `json:"activeHostname"`
	Hosts          map[string]hostConfig `json:"hosts"`
}

type configStore interface {
	Load() (config, error)
	Save(config) error
}

type fileConfigStore struct {
	path string
}

func newFileConfigStore(path string) *fileConfigStore {
	return &fileConfigStore{path: path}
}

func defaultConfigPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}
	return filepath.Join(root, "nextcloud-tables", configFileName), nil
}

func (s *fileConfigStore) Load() (config, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return newConfig(), nil
	}
	if err != nil {
		return config{}, fmt.Errorf("read config: %w", err)
	}
	var result config
	if err := json.Unmarshal(data, &result); err != nil {
		return config{}, fmt.Errorf("decode config: %w", err)
	}
	normalizeConfig(&result)
	return result, nil
}

func (s *fileConfigStore) Save(value config) error {
	normalizeConfig(&value)
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(s.path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func newConfig() config {
	return config{Hosts: map[string]hostConfig{}}
}

func normalizeConfig(value *config) {
	if value.Hosts == nil {
		value.Hosts = map[string]hostConfig{}
	}
	for hostname, host := range value.Hosts {
		if host.Accounts == nil {
			host.Accounts = map[string]accountConfig{}
			value.Hosts[hostname] = host
		}
	}
}
