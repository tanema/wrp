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
	Destination  string                 `json:"destination"`
	Dependencies map[string]*Dependency `json:"dependencies"`
}

// Parse will find and parse the config file
func Parse() (*Config, error) {
	jsonFile, err := findConfig()
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	return parseConfigFile(jsonFile)
}

func findConfig() (*os.File, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Could not get working directory: %v", err)
	}
	jsonFile, err := os.Open(filepath.Join(pwd, "wrp.json"))
	if err != nil {
		return nil, fmt.Errorf("Could not find wrp.json file")
	}
	return jsonFile, err
}

func parseConfigFile(file *os.File) (*Config, error) {
	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Could not read wrp.json file")
	}
	config := &Config{}
	if err := json.Unmarshal([]byte(configBytes), config); err != nil {
		return nil, fmt.Errorf("Problem parsing json in wrp.json file: %v", err)
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
		dep := config.Dependencies[url]
		if err := dep.remove(config.Destination); err != nil {
			return err
		}
		if err := dep.fetch(url); err != nil {
			return err
		}
		if path := dep.formatURL(); path != url {
			config.Dependencies[path] = dep
			delete(config.Dependencies, url)
		}
		// TODO: nested deps?
		if err := dep.write(config.Destination); err != nil {
			return err
		}
	}
	return nil
}

// Add will add a new dependency
// name@tag
// name#branch
// name!hash
func (config *Config) Add(url string, pick []string) error {
	dep := &Dependency{Pick: pick}
	if err := dep.fetch(url); err != nil {
		return err
	}
	if err := dep.write(config.Destination); err != nil {
		return err
	}
	config.Dependencies[dep.formatURL()] = dep
	return nil
}

// Remove will remove a dependency
func (config *Config) Remove(url string) error {
	u, dep, err := config.findDep(url)
	if err != nil {
		return err
	}
	if err := dep.remove(config.Destination); err != nil {
		return err
	}
	delete(config.Dependencies, u)
	return nil
}

func (config *Config) findDep(repourl string) (string, *Dependency, error) {
	dep, ok := config.Dependencies[repourl]
	if ok {
		return repourl, dep, nil
	}
	for u := range config.Dependencies {
		if strings.HasPrefix(u, repourl) {
			return u, config.Dependencies[u], nil
		}
	}
	return "", nil, fmt.Errorf("dependency %v is not part of the project", repourl)
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
