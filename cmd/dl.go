/*
Copyright © 2020 Matthew Boran <mattboran@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/AlecAivazis/survey"
	"github.com/AlecAivazis/survey/terminal"
	"github.com/spf13/cobra"
)

// dlCmd represents the dl command
var dlCmd = &cobra.Command{
	Use:   "dl",
	Short: "Set default download path",
	Long:  `Use dl to set a default download path and save it in config.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   handleDlCommand,
}

func init() {
	rootCmd.AddCommand(dlCmd)
}

func handleDlCommand(cmd *cobra.Command, args []string) {
	var directory string
	var err error
	if len(args) == 0 {
		directory, err = launchDLSurvey()
		if err != nil {
			os.Exit(1)
		}
	} else {
		directory = args[0]
		err = validateDirectory(directory)
		if err != nil {
			fmt.Printf("%s is not a valid path\n", directory)
			os.Exit(1)
		}
	}
	viper.Set("download", directory)
	configPath := path.Join(home, cfgFile)
	if err := viper.SafeWriteConfigAs(configPath); err != nil {
		if os.IsNotExist(err) {
			err = viper.WriteConfigAs(configPath)
		}
	}
	if err == nil {
		fmt.Printf("%s set as default download directory.\n", directory)
	} else {
		fmt.Println("Could not save config.")
	}
}

func launchDLSurvey() (string, error) {
	var question = []*survey.Question{surveyQuestionForDownloadDirectory()}
	var directory string
	err := survey.Ask(question, &directory)
	if err == terminal.InterruptErr {
		return "", err
	}
	return directory, nil
}
