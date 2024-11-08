// How to run: go run vql_linter.go <YAML_FILE>
// package vql_linter
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"www.velocidex.com/golang/velociraptor/config"
	config_proto "www.velocidex.com/golang/velociraptor/config/proto"
	"www.velocidex.com/golang/velociraptor/services"
	"www.velocidex.com/golang/velociraptor/services/orgs"
)

// Takes 1 argument yaml file or directory with yaml files to run linter on
// and optionally flag -r to search yamls in subdirectories

var (
	app      = kingpin.New("vql-linter", "VQL linter for Velociraptor YAML artifacts.")
	filename = app.Arg("target", "Yaml file or dir with yaml files to lint").Required().String()
	verbose  = app.Flag("verbose", "Verbose output").Short('v').Bool()
)

func getFilesInDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func main() {
	// Parse command line arguments
	kingpin.MustParse(app.Parse(os.Args[1:]))
	// Check if argument is a file or directory
	fileInfo, err := os.Stat(*filename)
	if err != nil {
		fmt.Printf("Failed to get file info: %v\n", err)
		os.Exit(1)
	}

	is_dir := fileInfo.IsDir()
	var yamlFiles []string

	if !is_dir {
		yamlFiles = append(yamlFiles, *filename)
	} else {
		// If it's a directory, get all yaml files in it
		yamlFiles, err = getFilesInDir(*filename)
		if err != nil {
			fmt.Printf("Failed to get files in directory: %v\n", err)
			os.Exit(1)
		}
	}

	serverFilePath, err := saveServerConfigToTmpFile()
	if err != nil {
		fmt.Printf("Failed to save server config to tmp file: %v\n", err)
		os.Exit(1)
	}

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

	returnCode := 0

	for _, yamlFile := range yamlFiles {
		yamlContent, err := os.ReadFile(yamlFile)
		if err != nil {
			fmt.Printf("Failed to read YAML file: %v\n", err)
			os.Exit(1)
		}
		yamlContentStr := string(yamlContent)

		_, err = repository.LoadYaml(yamlContentStr, services.ArtifactOptions{
			ArtifactIsBuiltIn: true,
			ValidateArtifact:  true,
		})
		if err != nil {
			fmt.Printf("[%s] Failed to load YAML: %v\n", yamlFile, err)
			returnCode = 1
		} else {
			if *verbose {
				fmt.Printf("[%s] Successfully loaded YAML\n", yamlFile)
			}
		}

	}

	os.Exit(returnCode)
}
