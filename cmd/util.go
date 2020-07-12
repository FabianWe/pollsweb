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
	templatesSubDir = "templates"
)

func getConfig() *server.AppConfig {
	config := server.NewAppConfig()
	unmarshalErr := viper.Unmarshal(config)
	if unmarshalErr != nil {
		log.Fatalln("invalid config file:", unmarshalErr)
	}
	return config
}

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

func searchForTemplatesDir(candidateDirs []string) string {
	for _, candidate := range candidateDirs {
		fullPath := filepath.Join(candidate, templatesSubDir)
		if doesDirExist(fullPath) {
			return fullPath
		}
	}
	return ""
}

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
