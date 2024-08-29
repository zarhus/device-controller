package main

import (
	"3mdeb/device-controller/pkg/config"
	"3mdeb/device-controller/pkg/controller"
	"3mdeb/device-controller/pkg/server"
	"flag"
	"log"
)

var (
	version        = "0.1.0"
	configDirPath  = "./config"
	configFilePath = flag.String("c", configDirPath+"/config.json",
		"path to configuration file")
	schemaConfigFilePath = flag.String("s", configDirPath+"/config.schema.json",
		"path to schema of configuration file")
)

func main() {
	flag.Parse()
	log.Println("device-controller version:", version)

	cfg, err := config.LoadConfig(*configFilePath, *schemaConfigFilePath)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Initializing controllers")
	err = controller.Init(cfg.Devices)
	if err != nil {
		log.Printf("Error during initialization: %s", err.Error())
	}
	log.Println("Initializing server")
	defer func() {
		log.Println("Cleaning controllers")
		controller.Clean()
	}()
	server.StartServer(*cfg)
}
