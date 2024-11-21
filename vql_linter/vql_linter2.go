package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/assert"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	api_proto "www.velocidex.com/golang/velociraptor/api/proto"

	artifacts_proto "www.velocidex.com/golang/velociraptor/artifacts/proto"
	"www.velocidex.com/golang/velociraptor/file_store/test_utils"

	flows_proto "www.velocidex.com/golang/velociraptor/flows/proto"
	"www.velocidex.com/golang/velociraptor/services"
	"www.velocidex.com/golang/velociraptor/vql/acl_managers"
)

//go:embed definitions/*
var builtinArtifactDefinitions embed.FS

var (
	app                 = kingpin.New("vql-linter", "VQL linter for Velociraptor YAML artifacts.")
	target              = app.Arg("target", "Path to yaml file or dir with yaml files to lint").Required().String()
	disable_nested_lint = app.Flag("disable-nested-lint", "Disable linting of nested VQLs").Bool()
)

type HuntTestSuite struct {
	test_utils.TestSuite
}

func (self *HuntTestSuite) SetupTest() {
	self.ConfigObj = self.TestSuite.LoadConfig()
	self.ConfigObj.Services.HuntDispatcher = true
	self.TestSuite.SetupTest()
}

func (self *HuntTestSuite) GetAllYamlFilesInDir(dirPath string) []string {
	var yamlFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".yaml" {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return yamlFiles
}

func (self *HuntTestSuite) LoadBuiltinArtifacts(repository services.Repository) ([]*artifacts_proto.Artifact, error) {

	all_artifacts := []*artifacts_proto.Artifact{}

	// Load all artifacts from the embedded filesystem
	err := fs.WalkDir(builtinArtifactDefinitions, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, err := builtinArtifactDefinitions.ReadFile(path)
		artifact, err := repository.LoadYaml(string(data), services.ArtifactOptions{
			ValidateArtifact:  true,
			ArtifactIsBuiltIn: true})
		if err != nil {
			//fmt.Printf("Error: Failed to load built-in artifact %s: %v\n", path, err)
			return nil
		}
		all_artifacts = append(all_artifacts, artifact)
		return nil
	})
	return all_artifacts, err

}

func (self *HuntTestSuite) LoadArtifactsFromPath(path string, repository services.Repository) ([]*artifacts_proto.Artifact, bool) {
	// Get all .yaml files in dir and load them all with repository.LoadYaml

	// If target is file then create a list with one file
	// If target is directory then get all .yaml
	// If target doesn't exist then exit with error
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("Error: %s does not exist\n", path)
		os.Exit(1)
	}

	var yaml_files []string
	if !fileInfo.IsDir() {
		yaml_files = []string{path}
	} else {
		yaml_files = self.GetAllYamlFilesInDir(path)
	}

	artifacts := []*artifacts_proto.Artifact{}

	error_flag := false

	for _, file := range yaml_files {
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		artifact, err := repository.LoadYaml(string(data), services.ArtifactOptions{
			ValidateArtifact:  true,
			ArtifactIsBuiltIn: true})
		if err != nil {
			fmt.Printf("- [%s] Failed to load YAML: %v\n", file, err)
			error_flag = true
		} else {
			artifacts = append(artifacts, artifact)
		}

	}
	return artifacts, error_flag
}

func (self *HuntTestSuite) CompileHunt(artifact_name string) (string, error) {
	// Create a new hunt with the given artifact names
	request := &api_proto.Hunt{
		HuntDescription: "My hunt",
		StartRequest: &flows_proto.ArtifactCollectorArgs{
			Artifacts: []string{artifact_name},
		},
	}

	acl_manager := acl_managers.NullACLManager{}
	hunt_dispatcher, err := services.GetHuntDispatcher(self.ConfigObj)
	assert.NoError(self.T(), err)

	_, err = hunt_dispatcher.CreateHunt(
		self.Ctx, self.ConfigObj, acl_manager, request)

	if err != nil {
		return artifact_name, err
	} else {
		return artifact_name, nil
	}
}

func (self *HuntTestSuite) TestCompilation() {
	manager, err := services.GetRepositoryManager(self.ConfigObj)
	assert.NoError(self.T(), err)

	repository, err := manager.GetGlobalRepository(self.ConfigObj)
	assert.NoError(self.T(), err)

	_, err = self.LoadBuiltinArtifacts(repository)
	if err != nil {
		fmt.Println("Error: Failed to load built-in artifacts")
		os.Exit(1)
	}

	artifacts, error_flag := self.LoadArtifactsFromPath(*target, repository)

	if !*disable_nested_lint {
		for _, artifact := range artifacts {
			//fmt.Println(artifact.Name)
			// Try to compile a hunt with the artifact name

			artifact_name, err := self.CompileHunt(artifact.Name)
			if err != nil {
				// if error contains 'context deadline exceeded' then we print we dont support plugins with files yet
				if strings.Contains(err.Error(), "context deadline exceeded") {
					fmt.Println("% [", artifact_name, "] Linting artifacts with files is not supported yet")
				} else {
					fmt.Println("- [", artifact_name, "] Failed to compile VQL: ", err)
					error_flag = true
				}
			} else {
				fmt.Println("+ [", artifact_name, "] Successfully compiled VQL")
			}

		}
	}

	if error_flag {
		fmt.Println("Error: At least one YAML failed to compile")
		os.Exit(1)
	} else {
		fmt.Println("All YAML files compiled successfully")
	}
}

func main() {
	// Parse command line arguments
	kingpin.MustParse(app.Parse(os.Args[1:]))

	hunt_test := &HuntTestSuite{}
	hunt_test.SetupTest()
	hunt_test.TestCompilation()

}
