package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	SearchCriteriaAuthors = "authors"
	SearchCriteriaTitle   = "title"
	SearchCriteriaSeries  = "series"
	SortOrderAsc          = "ASC"
	SortOrderDesc         = "DESC"
	SortOrderAuthor       = "author"
	SortOrderTitle        = "title"
	SortOrderPublisher    = "publisher"
	SortOrderYear         = "year"
	SortOrderPage         = "pages"
	SortOrderLanguage     = "language"
	SortOrderID           = "id"
	SortOrderExtension    = "extension"
	SortOrderSize         = "size"
	FormatEPUB            = "epub"
	FormatMOBI            = "mobi"
	FormatAZW             = "azw"
	FormatAZW3            = "azw3"
	FormatFB2             = "fb2"
	FormatPDF             = "pdf"
	FormatRTF             = "rtf"
	FormatTXT             = "txt"
	BaseURL               = "http://gen.lib.rus.ec"
)

// SearchInput is implemented separately by each specific API type
type SearchInput interface {
	CurrentPage() int
	NextPage() SearchInput
	PreviousPage() SearchInput
	resultParser() resultParser
	url() (*url.URL, error)
}

// SearchResults encapsulates the result type and also provides information on
// current and the following page.
type SearchResults struct {
	PageNumber  int
	Results     []DownloadableResult
	HasNextPage bool
}

// DownloadableResult is implemented specifically by each result type
type DownloadableResult interface {
	Name() string
	Mirrors() []string
	Filename() string
}

// resultParser encapsultes the methods required to parse the body of
// an api request into DownloadableResult.
type resultParser interface {
	parsedResults() []DownloadableResult
	currentPage() int
	hasNextPage() bool
	parseResultsFromTableRows() func(int, *goquery.Selection)
}

type book struct {
	authors  []string
	title    string
	language string
	fileType string
	fileSize string
	mirrors  []string
}

// Search takes the SearchInput and returns a pointer to
// SearchResults. It performs the necessary HTTP requests and parses
// the resulting HTML.
func Search(input SearchInput) (*SearchResults, error) {
	url, err := input.url()
	if err != nil {
		return nil, err
	}

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		errorMessage := fmt.Sprintf("Got status code %d", res.StatusCode)
		return nil, errors.New(errorMessage)
	}

	searchResults, err := parseBody(res.Body, input.resultParser())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return searchResults, nil
}

func parseBody(body io.ReadCloser, parser resultParser) (*SearchResults, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	rows := doc.Find("tr")
	rows.Each(parser.parseResultsFromTableRows())

	return &SearchResults{
		PageNumber:  parser.currentPage(),
		Results:     parser.parsedResults(),
		HasNextPage: parser.hasNextPage(),
	}, nil
}

// Name is the displayable name for a Downloadable book
func (b book) Name() string {
	authors := strings.Join(b.authors, ", ")
	return fmt.Sprintf("%s (%s) by %s", b.title, b.fileType, authors)
}

// Mirrors returns the list of mirrors available for a given book
func (b book) Mirrors() []string {
	return b.mirrors
}

// ShortName provides a default filename for use in downloading
func (b book) Filename() string {
	title := strings.ReplaceAll(b.title, " ", "_")
	return fmt.Sprintf("%s.%s", title, strings.ToLower(b.fileType))
}

// HTTPResult is used as a channel input for async HTTP requests and
// document parsing.
type HTTPResult struct {
	Result string
	Error  error
}

// GetDownloadURL performs the required HTTP requests to find the download
// URL for a given mirror url
func GetDownloadURL(mirror string, ch chan<- HTTPResult) {
	res, err := http.Get(mirror)
	if err != nil {
		ch <- HTTPResult{"", err}
		return
	}
	if res.StatusCode != 200 {
		errorMessage := fmt.Sprintf("Got status code %d", res.StatusCode)
		ch <- HTTPResult{"", errors.New(errorMessage)}
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		ch <- HTTPResult{"", err}
		return
	}

	link := doc.Find(":contains(GET) > a")
	if link.Length() == 0 {
		ch <- HTTPResult{"", errors.New("Could not find download link")}
		return
	}

	href, present := link.Attr("href")
	if !present {
		ch <- HTTPResult{"", errors.New("Could not find download link")}
		return
	}
	ch <- HTTPResult{href, nil}
	return
}

// DownloadFile downloads the file from the provided uri to the provided path
func DownloadFile(uri string, filepath string) error {
	res, err := http.Get(uri)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func trim(s string) string {
	var text = strings.ReplaceAll(s, "\n", "")
	text = strings.ReplaceAll(text, "\t", "")
	return text
}
