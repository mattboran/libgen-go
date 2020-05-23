package api

import (
	"net/url"
)

const (
	SearchCriteriaAuthors = "authors"
	SearchCriteriaTitle   = "title"
	SearchCriteriaSeries  = "series"
)

const (
	FormatEPUB = "epub"
	FormatMOBI = "mobi"
	FormatAZW  = "azw"
	FormatAZW3 = "azw3"
	FormatFB2  = "fb2"
	FormatPDF  = "pdf"
	FormatRTF  = "rtf"
	FormatTXT  = "txt"
)

const (
	BaseURL = "http://gen.lib.rus.ec"
)

type SearchInput interface {
	URL() (*url.URL, error)
	NextPage() SearchInput
}

type Book struct {
	Authors  []string
	Title    string
	Language string
	FileType string
	FileSize string
	Mirrors  []string
}

type SearchResults struct {
	PageNumber  int
	Books       []Book
	HasNextPage bool
}
