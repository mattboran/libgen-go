package api

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TextbookSearchInput contains the fields required to search
// Library Genesis' fiction endpoint.
type TextbookSearchInput struct {
	Query    []string
	Criteria string
	SortBy   string
	Page     int
}

// TextbookSearchCriteria contains the possible Search Criteria strings
var TextbookSearchCriteria = []string{
	SearchCriteriaAuthors,
	SearchCriteriaTitle,
}

// CurrentPage returns the selected page number for the given search input
func (input TextbookSearchInput) CurrentPage() int {
	return input.Page
}

// NextPage returns a copy of FictionSearchInput but with Page incremented
func (input TextbookSearchInput) NextPage() SearchInput {
	return TextbookSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		SortBy:   input.SortBy,
		Page:     input.Page + 1,
	}
}

// PreviousPage returns a copy of FictionSearchInput but with Page decremented
func (input TextbookSearchInput) PreviousPage() SearchInput {
	return TextbookSearchInput{
		Query:    input.Query,
		Criteria: input.Criteria,
		SortBy:   input.SortBy,
		Page:     input.Page - 1,
	}
}

func (input TextbookSearchInput) url() (*url.URL, error) {
	params := url.Values{}

	params.Add("req", strings.Join(input.Query, " "))
	params.Add("column", input.Criteria)
	params.Add("page", strconv.Itoa(input.Page))

	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path += "search.php"
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func (input TextbookSearchInput) bodyParser(body io.ReadCloser) (*SearchResults, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tr")
	var results = []DownloadableResult{}
	rows.Each(parseTextbooksFromTableRows(&results))

	currentPage := input.CurrentPage()
	lastPage, err := parseNumberOfTextbookPages(doc)

	if err != nil {
		return &SearchResults{
			PageNumber:  1,
			Books:       results,
			HasNextPage: false,
		}, nil
	}
	return &SearchResults{
		PageNumber:  currentPage,
		Books:       results,
		HasNextPage: currentPage < lastPage,
	}, nil
}

func parseNumberOfTextbookPages(doc *goquery.Document) (int, error) {
	sel := doc.Find("#paginator_example_top.td")
	return sel.Length() - 1, nil
}

func parseTextbooksFromTableRows(books *[]DownloadableResult) func(int, *goquery.Selection) {

	trim := func(s string) string {
		var text = strings.ReplaceAll(s, "\n", "")
		text = strings.ReplaceAll(text, "\t", "")
		return text
	}

	extractMirror := func(sel *goquery.Selection) string {
		link := sel.Find("a[href]").First()
		href, _ := link.Attr("href")
		return href
	}

	return func(i int, sel *goquery.Selection) {
		if i < 4 {
			return
		}
		var authors, mirrors []string
		var title, language, fileType, fileSize string
		sel.Find("td").Each(func(j int, col *goquery.Selection) {
			switch j {
			case 1:
				text := trim(col.Text())
				authors = []string{text}
			case 2:
				title = trim(col.Text())
			case 5:
				language = trim(col.Text())
			case 7:
				fileSize = trim(col.Text())
			case 8:
				fileType = trim(col.Text())
			case 9, 10, 11, 12, 13:
				mirrors = append(mirrors, extractMirror(col))
			}
		})

		*books = append(*books, book{
			authors:  authors,
			title:    title,
			language: language,
			fileType: fileType,
			fileSize: fileSize,
			mirrors:  mirrors,
		})
	}
}
