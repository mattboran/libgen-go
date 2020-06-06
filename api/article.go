package api

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ArticleSearchInput contains the fields required to search
// Library Genesis' article endpoint.
type ArticleSearchInput struct {
	Query []string
	Page  int
}

type article struct {
	authors  []string
	title    string
	journal  string
	fileSize string
	mirrors  []string
}

type articleResultParser struct {
	articles *[]article
	page     int
}

// Name is the displayable name for a Downloadable article
func (a article) Name() string {
	authors := strings.Join(a.authors, ", ")
	return fmt.Sprintf("%s (%s) by %s", a.title, a.journal, authors)
}

// Mirrors returns the list of mirrors available for a given article
func (a article) Mirrors() []string {
	return a.mirrors
}

// ShortName provides a default filename for use in downloading
func (a article) Filename() string {
	title := strings.ReplaceAll(a.title, " ", "_")
	return fmt.Sprintf("%s.pdf", title)
}

// CurrentPage returns the selected page number for the given search input
func (input ArticleSearchInput) CurrentPage() int {
	return input.Page
}

// NextPage returns a copy of FictionSearchInput but with Page incremented
func (input ArticleSearchInput) NextPage() SearchInput {
	return ArticleSearchInput{
		Query: input.Query,
		Page:  input.Page + 1,
	}
}

// PreviousPage returns a copy of FictionSearchInput but with Page decremented
func (input ArticleSearchInput) PreviousPage() SearchInput {
	return ArticleSearchInput{
		Query: input.Query,
		Page:  input.Page - 1,
	}
}

func (input ArticleSearchInput) url() (*url.URL, error) {
	params := url.Values{}

	params.Add("q", strings.Join(input.Query, " "))
	params.Add("page", strconv.Itoa(input.Page))

	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return nil, err
	}

	baseURL.Path += "scimag/"
	baseURL.RawQuery = params.Encode()
	return baseURL, nil
}

func (input ArticleSearchInput) resultParser() resultParser {
	return &articleResultParser{
		articles: &[]article{},
		page:     input.Page,
	}
}

func (parser articleResultParser) currentPage() int {
	return parser.page
}

func (parser articleResultParser) parsedResults() []DownloadableResult {
	result := []DownloadableResult{}
	for _, article := range *parser.articles {
		result = append(result, article)
	}
	return result
}

func (parser articleResultParser) hasNextPage() bool {
	return (len(*parser.articles) % 25) == 0
}

func (parser articleResultParser) parseResultsFromTableRows() func(int, *goquery.Selection) {

	return func(i int, sel *goquery.Selection) {
		if i == 0 {
			return
		}
		var authors, mirrors []string
		var title, journal, fileSize string
		sel.Find("td").Each(func(j int, col *goquery.Selection) {
			switch j {
			case 0:
				text := trim(col.Text())
				authors = strings.Split(text, ";")
			case 1:
				titleText := col.Find("a").Text()
				title = trim(titleText)
			case 2:
				journal = trim(col.Text())
			case 3:
				fileSizeText := col.First().Text()
				fileSize = trim(fileSizeText)
			case 4:
				col.Find("a[href]").Each(func(k int, item *goquery.Selection) {
					href, found := item.Attr("href")
					if found {
						mirrors = append(mirrors, href)
					}
				})
			}
		})

		*parser.articles = append(*parser.articles, article{
			authors:  authors,
			title:    title,
			journal:  journal,
			fileSize: fileSize,
			mirrors:  mirrors,
		})
	}
}
