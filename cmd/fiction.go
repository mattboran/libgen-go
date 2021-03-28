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

	"gopkg.in/AlecAivazis/survey.v1/terminal"

	"github.com/mattboran/libgen-go/api"

	"github.com/spf13/cobra"
)

var searchFictionCmd = &cobra.Command{
	Use:   "fiction [string to search for]",
	Short: "Search for a fiction book on Library Genesis",
	Long: `Search for a fiction book on Library Genesis by book title, author name,
	or series name.`,
	Args: cobra.MinimumNArgs(1),
	Run:  handleFictionSearch,
}

func init() {
	rootCmd.AddCommand(searchFictionCmd)
	searchFictionCmd.Flags().StringP("criteria", "c", "", "Criteria")
	searchFictionCmd.Flags().StringP("format", "f", "", "Result format")
	searchFictionCmd.Flags().IntP("page", "p", 1, "Page number")
}

func processFictionOpt(cmd *cobra.Command, args []string) (*api.FictionSearchInput, error) {
	criteria, err := cmd.Flags().GetString("criteria")
	if err != nil {
		return nil, err
	}
	criteria = strings.ToLower(criteria)
	if criteria != "" && !isContainedInSlice(criteria, api.FictionSearchCriteria) {
		return nil, handleUnsupportedCriteria(criteria)
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	format = strings.ToLower(format)
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

func handleFictionSearch(cmd *cobra.Command, args []string) {
	input, err := processFictionOpt(cmd, args)
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
