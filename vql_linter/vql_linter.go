// How to run: go run vql_linter.go <YAML_FILE>
// package vql_linter
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"www.velocidex.com/golang/velociraptor/config"
	config_proto "www.velocidex.com/golang/velociraptor/config/proto"
	"www.velocidex.com/golang/velociraptor/services"
	"www.velocidex.com/golang/velociraptor/services/orgs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: main <path_to_yaml_file>")
		os.Exit(1)
	}

	serverFilePath, err := saveServerConfigToTmpFile()
	if err != nil {
		fmt.Printf("Failed to save server config to tmp file: %v\n", err)
		os.Exit(1)
	}

	yamlFilePath := os.Args[1]

	// load content of file to yamlContent
	yamlContent, err := os.ReadFile(yamlFilePath)
	// convert it to string
	yamlContentStr := string(yamlContent)

	config_obj, err := new(config.Loader).
		//WithFileLoader("server.config.yaml").
		WithFileLoader(serverFilePath).
		LoadAndValidate()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config_obj.Services = &config_proto.ServerServicesConfig{
		JournalService:    true,
		RepositoryManager: true,
	}
	err = orgs.StartTestOrgManager(ctx, wg, config_obj, nil)

	manager, err := services.GetRepositoryManager(config_obj)
	if err != nil {
		fmt.Printf("Failed to get repository manager: %v\n", err)
		os.Exit(1)
	}

	repository, err := manager.GetGlobalRepository(config_obj)
	if err != nil {
		fmt.Printf("Failed to get global repository: %v\n", err)
		os.Exit(1)
	}

	_, err = repository.LoadYaml(yamlContentStr, services.ArtifactOptions{
		ArtifactIsBuiltIn: true,
		ValidateArtifact:  true,
	})
	if err != nil {
		fmt.Printf("Failed to load YAML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("YAML loaded successfully")
}
