package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/xeipuuv/gojsonschema"
)

type Endpoint struct {
	Path      string `json:"path"`
	Function  string `json:"function"`
	Type      string `json:"type"`
	Multipart bool   `json:"multipart" default:"false"`
}

type ServerCfg struct {
	ServerAddress string     `json:"server_address"`
	Endpoints     []Endpoint `json:"endpoints"`
}

type Device struct {
	Id            int            `json:"id"`
	Name          string         `json:"name"`
	Controller    string         `json:"controller"`
	ControllerCfg map[string]any `json:"configuration"`
}

type Config struct {
	Server  ServerCfg `json:"server"`
	Devices []Device  `json:"device"`
}

var cfg Config

func LoadConfig(configPath, schemaPath string) (*Config, error) {
	log.Println("Loading configuration")
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewReferenceLoader("file://" + configPath)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	} else if !result.Valid() {
		var validation_errors []error
		for _, error := range result.Errors() {
			validation_errors = append(validation_errors, fmt.Errorf("%s", error))
		}
		return nil, errors.Join(validation_errors...)
	}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		return nil, err
	}

	set := map[int]int{}
	for _, device := range cfg.Devices {
		set[device.Id] = 0
	}
	if len(set) != len(cfg.Devices) {
		return nil, errors.New("duplicate device IDs in config file")
	}

	return &cfg, nil
}

func GetDeviceById(id int) (device *Device, err error) {
	i := slices.IndexFunc(cfg.Devices, func(displ Device) bool {
		return displ.Id == id
	})
	if i == -1 {
		return nil, fmt.Errorf("couldn't find device with ID %d", id)
	}
	return &cfg.Devices[i], nil
}
