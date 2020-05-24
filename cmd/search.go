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
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"

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

func surveyPromptFromResults(results *api.SearchResults) *survey.Select {
	var options []string
	if results.PageNumber > 1 {
		options = append(options, "back")
	}
	for i, book := range results.Books {
		option := fmt.Sprintf("%d - %s", i, book.Name())
		options = append(options, truncateForTerminalOut(option))
	}
	if results.HasNextPage {
		options = append(options, "more")
	}
	options = append(options, "exit")

	return &survey.Select{
		Options: options,
	}
}

func surveyPromptForMirrorSelection(selection api.DownloadableResult) *survey.Select {
	var options []string
	for i, result := range selection.Mirrors() {
		option := fmt.Sprintf("[%d] - %s", i, truncateForTerminalOut(result))
		options = append(options, option)
	}
	return &survey.Select{
		Message: "Choose a mirror",
		Options: options,
	}
}

func surveyQuestionForDownloadDirectory() *survey.Question {
	dir, err := homedir.Dir()
	if err != nil {
		dir, _ = os.Getwd()
	}
	return &survey.Question{
		Prompt: &survey.Input{
			Message: "Choose download directory",
			Default: dir,
		},
		Validate: func(val interface{}) error {
			path, _ := val.(string)
			_, err := os.Stat(path)
			return err
		},
	}
}

func surveyQuestionForDownloadFilepath(dir string, d api.DownloadableResult) *survey.Question {
	// Todo: os path join intsead of + "/"
	return &survey.Question{
		Prompt: &survey.Input{
			Message: "Choose a filename",
			Default: d.Filename(),
		},
		Transform: survey.TransformString(func(s string) string {
			return dir + "/" + s
		}),
		Validate: func(val interface{}) error {
			filename, _ := val.(string)
			_, err := os.Stat(filename)
			if err == nil {
				return errors.New("File already exists")
			}
			return nil
		},
	}
}

func handleSearch(cmd *cobra.Command, args []string) {
	input, err := processSearchOpt(cmd, args)
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

func askSurvey(input api.SearchInput) error {
	results, err := api.Search(input)
	if err != nil {
		return err
	}

	// Recursively call this function until a book is selected
	choice := ""
	prompt := surveyPromptFromResults(results)
	err = survey.AskOne(prompt, &choice)
	if err == terminal.InterruptErr {
		return err
	}

	if choice == "back" {
		return askSurvey(input.PreviousPage())
	}
	if choice == "more" {
		return askSurvey(input.NextPage())
	}

	// Use the choice to select a DownloadableResult based on index
	result, err := getResultFromChoice(choice, results.Books)
	if err != nil {
		return err
	}
	// Prompt the user to choose one of the mirrors to download from
	mirror, err := surveyChooseMirror(result)
	if err != nil {
		return err
	}

	// Get the download URL asynchronously as the user is prompted
	// for download location.
	ch := make(chan api.HTTPResult, 0)
	go api.GetDownloadURL(mirror, ch)

	// Prompt for the download directory and filename
	dir := ""
	var dirQuestion = []*survey.Question{surveyQuestionForDownloadDirectory()}
	err = survey.Ask(dirQuestion, &dir)
	if err != nil {
		return err
	}
	filepath := ""
	var pathQuestion = []*survey.Question{surveyQuestionForDownloadFilepath(dir, result)}
	err = survey.Ask(pathQuestion, &filepath)
	if err != nil {
		return err
	}

	downloadURLResult := <-ch
	if downloadURLResult.Error != nil {
		return downloadURLResult.Error
	}
	err = api.DownloadFile(downloadURLResult.Result, filepath)
	if err != nil {
		return err
	}

	fmt.Printf("Saved to %s\n", filepath)
	return nil
}

func getResultFromChoice(c string, results []api.DownloadableResult) (api.DownloadableResult, error) {
	index, err := strconv.Atoi(strings.Split(c, " ")[0])
	if err != nil {
		return nil, err
	}
	return results[index], nil
}

func surveyChooseMirror(result api.DownloadableResult) (string, error) {
	choice := 0
	prompt := surveyPromptForMirrorSelection(result)
	err := survey.AskOne(prompt, &choice)
	return result.Mirrors()[choice], err
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
