package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/tanema/gluey"
)

// ErrNotFound is the error returns when the config is not present
var ErrNotFound = fmt.Errorf("Could not find wrp.yaml file")

// Config descripts a wrp config
type Config struct {
	UI              *gluey.Ctx            `yaml:"-"`
	Destination     string                `yaml:"destination"`
	Dependencies    map[string]Dependency `yaml:"dependencies"`
	DependencyLocks map[string]Dependency `yaml:"dependency_locks,omitempty"`
}

// Default is the default config
var Default = &Config{
	Destination:     "vnd",
	Dependencies:    map[string]Dependency{},
	DependencyLocks: map[string]Dependency{},
}

// Parse will find and parse the config file
func Parse() (*Config, error) {
	configBytes, err := findConfig()
	if err != nil {
		return nil, err
	}
	return parseConfigFile(configBytes)
}

func findConfig() ([]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Could not get working directory: %v", err)
	}
	yamlFile, err := os.Open(filepath.Join(pwd, "wrp.yaml"))
	if err != nil {
		return nil, ErrNotFound
	}
	configBytes, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read wrp.yaml file")
	}
	if err := yamlFile.Close(); err != nil {
		return nil, err
	}
	return configBytes, err
}

func parseConfigFile(configBytes []byte) (*Config, error) {
	config := &Config{UI: gluey.New()}
	if err := yaml.Unmarshal([]byte(configBytes), config); err != nil {
		return nil, fmt.Errorf("Problem parsing yaml in wrp.yaml file: %v", err)
	}
	if config.Destination == "" {
		config.Destination = "vnd"
	}
	if config.Dependencies == nil {
		config.Dependencies = map[string]Dependency{}
	}
	if config.DependencyLocks == nil {
		config.DependencyLocks = map[string]Dependency{}
	}
	return config, nil
}

// FetchAllDependencies will fetch the dependencies
func (config *Config) FetchAllDependencies(force bool) error {
	urls := []string{}
	for url := range config.Dependencies {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	title := "Checking dependencies"
	if force {
		title = "Updating dependencies"
	}

	return config.UI.InMeasuredFrame(title, func(ctx *gluey.Ctx, frame *gluey.Frame) error {
		group := ctx.NewSpinGroup()
		for _, url := range urls {
			depurl := url
			group.Go(depurl, func() error {
				return config.addDep(depurl, force)
			})
		}
		return config.UI.Debreif(group.Wait())
	})
}

func (config *Config) addDep(url string, force bool) (err error) {
	dep := config.Dependencies[url]
	lock := config.DependencyLocks[url]
	if force || dep.requiresUpdate(config.Destination, lock) {
		if err = lock.remove(config.Destination); err != nil {
			return err
		}
		if err = dep.remove(config.Destination); err != nil {
			return err
		}
		if err = dep.fetch(config.Destination, url, config.DependencyLocks[url]); err != nil {
			return err
		}
		if err = dep.write(config.Destination); err != nil {
			return err
		}
		config.DependencyLocks[url] = dep
	}
	return nil
}

// Add a new dep to config
func (config *Config) Add(url string, pick []string) error {
	tag := ""
	if parts := strings.SplitN(url, "@", 2); len(parts) > 1 {
		url = parts[0]
		tag = parts[1]
	}
	config.Dependencies[url] = Dependency{Pick: pick, Tag: tag}
	return config.UI.Spinner(url, func() error { return config.addDep(url, false) })
}

// Update a new dep to config
func (config *Config) Update(url string) error {
	if _, ok := config.Dependencies[url]; !ok {
		return fmt.Errorf("dependency with url %v not found in config", url)
	}
	return config.UI.Spinner(url, func() error { return config.addDep(url, true) })
}

// Remove will remove a dependency
func (config *Config) Remove(url string) error {
	dep, ok := config.DependencyLocks[url]
	if !ok {
		dep, ok = config.Dependencies[url]
		if !ok {
			return fmt.Errorf("dependency %v is not part of the project", url)
		}
	}
	if err := dep.remove(config.Destination); err != nil {
		return err
	}
	delete(config.Dependencies, url)
	delete(config.DependencyLocks, url)
	return nil
}

// Save writes the config back out to save any pinned versions
func (config *Config) Save() error {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get working directory: %v", err)
	}
	return ioutil.WriteFile(filepath.Join(pwd, "wrp.yaml"), configYaml, 0644)
}
