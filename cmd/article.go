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

	"github.com/AlecAivazis/survey/terminal"
	"github.com/mattboran/libgen-go/api"
	"github.com/spf13/cobra"
)

// articleCmd represents the article command
var articleCmd = &cobra.Command{
	Use:   "article [string to search for]",
	Short: "Search for a scientific article on Library Genesis",
	Long:  `Search for a textbook on Library Genesis by title or author name`,
	Args:  cobra.MinimumNArgs(1),
	Run:   handleArticleSearch,
}

func init() {
	rootCmd.AddCommand(articleCmd)
	articleCmd.Flags().IntP("page", "p", 1, "Page number")
}

func processArticleOpt(cmd *cobra.Command, args []string) (*api.ArticleSearchInput, error) {

	page, err := cmd.Flags().GetInt("page")
	if err != nil {
		return nil, err
	}

	return &api.ArticleSearchInput{
		Query: args,
		Page:  page,
	}, nil
}

func handleArticleSearch(cmd *cobra.Command, args []string) {
	input, err := processArticleOpt(cmd, args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = askSurvey(*input)
	if err == terminal.InterruptErr {
		os.Exit(0)
	} else if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}
