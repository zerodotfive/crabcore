package config

import (
	"fmt"
	"runtime"

	"github.com/zerodotfive/crabcore/pkg/fetch"

	yaml "gopkg.in/yaml.v3"
)

type RuntimeConfig struct {
	BinURL  map[string]string `yaml:"binUrl"`
	Plugins []Plugin          `yaml:"plugins"`
}

func (c *RuntimeConfig) Read(url string) error {
	body, err := fetch.Fetch(url)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(body, c); err != nil {
		return err
	}

	osarch := runtime.GOOS + "-" + runtime.GOARCH
	if _, ok := c.BinURL[osarch]; !ok {
		return fmt.Errorf("no bin url platform: %s", osarch)
	}

	return nil
}
