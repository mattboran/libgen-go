package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FictionSearchInput struct {
	Query    []string
	Criteria string
	Format   string
	Page     int
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

	params.Add("q", strings.Join(input.Query, " "))
	params.Add("criteria", input.Criteria)
	params.Add("format", input.Format)
	params.Add("page", strconv.Itoa(input.Page))

	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path += "fiction/"
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func (input FictionSearchInput) NextPage() *FictionSearchInput {
	return &FictionSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		Format:   input.Format,
		Page:     input.Page + 1,
	}
}

func (input FictionSearchInput) PreviousPage() *FictionSearchInput {
	return &FictionSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		Format:   input.Format,
		Page:     input.Page - 1,
	}
}

func FictionSearch(input *FictionSearchInput) (*SearchResults, error) {
	url, err := input.URL()
	if err != nil {
		return nil, err
	}

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		errorMessage := fmt.Sprintf("Got status code %d", res.StatusCode)
		return nil, errors.New(errorMessage)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tr")
	var books []Book
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
				title = trim(col.Text())
			case 3:
				language = trim(col.Text())
			case 4:
				fileSection := strings.Split(trim(col.Text()), " / ")
				fileType = trim(fileSection[0])
				fileSize = trim(fileSection[1])
			case 5:
				col.Find("a[href]").Each(func(k int, item *goquery.Selection) {
					href, _ := item.Attr("href")
					mirrors = append(mirrors, href)
				})
			}
		})
		books = append(books, Book{
			Authors:  authors,
			Title:    title,
			Language: language,
			FileType: fileType,
			FileSize: fileSize,
			Mirrors:  mirrors,
		})
	})

	pageSelectionText := doc.Find(".page_selector").First().Text()
	pages := strings.Split(pageSelectionText[5:], " / ")
	var hasNextPage bool
	totalPages, err := strconv.Atoi(pages[1])
	if err != nil {
		hasNextPage = false
	} else {
		hasNextPage = totalPages > input.Page
	}

	return &SearchResults{
		PageNumber:  input.Page,
		Books:       books,
		HasNextPage: hasNextPage,
	}, nil
}

func trim(s string) string {
	var text = strings.ReplaceAll(s, "\n", "")
	text = strings.ReplaceAll(text, "\t", "")
	return text
}
