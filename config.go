package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const config_file_name = "/.gatorconfig.json"

type config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func get_config_file_path() (string, error) {
	home_path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home_path, config_file_name), nil
}

func read() (config, error) {
	this_config := config{}
	config_file_path, err := get_config_file_path()
	if err != nil {
		return this_config, err
	}
	config_file, err := os.Open(config_file_path)
	if err != nil {
		return this_config, err
	}
	defer config_file.Close()

	decoder := json.NewDecoder(config_file)
	err = decoder.Decode(&this_config)
	if err != nil {
		return this_config, err
	}

	return this_config, err
}

func (blog_config *config) set_user(user_name string) error {
	config_file_path, err := get_config_file_path()
	if err != nil {
		return err
	}
	config_file, err := os.Create(config_file_path)
	if err != nil {
		return err
	}
	defer config_file.Close()

	encoder := json.NewEncoder(config_file)
	blog_config.CurrentUserName = user_name
	err = encoder.Encode(blog_config)
	if err != nil {
		return err
	}

	return nil
}
