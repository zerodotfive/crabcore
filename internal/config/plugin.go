package config

type Plugin struct {
	Name         string            `yaml:"plugin"`
	Filename     string            `yaml:"filename"`
	URL          map[string]string `yaml:"url"`
	ModuleConfig map[string]string `yaml:"moduleConfig"`
}
