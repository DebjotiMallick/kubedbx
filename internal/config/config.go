package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Database struct {
	Hostname  string   `yaml:"hostname"` // Renamed to Hostname for clarity
	Namespace string   `yaml:"namespace"`
	TLSSecret string   `yaml:"tlsSecret"`
	Database  []string `yaml:"database"` // Changed to []string
}

type Config map[string][]Database // Changed to a map directly

func LoadConfig(configFilePath string) (Config, error) {
	if configFilePath == "" {
		configFilePath = "/etc/config/databases.yaml" // Default path
	}

	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading YAML file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	// Validation (Corrected and Simplified)
	for dbType, dbs := range config {
		for _, db := range dbs {
			switch dbType {
			case "mongodb":
				if db.Hostname == "" || db.Namespace == "" {
					log.Printf("Warning: MongoDB database in %s is missing required config parameters", db.Namespace)
				}
			case "mysql":
				if db.Hostname == "" || db.Namespace == "" {
					log.Printf("Warning: MySQL database in %s is missing required config parameters", db.Namespace)
				}
			case "postgresql":
				if db.Hostname == "" || db.Namespace == "" {
					log.Printf("Warning: PostgreSQL database in %s is missing required config parameters", db.Namespace)
				}
			// Add similar checks for other database types
			default:
				log.Printf("Warning: Unknown database type: %s", dbType) // Log unknown types
			}
		}
	}

	return config, nil
}
