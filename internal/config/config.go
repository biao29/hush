package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// HushConfig is the top-level configuration stored in hush.yaml.
type HushConfig struct {
	Projects []ProjectEntry `yaml:"projects"`
}

// ProjectEntry represents a registered project.
type ProjectEntry struct {
	ID       string `yaml:"id"`
	Remote   string `yaml:"remote"`
	Path     string `yaml:"path"`
	Subdir   string `yaml:"subdir,omitempty"`
	Hostname string `yaml:"hostname,omitempty"`
}

// Load reads hush.yaml from the private repo.
func Load() (*HushConfig, error) {
	return LoadFrom(ConfigPath())
}

// LoadFrom reads a hush.yaml from a specific path.
func LoadFrom(path string) (*HushConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &HushConfig{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg HushConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes hush.yaml to the private repo.
func (c *HushConfig) Save() error {
	return c.SaveTo(ConfigPath())
}

// SaveTo writes hush.yaml to a specific path.
func (c *HushConfig) SaveTo(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// FindProject returns the project entry with the given ID, or nil.
func (c *HushConfig) FindProject(id string) *ProjectEntry {
	for i := range c.Projects {
		if c.Projects[i].ID == id {
			return &c.Projects[i]
		}
	}
	return nil
}

// FindByRemoteAndSubdir returns the project matching remote+subdir, or nil.
func (c *HushConfig) FindByRemoteAndSubdir(remote, subdir string) *ProjectEntry {
	for i := range c.Projects {
		if c.Projects[i].Remote == remote && c.Projects[i].Subdir == subdir {
			return &c.Projects[i]
		}
	}
	return nil
}

// AddProject appends a project entry after checking for duplicates.
func (c *HushConfig) AddProject(p ProjectEntry) error {
	if existing := c.FindProject(p.ID); existing != nil {
		return fmt.Errorf("project %q already registered (remote: %s)", p.ID, existing.Remote)
	}
	if existing := c.FindByRemoteAndSubdir(p.Remote, p.Subdir); existing != nil {
		return fmt.Errorf("remote %s (subdir: %s) already registered as %q", p.Remote, p.Subdir, existing.ID)
	}
	c.Projects = append(c.Projects, p)
	return nil
}

// ProjectsForHost returns projects matching the current hostname.
// Projects without a hostname match all hosts.
func (c *HushConfig) ProjectsForHost(hostname string) []ProjectEntry {
	var result []ProjectEntry
	for _, p := range c.Projects {
		if p.Hostname == "" || p.Hostname == hostname {
			result = append(result, p)
		}
	}
	return result
}
