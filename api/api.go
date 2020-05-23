package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	SearchCriteriaAuthors = "authors"
	SearchCriteriaTitle   = "title"
	SearchCriteriaSeries  = "series"
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
	URL() (*url.URL, error)
	NextPage() SearchInput
	PreviousPage() SearchInput
	bodyParser(io.ReadCloser) (*SearchResults, error)
}

// Book is the basic return result type
type Book struct {
	Authors  []string
	Title    string
	Language string
	FileType string
	FileSize string
	Mirrors  []string
}

// SearchResults encapsulates the result type and also provides information on
// current and the following page.
type SearchResults struct {
	PageNumber  int
	Books       []Book
	HasNextPage bool
}

// Search takes the SearchInput and returns a pointer to
// SearchResults. It performs the necessary HTTP requests and parses
// the resulting HTML.
func Search(input SearchInput) (*SearchResults, error) {
	url, err := input.URL()
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

	searchResults, err := input.bodyParser(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return searchResults, nil
}
