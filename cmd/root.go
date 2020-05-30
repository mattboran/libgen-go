/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// cfgFile contains the filename for the config file. Default libgen.yaml
var cfgFile string

// home contains the directory to save the config file. Default $HOME
var home string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "libgen",
	Short: "A cli for downloading ebooks from Library Genesis",
	Long:  `Libgen is a CLI application for interacting with Library Genesis.`,
	Run:   helpFunc,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.libgen.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	// Find home directory.
	home, err = homedir.Dir()
	if err != nil {
		fmt.Printf("Could not get home directory.\n")
		os.Exit(1)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		cfgFile = ".libgen"
		viper.AddConfigPath(home)
		viper.SetConfigName(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// TODO: store defualt download location in config
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	viper.SetDefault("download", home)
}

func helpFunc(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

var terminalWidth = func() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	termSize := strings.Split(string(out), " ")
	width, err := strconv.Atoi(termSize[1])
	if err != nil {
		return 80
	}
	return width
}()

func isContainedInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func truncateForTerminalOut(s string) string {
	if len(s) < (terminalWidth - 5) {
		return s
	}
	return s[:terminalWidth-5] + "..."
}

func validateDirectory(val interface{}) error {
	path, _ := val.(string)
	_, err := os.Stat(path)
	return err
}
