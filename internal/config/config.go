package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/msherburne/dokbox/internal/docker"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/theme"
)

const CurrentConfigVersion = 1

func SaveConfig(cfg domain.Config) error {
	configPath, err := Path()
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, payload, 0o600)
}

func LoadRawConfig() (map[string]any, error) {
	configPath, err := Path()
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func LoadConfig() (*domain.Config, error) {
	raw, err := LoadRawConfig()
	if err != nil || raw == nil {
		return nil, err
	}

	cfg, err := decodeConfig(raw)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func MigrateConfig(raw map[string]any) map[string]any {
	if raw == nil {
		return nil
	}

	migrated := make(map[string]any, len(raw)+1)
	for key, value := range raw {
		migrated[key] = value
	}

	if _, ok := migrated["config_version"]; !ok {
		migrated["config_version"] = CurrentConfigVersion
	}

	return migrated
}

func ResolveCurrentConfig() (*domain.Config, error) {
	raw, err := LoadRawConfig()
	if err != nil || raw == nil {
		return nil, err
	}

	cfg, err := decodeConfig(MigrateConfig(raw))
	if err != nil {
		return nil, err
	}

	if err := SaveConfig(*cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func GenerateDefaultConfig(writer io.Writer) (domain.Config, error) {
	if writer == nil {
		writer = io.Discard
	}

	cfg := domain.Config{
		ConfigVersion: CurrentConfigVersion,
		DockerHost:    docker.DetectDockerHost(),
		Theme:         theme.DefaultPreset,
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return domain.Config{}, err
	}

	if _, err := fmt.Fprintf(writer, "Using default configuration:\n%s\n", payload); err != nil {
		return domain.Config{}, err
	}

	return cfg, nil
}

func LoadOrSetupConfig(reader io.Reader, writer io.Writer) (domain.Config, error) {
	cfg, err := ResolveCurrentConfig()
	if err != nil {
		return domain.Config{}, err
	}
	if cfg != nil {
		return *cfg, nil
	}

	generated, err := GenerateDefaultConfig(writer)
	if err != nil {
		return domain.Config{}, err
	}

	confirmed, err := ConfirmSave(reader, writer, "Do you want to save this configuration?")
	if err != nil {
		return domain.Config{}, err
	}
	if confirmed {
		if err := SaveConfig(generated); err != nil {
			return domain.Config{}, err
		}
	}

	return generated, nil
}

func decodeConfig(raw map[string]any) (*domain.Config, error) {
	payload, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return nil, err
	}

	if cfg.Theme == "" {
		cfg.Theme = theme.DefaultPreset
	}

	return &cfg, nil
}
