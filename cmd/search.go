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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/AlecAivazis/survey/terminal"

	"github.com/mattboran/libgen-go/api"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [string to search for]",
	Short: "Search for a fiction book on Library Genesis",
	Long: `Search for a fiction book on Library Genesis by book title, author name,
	or series name.`,
	Args: cobra.MinimumNArgs(1),
	Run:  handleSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("criteria", "c", "", "Criteria")
	searchCmd.Flags().StringP("format", "f", "", "Result format")
	searchCmd.Flags().IntP("page", "p", 1, "Page number")
}

func processSearchOpt(cmd *cobra.Command, args []string) (*api.FictionSearchInput, error) {
	criteria, err := cmd.Flags().GetString("criteria")
	if err != nil {
		return nil, err
	}
	if criteria != "" && !isContainedInSlice(criteria, api.FictionSearchCriteria) {
		return nil, handleUnsupportedCriteria(criteria)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	if format != "" && !isContainedInSlice(format, api.FictionFormats) {
		return nil, handleUnsupportedFormat(format)
	}

	page, err := cmd.Flags().GetInt("page")
	if err != nil {
		return nil, err
	}

	return &api.FictionSearchInput{
		Query:    args,
		Criteria: criteria,
		Format:   format,
		Page:     page,
	}, nil
}

func options(s *api.SearchResults) []string {
	var options []string
	if s.PageNumber > 1 {
		options = append(options, "back")
	}
	for i, book := range s.Books {
		authors := strings.Join(book.Authors, ", ")
		option := fmt.Sprintf("%d - %s by %s (%s)", i, book.Title, authors, book.FileType)
		options = append(options, option)
	}
	if s.HasNextPage {
		options = append(options, "more")
	}
	options = append(options, "exit")
	return options
}

func handleSearch(cmd *cobra.Command, args []string) {
	input, err := processSearchOpt(cmd, args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if askSurvey(*input) != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func askSurvey(input api.SearchInput) error {
	results, err := api.Search(input)
	if err != nil {
		return err
	}

	choice := ""
	prompt := &survey.Select{
		Options: options(results),
	}
	err = survey.AskOne(prompt, &choice)
	if err == terminal.InterruptErr {
		os.Exit(0)
	} else if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	if choice == "back" {
		return askSurvey(input.PreviousPage())
	}
	if choice == "more" {
		return askSurvey(input.NextPage())
	}
	fmt.Printf("Chose %s\n", choice)
	return nil
}

func handleUnsupportedCriteria(choice string) error {
	supportedCriteriaString := strings.Join(api.FictionSearchCriteria, ", ")
	errorMessage := fmt.Sprintf("%s is not an accepted criteria. Choose from [%s]",
		choice,
		supportedCriteriaString)
	return errors.New(errorMessage)
}

func handleUnsupportedFormat(choice string) error {
	supportedFormatString := strings.Join(api.FictionFormats, ", ")
	errorMessage := fmt.Sprintf("%s is not an accepted criteria. Choose from [%s]",
		choice,
		supportedFormatString)
	return errors.New(errorMessage)
}

func isContainedInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}
