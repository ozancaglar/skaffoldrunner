package parser

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"os"
)

type Config struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Requires   []Requires `yaml:"requires"`
}

type Requires struct {
	Configs []string `yaml:"configs"`
	Path    string   `yaml:"path"`
}

func ParseYamlForModules(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {

		return nil, fmt.Errorf("error reading file: %w", err)
	}

	t := Config{}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml: %w", err)
	}

	var modules []string
	for _, v := range t.Requires {
		modules = append(modules, v.Configs[0])
	}
	return modules, nil
}
