// Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/FabianWe/pollsweb/server"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

const (
	// templatesSubDir is the directory name in which the templates reside.
	templatesSubDir = "templates"
)

// getConfig parses the app config from the file passed to the main command of cobra.
// On error this function will end the application.
func getConfig() *server.AppConfig {
	config := server.NewAppConfig()
	unmarshalErr := viper.Unmarshal(config)
	if unmarshalErr != nil {
		log.Fatalln("invalid config file:", unmarshalErr)
	}
	return config
}

// doesDirExist checks if the given path is an existing directory.
func doesDirExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !stat.IsDir() {
		return false
	}
	return true
}

// searchForTemplatesDir will search for the template directory.
// It takes a list of candidates without templatesSubDir, that is it appends templatesSubDir to all candidate and checks
// if this directory exists. The first full path that exists is returned. If no directory was found it returns an empty
// string.
func searchForTemplatesDir(candidateDirs []string) string {
	for _, candidate := range candidateDirs {
		fullPath := filepath.Join(candidate, templatesSubDir)
		if doesDirExist(fullPath) {
			return fullPath
		}
	}
	return ""
}

// guessTemplateRoot guesses the template directory using searchForTemplatesDir.
// It will look in the directory of the application and "./".
// If no template directory is found it will return an empty string.
func guessTemplateRoot() string {
	// first try executable path
	candidates := make([]string, 0, 2)

	if execPath, execPathErr := os.Executable(); execPathErr == nil {
		candidates = append(candidates, filepath.Dir(execPath))
	}
	candidates = append(candidates, "./")
	templateDir := searchForTemplatesDir(candidates)
	if templateDir == "" {
		log.Fatalln("unable to determine template directory, set with \"template-root\"")
	}
	return templateDir
}
