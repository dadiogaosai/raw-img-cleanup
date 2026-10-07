package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dadiogaosai/rawtidy/cleanup"
	"gopkg.in/yaml.v3"
)

type config struct {
	DefaultParent string `yaml:"default_parent"`
	LastJPEG      string `yaml:"last_jpeg"`
	LastRAW       string `yaml:"last_raw"`
	home          string
}

func loadConfig(home string) (config, error) {
	path := filepath.Join(home, ".config", "rawtidy", "config")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := config{DefaultParent: home, home: home}
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return config{}, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return config{}, fmt.Errorf("create config directory: %w", err)
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return config{}, fmt.Errorf("create config: %w", err)
		}
		_, writeErr := file.Write(data)
		closeErr := file.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			os.Remove(path)
			return config{}, fmt.Errorf("write config: %w", err)
		}
		return cfg, nil
	}
	if err != nil {
		return config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.home = home
	return cfg, nil
}

func (cfg config) startDirectory(last string) (string, error) {
	for _, candidate := range []string{last, cfg.DefaultParent, cfg.home} {
		if strings.HasPrefix(candidate, "~/") {
			candidate = filepath.Join(cfg.home, strings.TrimPrefix(candidate, "~/"))
		}
		if !filepath.IsAbs(candidate) {
			continue
		}
		if path, err := cleanup.CheckDirectory(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no readable starting directory")
}

func (cfg config) save() error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	path := filepath.Join(cfg.home, ".config", "rawtidy", "config")
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}
