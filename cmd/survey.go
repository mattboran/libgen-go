package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/mattboran/libgen-go/api"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
)

func surveyPromptFromResults(results *api.SearchResults) *survey.Select {
	var options []string
	if results.PageNumber > 1 {
		options = append(options, "back")
	}
	for i, result := range results.Results {
		option := fmt.Sprintf("%d - %s", i, result.Name())
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
		option := fmt.Sprintf("[%d] - %s", i, result.Link())
		options = append(options, truncateForTerminalOut(option))
	}
	return &survey.Select{
		Message: "Choose a mirror",
		Options: options,
	}
}

func surveyQuestionForDownloadDirectory() *survey.Question {
	dir := viper.GetString("download")
	return &survey.Question{
		Prompt: &survey.Input{
			Message: "Choose download directory",
			Default: dir,
		},
		Validate: validateDirectory,
	}
}

func surveyQuestionForDownloadFilepath(dir string, d api.DownloadableResult) *survey.Question {
	return &survey.Question{
		Prompt: &survey.Input{
			Message: "Choose a filename",
			Default: d.Filename(),
		},
		Transform: survey.TransformString(func(s string) string {
			return path.Join(dir, s)
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

// askSurvey does the main work of this CLI. It queries for books
// and prepares to follow down a path depending on the results.
func askSurvey(input api.SearchInput) error {
	results, err := api.Search(input)
	if err != nil {
		return err
	}

	if len(results.Results) == 0 && input.CurrentPage() == 1 {
		return errors.New("No results were found")
	}

	// Recursively call this function until a book is selected
	choice := ""
	prompt := surveyPromptFromResults(results)
	err = survey.AskOne(prompt, &choice, nil)
	if err == terminal.InterruptErr {
		return err
	}

	if choice == "back" {
		return askSurvey(input.PreviousPage())
	}
	if choice == "more" {
		return askSurvey(input.NextPage())
	}
	if choice == "exit" {
		return nil
	}

	// Use the choice to select a DownloadableResult based on index
	result, err := getResultFromChoice(choice, results.Results)
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
	go mirror.DownloadURL(ch)

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

// Get the downloadable result based on the string survey choice
func getResultFromChoice(c string, results []api.DownloadableResult) (api.DownloadableResult, error) {
	index, err := strconv.Atoi(strings.Split(c, " ")[0])
	if err != nil {
		return nil, err
	}
	return results[index], nil
}

func surveyChooseMirror(result api.DownloadableResult) (api.Mirror, error) {
	choice := 0
	prompt := surveyPromptForMirrorSelection(result)
	err := survey.AskOne(prompt, &choice, nil)
	return result.Mirrors()[choice], err
}
