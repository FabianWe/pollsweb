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
	"github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"time"
)

// variables used for the command parser
var templateRoot, host string
var port int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UTC().UnixNano())
		if templateRoot == "" {
			templateRoot = guessTemplateRoot()
		}
		if !doesDirExist(templateRoot) {
			log.Fatalln("template directory not found, set with \"template-root\"")
		}
		config := getConfig()
		// validate config
		// TODO remove this!
		if ok, validateErr := govalidator.ValidateStruct(config); !ok || validateErr != nil {
			log.Fatalf("invalid config file, validation failed: ok=%v, error=%v\n", ok, validateErr)
		}
		server.RunServerMongo(config, templateRoot, host, port, true)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVar(&templateRoot, "template-root", "", "The directory containing the template files (.gohtml), default is to look for it in the directory where the executable is")
	serveCmd.PersistentFlags().StringVar(&host, "host", "localhost", "The host to run on")
	serveCmd.PersistentFlags().IntVar(&port, "port", 8080, "The port to run on")
}
