/*
Copyright © 2020 Matt Boran <mattboran@gmail.com>

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

	"github.com/AlecAivazis/survey/terminal"
	"github.com/mattboran/libgen-go/api"
	"github.com/spf13/cobra"
)

// textbookCmd represents the textbook command
var textbookCmd = &cobra.Command{
	Use:   "textbook [string to search for]",
	Short: "Search for a textbook on Library Genesis",
	Long:  `Search for a textbook on Library Genesis by book title or author name`,
	Args:  cobra.MinimumNArgs(1),
	Run:   handleTextbookSearch,
}

func init() {
	rootCmd.AddCommand(textbookCmd)
	textbookCmd.Flags().StringP("criteria", "c", "", "Criteria")
	textbookCmd.Flags().StringP("sort", "s", "", "Sort criteria")
	textbookCmd.Flags().BoolP("reverse", "r", false, "Reverse sort order")
	textbookCmd.Flags().IntP("page", "p", 1, "Page number")
}

func processTextbookOpt(cmd *cobra.Command, args []string) (*api.TextbookSearchInput, error) {
	criteria, err := cmd.Flags().GetString("criteria")
	if err != nil {
		return nil, err
	}
	criteria = strings.ToLower(criteria)
	if criteria != "" && !isContainedInSlice(criteria, api.TextbookSearchCriteria) {
		return nil, handleUnsupportedCriteria(criteria)
	}

	sortBy, err := cmd.Flags().GetString("sort")
	if err != nil {
		return nil, err
	}
	sortBy = strings.ToLower(sortBy)
	if sortBy != "" && !isContainedInSlice(sortBy, api.TextbookSortOrder) {
		return nil, handleUnsupportedSortBy(sortBy)
	}

	page, err := cmd.Flags().GetInt("page")
	if err != nil {
		return nil, err
	}

	reverse, err := cmd.Flags().GetBool("reverse")
	if err != nil {
		return nil, err
	}
	sortOrder := api.SortOrderDesc
	if reverse {
		sortOrder = api.SortOrderAsc
	}

	return &api.TextbookSearchInput{
		Query:     args,
		Criteria:  criteria,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
	}, nil
}

func handleTextbookSearch(cmd *cobra.Command, args []string) {
	input, err := processTextbookOpt(cmd, args)
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

func handleUnsupportedSortBy(choice string) error {
	supportedCriteriaString := strings.Join(api.TextbookSortOrder, ", ")
	errorMessage := fmt.Sprintf("%s is not an accepted sort order. Choose from [%s]",
		choice,
		supportedCriteriaString)
	return errors.New(errorMessage)
}
