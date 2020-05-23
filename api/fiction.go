package api

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FictionSearchInput contains the fields required to search
// Library Genesis' fiction endpoint.
type FictionSearchInput struct {
	Query    []string
	Criteria string
	Format   string
	Page     int
}

// FictionSearchCriteria contains the possible Search Criteria strings
var FictionSearchCriteria = []string{
	SearchCriteriaAuthors,
	SearchCriteriaSeries,
	SearchCriteriaTitle,
}

// FictionFormats contains the possible Ebook format strings
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

// URL returns the encoded URL from FictionSearchInput.
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

// NextPage returns a copy of FictionSearchInput but with Page incremented
func (input FictionSearchInput) NextPage() SearchInput {
	return FictionSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		Format:   input.Format,
		Page:     input.Page + 1,
	}
}

// PreviousPage returns a copy of FictionSearchInput but with Page decremented
func (input FictionSearchInput) PreviousPage() SearchInput {
	return FictionSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		Format:   input.Format,
		Page:     input.Page - 1,
	}
}

func (input FictionSearchInput) bodyParser(body io.ReadCloser) (*SearchResults, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tr")
	var books = []Book{}
	rows.Each(parseBooksFromTableRows(&books))

	pageNumbers, err := parsePageNumbers(doc.Find(".page_selector"))
	if err != nil {
		return &SearchResults{
			PageNumber:  1,
			Books:       books,
			HasNextPage: false,
		}, nil
	}
	return &SearchResults{
		PageNumber:  pageNumbers.currentPage,
		Books:       books,
		HasNextPage: pageNumbers.currentPage < pageNumbers.lastPage,
	}, nil
}

type pageNumbers struct {
	currentPage int
	lastPage    int
}

func parsePageNumbers(sel *goquery.Selection) (*pageNumbers, error) {
	pageSelectionText := sel.First().Text()
	pages := strings.Split(pageSelectionText[5:], " / ")
	currentPage, err := strconv.Atoi(pages[0])
	if err != nil {
		return nil, err
	}
	totalPages, err := strconv.Atoi(pages[1])
	if err != nil {
		return nil, err
	}
	return &pageNumbers{currentPage, totalPages}, nil
}

func parseBooksFromTableRows(books *[]Book) func(int, *goquery.Selection) {

	trim := func(s string) string {
		var text = strings.ReplaceAll(s, "\n", "")
		text = strings.ReplaceAll(text, "\t", "")
		return text
	}

	return func(i int, sel *goquery.Selection) {
		if i == 0 {
			return
		}
		var authors, mirrors []string
		var title, language, fileType, fileSize string
		sel.Find("td").Each(func(j int, col *goquery.Selection) {
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

		*books = append(*books, Book{
			Authors:  authors,
			Title:    title,
			Language: language,
			FileType: fileType,
			FileSize: fileSize,
			Mirrors:  mirrors,
		})
	}
}
