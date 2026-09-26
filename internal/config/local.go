package config

import (
	"os"

	"github.com/zerodotfive/crabcore/pkg/paths"

	yaml "gopkg.in/yaml.v3"
)

type LocalConfig struct {
	URL   string         `yaml:"url"`
	Cache *RuntimeConfig `yaml:"cache"`
}

func (c *LocalConfig) Read() error {
	body, err := os.ReadFile(paths.LocalConfigFile())
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(body, c); err != nil {
		return err
	}

	return nil
}

func (c *LocalConfig) Write() error {
	body, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(paths.LocalConfigFile(), body, 0644)
}
