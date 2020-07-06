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
	"fmt"
	"github.com/FabianWe/pollsweb/server"
	"github.com/spf13/cobra"
	"os"
)

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
		templateRoot, getErr := cmd.Flags().GetString("template-root")
		if getErr != nil {
			fmt.Println("can't get flag \"template-root\"")
			os.Exit(1)
		}
		if templateRoot == "" {
			templateRoot = guessTemplateRoot()
		}
		if !doesDirExist(templateRoot) {
			fmt.Println("template directory not found, set with \"template-root\"")
			os.Exit(1)
		}
		config := getConfig()
		server.RunServerMongo(config, templateRoot, true)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().String("template-root", "", "The directory containing the template files (.gohtml), default is to look for it in the directory where the executable is")
}
