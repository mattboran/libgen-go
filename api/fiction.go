package api

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

// FictionSearchInput contains the fields required to search
// Library Genesis' fiction endpoint.
type FictionSearchInput struct {
	Query    []string
	Criteria string
	Format   string
	Page     int
}

type fictionResultParser struct {
	books *[]book
	page  int
}

type fictionMirror struct {
	mirror string
}

// CurrentPage returns the selected page number for the given search input
func (input FictionSearchInput) CurrentPage() int {
	return input.Page
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

func (input FictionSearchInput) url() (*url.URL, error) {
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

func (input FictionSearchInput) resultParser() resultParser {
	return &fictionResultParser{
		books: &[]book{},
		page:  input.Page,
	}
}

func (parser fictionResultParser) currentPage() int {
	return parser.page
}

func (parser fictionResultParser) parsedResults() []DownloadableResult {
	result := []DownloadableResult{}
	for _, book := range *parser.books {
		result = append(result, book)
	}
	return result
}

func (parser fictionResultParser) hasNextPage() bool {
	return (len(*parser.books) % 25) == 0
}

func (parser fictionResultParser) parseResultsFromTableRows() func(int, *goquery.Selection) {

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

func (m fictionMirror) Link() string {
	return m.mirror
}

// DownloadURL performs the required HTTP requests to find the download
// URL for a given mirror url
func (m fictionMirror) DownloadURL(ch chan<- HTTPResult) {
	downloadURLFromGET(m.mirror, ch)
}
