// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jcodagnone/chapauy/utils/htmlutils"
	"golang.org/x/net/html"
)

// A single search result from the IMPO database.
type SearchResultEntry struct {
	Title    string `json:"title"`    // Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 1/025
	Href     string `json:"href"`     // https://impo.com.uy/bases/notificaciones-transito-maldonado/1-2025
	Subtitle string `json:"subtitle"` // NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO
}

func (e *SearchResultEntry) Validate() error {
	ret := []string{}

	if e.Title == "" {
		ret = append(ret, "Title is empty")
	}

	if e.Href == "" {
		ret = append(ret, "Href is empty")
	}

	if ret == nil {
		return nil
	}

	return errors.New(strings.Join(ret, "; "))
}

// Search results and pagination information.
type SearchResults struct {
	Entries []SearchResultEntry `json:"entries"` // entries found
	Next    string              `json:"next"`    // next page information
}

// Tracks statistics during the search phase.
type SearchMetrics struct {
	SearchPages        int // number of pages traversed
	SearchTotalRecords int // number of records discovered
	SearchTotalStored  int // number of records new to the database
}

// Combines two SearchMetrics objects.
func (f *SearchMetrics) Merge(o *SearchMetrics) *SearchMetrics {
	f.SearchPages += o.SearchPages
	f.SearchTotalRecords += o.SearchTotalRecords
	f.SearchTotalStored += o.SearchTotalStored

	return f
}

// Processes the HTML node and extracts search results.
func parseSearches(n *html.Node) (*SearchResults, error) {
	ret := SearchResults{
		Entries: make([]SearchResultEntry, 0, 50),
	}
	err := visitSearch(&ret, n)

	return &ret, err
}

// signIn ensures we have the necessary cookies to access the database.
func (c *Client) signIn() error {
	parsedURL, err := url.Parse(c.dbRef.QueryURL)
	if err != nil {
		return fmt.Errorf("parsing query URL: %w", err)
	}

	// Before signin we search if we have a cookie for the db
	hasCookie := false

	cookieName := fmt.Sprintf("usrts_%d", c.dbRef.ID)
	for _, c := range c.client.Jar.Cookies(parsedURL) {
		if c.Name == cookieName {
			hasCookie = true

			break
		}
	}

	if hasCookie {
		return nil
	}

	// The anonymous login sequence consists of some redirects
	ctx := context.Background()
	ctx = context.WithValue(ctx, allowRedirectKey, true)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.dbRef.SeedURL, nil)
	if err != nil {
		return fmt.Errorf("creating sign-in request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("executing sign-in request: %w", err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("closing resp.Body: %w", cerr))
		}
	}()

	// We don't care much about the status code, we care about
	// the cookie being set
	if resp.StatusCode > 399 {
		return fmt.Errorf("signing in to %s: status %d", c.dbRef.SeedURL, resp.StatusCode)
	}

	return err
}

// fetches a single page of search results from the IMPO database.
func (c *Client) retrieveSearchPage(page string) (*SearchResults, error) {
	if c.dbRef.SeedURL == "" {
		return nil, errors.New("db entry - seed url is missing")
	}

	if c.dbRef.QueryURL == "" {
		return nil, errors.New("db entry - query url is missing")
	}

	if c.dbRef.BaseURL == "" {
		return nil, errors.New("db entry - base url is missing")
	}

	impoBaseURL, err := url.Parse(c.dbRef.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL <%s>: %w", c.dbRef.BaseURL, err)
	}

	if err := c.signIn(); err != nil {
		return nil, fmt.Errorf("signing in to database %d: %w", c.dbRef.ID, err)
	}

	var resp *http.Response

	if page == "" {
		// First page request
		log.Printf("Search - Retrieving first page <%s>", c.dbRef.QueryURL)
		resp, err = c.client.PostForm(
			c.dbRef.QueryURL,
			url.Values{
				"realizarconsulta":       {"SI"},
				"nuevaconsulta":          {"SI"},
				"parlistabases":          {""},
				"tipoServicio":           {strconv.Itoa(c.dbRef.ID)},
				"combo1":                 {strconv.Itoa(c.dbRef.TodosID)},
				"numeros":                {""},
				"articulos":              {""},
				"textolibre":             {""},
				"texto1":                 {""},
				"campotexto1":            {"TODOS"},
				"optexto1":               {"Y"},
				"texto2":                 {""},
				"campotexto2":            {"TODOS"},
				"optexto2":               {"Y"},
				"texto3":                 {""},
				"campotexto3":            {"TODOS"},
				"fechadiar1":             {""},
				"fechadiar2":             {""},
				"fechapro1":              {""},
				"fechapro2":              {""},
				"indexcombobasetematica": {"-1"},
				"tema":                   {""},
				"ntema":                  {""},
				"refinar":                {""},
			},
		)
	} else {
		// Subsequent page request
		var parsedURL *url.URL

		parsedURL, err = url.Parse(c.dbRef.QueryURL)
		if err != nil {
			return nil, fmt.Errorf("parsing QueryUrl <%s>: %w", c.dbRef.QueryURL, err)
		}

		log.Printf("Search - Retrieving next page %s", page)
		parsedURL.RawQuery = page
		resp, err = c.client.Get(parsedURL.String())
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("closing resp.Body: %w", cerr))
		}
	}()

	r, err := htmlutils.AsReader(resp)
	if err != nil {
		return nil, fmt.Errorf("converting response to reader: %w", err)
	}

	node, err := htmlutils.AsNode(r)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	response, err := parseSearches(node)
	if err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}

	// Make sure that all the references are in fact URLs
	for i := range response.Entries {
		href, err := url.Parse(response.Entries[i].Href)
		if err != nil {
			return nil, fmt.Errorf("parsing href %s: %w", response.Entries[i].Href, err)
		}

		response.Entries[i].Href = impoBaseURL.ResolveReference(href).String()
	}

	return response, err
}

