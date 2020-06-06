package api

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TextbookSortOrder are the accepted parameters for sort order
var TextbookSortOrder = []string{
	SortOrderAuthor,
	SortOrderTitle,
	SortOrderPublisher,
	SortOrderYear,
	SortOrderPage,
	SortOrderLanguage,
	SortOrderID,
	SortOrderExtension,
	SortOrderSize,
}

// SortOrder can either be ASC (default) or DESC
var SortOrder = []string{
	SortOrderAsc,
	SortOrderDesc,
}

// TextbookSearchInput contains the fields required to search
// Library Genesis' fiction endpoint.
type TextbookSearchInput struct {
	Query     []string
	Criteria  string
	SortBy    string
	SortOrder string
	Page      int
}

type textbookResultParser struct {
	books *[]book
	page  int
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
		Query:     input.Query,
		Criteria:  input.Criteria,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
		Page:      input.Page + 1,
	}
}

// PreviousPage returns a copy of FictionSearchInput but with Page decremented
func (input TextbookSearchInput) PreviousPage() SearchInput {
	return TextbookSearchInput{
		Query:     input.Query,
		Criteria:  input.Criteria,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
		Page:      input.Page - 1,
	}
}

func (input TextbookSearchInput) url() (*url.URL, error) {
	params := url.Values{}

	params.Add("req", strings.Join(input.Query, " "))
	params.Add("column", input.Criteria)
	params.Add("page", strconv.Itoa(input.Page))
	params.Add("sort", input.SortBy)
	params.Add("sortmode", input.SortOrder)

	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path += "search.php"
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func (input TextbookSearchInput) resultParser() resultParser {
	return &textbookResultParser{
		books: &[]book{},
		page:  input.Page,
	}
}

func (parser textbookResultParser) currentPage() int {
	return parser.page
}

func (parser textbookResultParser) parsedResults() []DownloadableResult {
	result := []DownloadableResult{}
	for _, book := range *parser.books {
		result = append(result, book)
	}
	return result
}

func (parser textbookResultParser) hasNextPage() bool {
	return (len(*parser.books) % 25) == 0
}

func (parser textbookResultParser) parseResultsFromTableRows() func(int, *goquery.Selection) {

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
		if i < 3 {
			return
		}
		var authors, mirrors []string
		var title, language, fileType, fileSize string
		sel.Find("td").Each(func(j int, col *goquery.Selection) {
			switch j {
			case 1:
				text := trim(col.Text())
				authors = strings.Split(text, ",")
			case 2:
				link := sel.Find("a[title]").First()
				isbns := link.Find("i").Last().Text()
				titleText := link.Text()
				// Suffix is usually the ISBNs. Occasionally this also
				// snips a [2nd ed.] or equivalent if there are no isbns.
				lengthOfSuffix := len(titleText) - len(isbns)
				title = trim(titleText[:lengthOfSuffix])
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

		if len(authors) == 0 || len(mirrors) == 0 {
			return
		}

		*parser.books = append(*parser.books, book{
			authors:  authors,
			title:    title,
			language: language,
			fileType: fileType,
			fileSize: fileSize,
			mirrors:  mirrors,
		})
	}
}
