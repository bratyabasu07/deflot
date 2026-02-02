package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the structure of the configuration file.
type Config struct {
	ApiKeys ApiKeys `mapstructure:"api_keys"`
}

type ApiKeys struct {
	VirusTotal string `mapstructure:"virustotal"`
	URLScan    string `mapstructure:"urlscan"`
	AlienVault string `mapstructure:"alienvault"`
	GitHub     string `mapstructure:"github"`
}

// defaultConfigFileContent defines the default YAML content.
const defaultConfigFileContent = `api_keys:
  virustotal: ""
  urlscan: ""
  alienvault: ""
  github: ""
`

// InitConfig initializes the configuration.
// It reads from ~/.deflot/config.yml if it exists.
func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".deflot")
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Read environment variables with prefix DEFLOT_
	viper.SetEnvPrefix("DEFLOT")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// keeping silent unless debug is needed
	}
}

// GenerateDefaultConfig creates ~/.deflot/config.yml with default values.
func GenerateDefaultConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".deflot")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			os.Exit(1)
		}
	}

	configPath := filepath.Join(configDir, "config.yml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		return
	}

	if err := os.WriteFile(configPath, []byte(defaultConfigFileContent), 0644); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created default config at %s\n", configPath)
}

// GetAPIKeys returns the API keys from the configuration.
func GetAPIKeys() ApiKeys {
	var keys ApiKeys
	if err := viper.UnmarshalKey("api_keys", &keys); err != nil {
		return ApiKeys{}
	}
	return keys
}
