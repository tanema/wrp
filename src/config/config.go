package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Config descripts a wrp config
type Config struct {
	Destination     string                `json:"destination"`
	Dependencies    map[string]Dependency `json:"dependencies"`
	DependencyLocks map[string]Dependency `json:"dependency_locks"`
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
	jsonFile, err := os.Open(filepath.Join(pwd, "wrp.json"))
	if err != nil {
		return nil, fmt.Errorf("Could not find wrp.json file")
	}
	configBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read wrp.json file")
	}
	if err := jsonFile.Close(); err != nil {
		return nil, err
	}
	return configBytes, err
}

func parseConfigFile(configBytes []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal([]byte(configBytes), config); err != nil {
		return nil, fmt.Errorf("Problem parsing json in wrp.json file: %v", err)
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
func (config *Config) FetchAllDependencies() error {
	urls := []string{}
	for url := range config.Dependencies {
		urls = append(urls, url)
	}
	for _, url := range urls {
		fmt.Println(url)
		config.addDep(url, false)
	}
	return nil
}

func (config *Config) addDep(url string, force bool) error {
	dep := config.Dependencies[url]
	lock := config.DependencyLocks[url]
	if force || dep.requiresUpdate(config.Destination, lock) {
		if err := lock.remove(config.Destination); err != nil {
			return err
		}
		if err := dep.remove(config.Destination); err != nil {
			return err
		}
		if err := dep.fetch(config.Destination, url, config.DependencyLocks[url]); err != nil {
			return err
		}
		if err := dep.write(config.Destination); err != nil {
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
	return config.addDep(url, false)
}

// Add a new dep to config
func (config *Config) Update(url string) error {
	if _, ok := config.Dependencies[url]; !ok {
		return fmt.Errorf("dependency with url %v not found in config", url)
	}
	return config.addDep(url, true)
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
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get working directory: %v", err)
	}
	return ioutil.WriteFile(filepath.Join(pwd, "wrp.json"), configJSON, 0644)
}
