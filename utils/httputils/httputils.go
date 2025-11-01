// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

// Package httputils provides utility functions for working with HTTP.
package httputils

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

/////////////////////////////////////////
/// RountTrippers

// LoggingRoundTripper adds a very primitive logging to a http transaction.
type LoggingRoundTripper struct {
	Transport http.RoundTripper
	Writer    io.Writer
	DumpBody  bool
}

// reduce the content the liens.
func abbreviate(lines []string, prefix rune) []string {
	const maxLines, maxChars = 2048, 512

	for i, line := range lines {
		if i < maxLines {
			// TODO(juan) trim Authorization header
			lines[i] = fmt.Sprintf("%c %s", prefix, line)
		} else {
			break
		}
	}

	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "…")
	}

	for i, line := range lines {
		if len(line) > maxChars {
			lines[i] = line[0:maxChars] + "…"
		}
	}

	return lines
}

func (t *LoggingRoundTripper) dumpRequest(req *http.Request) error {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Errorf("tracing HTTP request: %w", err)
	}

	lines := abbreviate(strings.Split(string(dump), "\n"), '>')
	lines = append(lines, "")
	_, err = fmt.Fprint(t.Writer, strings.Join(lines, "\n"))

	return err
}

func (t *LoggingRoundTripper) dumpResponse(resp *http.Response, duration time.Duration) error {
	dump, err := httputil.DumpResponse(resp, t.DumpBody)
	if err != nil {
		return fmt.Errorf("tracing HTTP request: %w", err)
	}

	lines := abbreviate(strings.Split(string(dump), "\n"), '<')

	_, err = fmt.Fprintf(t.Writer, "< RESPONSE: [%v]\n", duration)
	if err != nil {
		return fmt.Errorf("tracing HTTP request: %w", err)
	}

	lines = append(lines, "")
	_, err = fmt.Fprint(t.Writer, strings.Join(lines, "\n"))

	return err
}

// RoundTrip implements the http.RoundTripper interface.
func (t *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Writer == nil {
		return t.Transport.RoundTrip(req)
	}

	if err := t.dumpRequest(req); err != nil {
		return nil, err
	}

	start := time.Now()

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if err := t.dumpResponse(resp, time.Since(start)); err != nil {
		return nil, err
	}

	return resp, nil
}

// AppendRequestHeadersRoundTripper adds headers to the request.
type AppendRequestHeadersRoundTripper struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

// RoundTrip implements the http.RoundTripper interface.
func (t *AppendRequestHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}

	resp, err := t.Transport.RoundTrip(req)

	return resp, err
}

////////////////////////////////////////////////////

// implementation, but enforce expirations dates if missing.
type EnforceExpirationCookieJar struct {
	Target   *cookiejar.Jar
	Duration time.Duration
}

// SetCookies sets the cookies.
func (t *EnforceExpirationCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	now := time.Now()

	for _, cookie := range cookies {
		if cookie.Expires.IsZero() {
			cookie.Expires = now.Add(t.Duration)
		}
	}

	(*t.Target).SetCookies(u, cookies)
}

// Cookies returns the cookies.
func (t *EnforceExpirationCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return (*t.Target).Cookies(u)
}
