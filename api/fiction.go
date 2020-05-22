package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FictionSearchInput struct {
	Query    []string
	Criteria string
	Format   string
}

var FictionSearchCriteria = []string{
	SearchCriteriaAuthors,
	SearchCriteriaSeries,
	SearchCriteriaTitle,
}

var FictionFormats = []string{
	FormatEPUB,
	FormatMOBI,
	FormatAZW,
	FormatAZW3,
	FormatFB2,
	FormatPDF,
	FormatRTF,
	FormatTXT,
}

func (input FictionSearchInput) URL() (*url.URL, error) {
	params := url.Values{}

	if input.Criteria != "" {
		params.Add("criteria", input.Criteria)
	}

	if input.Format != "" {
		params.Add("format", input.Format)
	}

	params.Add("q", strings.Join(input.Query, " "))

	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path += "fiction"
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func FictionSearch(input *FictionSearchInput) ([]Book, error) {
	var result []Book
	url, err := input.URL()
	if err != nil {
		return result, err
	}

	fmt.Printf("Input url %s\n", url)

	res, err := http.Get(url.String())
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return result, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return result, err
	}

	rows := doc.Find("tr")
	fmt.Printf("Found %d rows\n", rows.Length())
	rows.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		var authors, mirrors []string
		var title, language, fileType, fileSize string
		s.Find("td").Each(func(j int, col *goquery.Selection) {
			switch j {
			case 0:
				text := trim(col.Text())
				authors = strings.Split(text, ";")
			case 2:
				title = strings.ReplaceAll(col.Text(), "\n", "")
			case 3:
				language = strings.ReplaceAll(col.Text(), "\n", "")
			case 4:
				fileSection := strings.Split(trim(col.Text()), " / ")
				fileType = strings.ReplaceAll(fileSection[0], "\n", "")
				fileSize = strings.ReplaceAll(fileSection[1], "\n", "")
			case 5:
				// TODO: this is wrong
				mirrors = strings.Split(col.Text(), " ")
			}
		})
		result = append(result, Book{
			Authors:  authors,
			Title:    trim(title),
			Language: trim(language),
			FileType: trim(fileType),
			FileSize: trim(fileSize),
			Mirrors:  mirrors,
		})
	})

	return result, nil
}

func trim(s string) string {
	var text = strings.ReplaceAll(s, "\n", "")
	text = strings.ReplaceAll(text, "\t", "")
	return text
}