// searchForNewDocuments performs the search phase by traversing pages and finding new documents.
func (c *Client) searchForNewDocuments() error {
	page := ""

	for range c.options.SearchDepth {
		metrics := SearchMetrics{}
		metrics.SearchPages++

		r, err := c.retrieveSearchPage(page)
		if err != nil {
			return fmt.Errorf("retrieving search page: %w", err)
		}

		metrics.SearchTotalRecords += len(r.Entries)

		storedCount, err := c.store.Upsert(r.Entries, c.options.DryRun)
		if err != nil {
			return fmt.Errorf("storing search results: %w", err)
		}

		metrics.SearchTotalStored = storedCount

		log.Printf(
			"Search - Page %d stats - %d new records from a total of %d records",
			metrics.SearchPages,
			metrics.SearchTotalStored,
			metrics.SearchTotalRecords,
		)

		c.Metrics.SearchMetrics.Merge(&metrics)

		page = r.Next

		// Stop conditions
		if (metrics.SearchTotalStored == 0 && !c.options.SearchFull) || strings.TrimSpace(page) == "" {
			break
		}
	}

	return nil
}

// Extracts search entries from a table.
func visitSearchTable(entries *[]SearchResultEntry, child *html.Node) error {
	sb := strings.Builder{}
	nr := 0

	for child := child.FirstChild; child != nil; child = child.NextSibling {
		// We're interested in <tr> elements
		if child.Type == html.ElementNode && strings.EqualFold("tr", child.Data) {
			record := SearchResultEntry{}

			for tdChild := child.FirstChild; tdChild != nil; tdChild = tdChild.NextSibling {
				if tdChild.Type == html.ElementNode && strings.EqualFold("td", tdChild.Data) {
					for tdContent := tdChild.FirstChild; tdContent != nil; tdContent = tdContent.NextSibling {
						if tdContent.Type == html.ElementNode && strings.EqualFold("a", tdContent.Data) {
							// Extract href attribute
							for _, attr := range tdContent.Attr {
								if strings.EqualFold("href", attr.Key) {
									record.Href = attr.Val
								}
							}

							// Extract title from <strong> element
							for aContent := tdContent.FirstChild; aContent != nil; aContent = aContent.NextSibling {
								if aContent.Type == html.ElementNode && strings.EqualFold("strong", aContent.Data) {
									sb.Reset()

									err := htmlutils.Node2string(aContent, &sb)
									if err != nil {
										return fmt.Errorf("extracting title for record %d: %w", nr, err)
									}

									record.Title = sb.String()
								}
							}
						} else if tdContent.Type == html.ElementNode && strings.EqualFold("font", tdContent.Data) {
							// Extract subtitle
							sb.Reset()

							err := htmlutils.Node2string(tdContent, &sb)
							if err != nil {
								return fmt.Errorf("extracting subtitle for record %d: %w", nr, err)
							}

							record.Subtitle = sb.String()
						}
					}
				}
			}

			*entries = append(*entries, record)
			nr++
		}
	}

	return nil
}

// Traverses the HTML document looking for search results and pagination.
func visitSearch(r *SearchResults, n *html.Node) error {
	// Look for table with id="resultadoConsulta"
	var isTable bool

	if n.Type == html.ElementNode && strings.EqualFold("tbody", n.Data) {
		for _, attr := range n.Attr {
			isTable = isTable || (strings.EqualFold("id", attr.Key) && attr.Val == "resultadoConsulta")
		}
	} else if n.Type == html.ElementNode && strings.EqualFold("a", n.Data) {
		// Look for next page link
		var href string

		var nextPage bool

		for _, attr := range n.Attr {
			if strings.EqualFold("class", attr.Key) && attr.Val == "nextPage" {
				nextPage = true
			} else if strings.EqualFold("href", attr.Key) {
				href = attr.Val
			}
		}

		if href != "" && nextPage {
			parsedURL, err := url.Parse(href)
			if err != nil {
				return fmt.Errorf("parsing next page URL: %w", err)
			}

			r.Next = parsedURL.RawQuery
		}
	}

	if isTable {
		return visitSearchTable(&r.Entries, n)
	}

	// Continue traversing
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if err := visitSearch(r, child); err != nil {
			return err
		}
	}

	return nil
}
