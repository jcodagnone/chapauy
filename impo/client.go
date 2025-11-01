// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"slices"
	"time"

	"github.com/jcodagnone/chapauy/utils/htmlutils"
	"github.com/jcodagnone/chapauy/utils/httputils"
)

// Common errors returned by the client.
var (
	ErrRedirectNotAllowed = errors.New("redirect not allowed")
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// Context keys used by the package.
const (
	allowRedirectKey contextKey = "allowRedirect"
)

// ClientOptions configuration for ImpoClient.
type ClientOptions struct {
	// DbPath is the root path for the database
	DbPath string

	// UserAgent is the User-Agent header to use in HTTP requests
	UserAgent string

	// Enables light tracing of HTTP requests and responses
	EnableHTTPTrace bool

	// Enables full HTTP body tracing
	EnableHTTPBodyTrace bool

	// Skips the search phase (discovering new documents)
	SkipSearch bool

	// Overrides incremental search and traverses all pages
	SearchFull bool

	// Skips the download phase (downloading known missing documents)
	SkipDownload bool

	// Skips the extraction phase (extracting information from available documents)
	SkipExtract bool

	// Overrides incremental extract and traverses all pages
	ExtractFull bool

	// Avoid storing documents with errors
	SkipErrDocs bool

	// Maximum number of pages to traverse during search phase
	SearchDepth int

	// Dry run, don't persist any change
	DryRun bool

	// Max number of processes to use in the extraction phase.
	ExtractMaxProcs int
}

// ClientMetrics tracks various metrics collected during client operations.
type ClientMetrics struct {
	SearchMetrics
	DownloadMetrics
	ExtractMetrics
}

// Merge combines the metrics from another ClientMetrics instance into this one.
func (m *ClientMetrics) Merge(other *ClientMetrics) *ClientMetrics {
	if other == nil {
		return m
	}

	m.SearchMetrics.Merge(&other.SearchMetrics)
	m.DownloadMetrics.Merge(&other.DownloadMetrics)
	m.ExtractMetrics.Merge(&other.ExtractMetrics)

	return m
}

// "Consultar bases de infracciones y multas de tránsito publicadas en el Diario Oficial".
type Client struct {
	dbRef   *DbReference
	client  *http.Client
	options *ClientOptions
	store   *FileStore
	repo    OffenseRepository
	Metrics ClientMetrics
}

// NewImpoClient creates a new client with the provided options and database reference.
func NewImpoClient(options *ClientOptions, dbRef *DbReference, repo OffenseRepository) *Client {
	if options == nil {
		options = &ClientOptions{}
	}

	var httpLogWriter io.Writer
	if options.EnableHTTPTrace {
		httpLogWriter = os.Stderr
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %v", err)
	}

	// The cookies IMPO sends don't have an expiration, but
	// they actually expire within 30 minutes
	cookieJar := &httputils.EnforceExpirationCookieJar{
		Target:   jar,
		Duration: 10 * time.Minute,
	}

	transport := &http.Transport{
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   4,
		MaxConnsPerHost:       4,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}

	loggingTransport := &httputils.LoggingRoundTripper{
		Writer:    httpLogWriter,
		DumpBody:  options.EnableHTTPBodyTrace,
		Transport: transport,
	}

	userAgent := "chapauy/unknown"
	if options.UserAgent != "" {
		userAgent = options.UserAgent
	}

	headerTransport := &httputils.AppendRequestHeadersRoundTripper{
		Headers: map[string]string{
			"User-Agent": userAgent,
			"Accept":     "*/*",
		},
		Transport: loggingTransport,
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			// For this kind of program, no redirects is the policy, but in this case
			// we need to allow redirects during login to get a valid session
			v := req.Context().Value(allowRedirectKey)
			allowRedirect, ok := v.(bool)
			if ok && allowRedirect {
				return nil
			}

			return http.ErrUseLastResponse
		},
		Jar:       cookieJar,
		Transport: headerTransport,
	}

	return &Client{
		dbRef:   dbRef,
		client:  client,
		store:   NewFileStore(options.DbPath, dbRef),
		repo:    repo,
		options: options,
	}
}

// DownloadMetrics tracks statistics about the download process.
type DownloadMetrics struct {
	DownloadsOk  int
	DownloadsErr int
}

// Merge combines two DownloadMetrics.
func (f *DownloadMetrics) Merge(o *DownloadMetrics) *DownloadMetrics {
	f.DownloadsOk += o.DownloadsOk
	f.DownloadsErr += o.DownloadsErr

	return f
}

// Downloads missing HTML documents.
func (c *Client) downloadMissing() error {
	missing, err := c.store.MissingDocuments()
	if err != nil {
		return fmt.Errorf("getting missing documents: %w", err)
	}

	if len(missing) == 0 {
		log.Println("Nothing to download")
	}

	slices.Sort(missing)
	n := len(missing)

	var errs []error

	for i, id := range missing {
		log.Printf("[%d/%d] Downloading %s", i+1, n, id)

		resp, err := c.client.Get(id)
		if err != nil {
			c.Metrics.DownloadsErr++

			errs = append(errs, err)
			log.Printf("[%d/%d] Download failed: %s", i+1, n, err)

			continue
		}

		r, err := htmlutils.AsReader(resp)
		if err != nil {
			errs = append(
				errs,
				errors.Join(
					resp.Body.Close(),
					fmt.Errorf("reading response body: %w", err),
				),
			)

			log.Printf("[%d/%d] Parsing: %s", i+1, n, err)

			continue
		}

		if !c.options.DryRun {
			if err := c.store.SaveDocument(id, r); err != nil {
				errs = append(
					errs,
					errors.Join(
						resp.Body.Close(),
						fmt.Errorf("saving document: %q %w", id, err),
					),
				)

				log.Printf("[%d/%d] Saving document: %s", i+1, n, err)
			}
		}

		if err := resp.Body.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing request: %q %w", id, err))
			log.Printf("[%d/%d] Closing response: %s", i+1, n, err)
		}

		c.Metrics.DownloadsOk++
	}

	c.Metrics.DownloadsErr += len(errs)
	if c.Metrics.DownloadsOk != 0 || c.Metrics.DownloadsErr != 0 {
		log.Printf(
			"Download phase completed - %d successful, %d failed",
			c.Metrics.DownloadsOk,
			c.Metrics.DownloadsErr,
		)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// 3. Extract: Parse downloaded documents to extract relevant information.
func (c *Client) Update() error {
	log.Printf("Updating database %d - %s", c.dbRef.ID, c.dbRef.Name)

	if !c.options.SkipSearch {
		if err := c.searchForNewDocuments(); err != nil {
			return fmt.Errorf("searching for new documents: %w", err)
		}

		log.Printf(
			"Total stats - %d new records from a total of %d records across %d pages",
			c.Metrics.SearchTotalStored,
			c.Metrics.SearchTotalRecords,
			c.Metrics.SearchPages,
		)
	}

	if c.options.SkipDownload {
		log.Println("Skipping download phase")
	} else {
		if err := c.downloadMissing(); err != nil {
			return err
		}
	}

	if c.options.SkipExtract {
		log.Println("Skipping extraction phase")
	} else {
		if err := c.extractDocuments(); err != nil {
			return err
		}
	}

	return nil
}
