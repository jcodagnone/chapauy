// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package httputils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

// dummyRoundTripper is useful to simulate a response.
type dummyRoundTripper struct {
	response *http.Response
}

func (d *dummyRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	if d.response != nil {
		return d.response, nil
	}

	return nil, nil
}

//////////////////////////////////
// Test LoggingRoundTripper

// TestLoggingRoundTripper verifies that the LoggingRoundTripper logs both the request and
// the response (including timing information).
func TestLoggingRoundTripper(t *testing.T) {
	// Buffer to capture log output.
	var logBuffer bytes.Buffer

	// Set up a dummy transport that returns a dummy response.
	drt := &dummyRoundTripper{
		response: &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("response body")),
		},
	}

	lt := &LoggingRoundTripper{
		Transport: drt,
		Writer:    &logBuffer,
		DumpBody:  true, // include body in the dump
	}

	// Create a basic request.
	req, err := http.NewRequest(http.MethodGet, "http://example.com/abc", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// RoundTrip through our logging round tripper.
	_, err = lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}

	// Check log contents.
	logContent := logBuffer.String()
	if !strings.Contains(logContent, "> GET /abc") {
		t.Errorf("log does not contain request info. Got: %s", logContent)
	}

	if !strings.Contains(logContent, "< RESPONSE: [") {
		t.Errorf("log does not contain response header with timing info. Got: %s", logContent)
	}

	if !strings.Contains(logContent, "response body") {
		t.Errorf("log does not contain response body. Got: %s", logContent)
	}
}

//////////////////////////////////
// Test AppendRequestHeadersRoundTripper

// dummyHeadersRoundTripper is used to verify that the headers are added.
type dummyHeadersRoundTripper struct {
	lastRequest *http.Request
}

func (d *dummyHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	d.lastRequest = req

	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestAppendRequestHeadersRoundTripper(t *testing.T) {
	// Create a dummy transport that captures the request.
	dummy := &dummyHeadersRoundTripper{}

	// Wrap it with AppendRequestHeadersRoundTripper.
	headersToAdd := map[string]string{
		"X-Test-Header": "TestValue",
	}
	atr := &AppendRequestHeadersRoundTripper{
		Transport: dummy,
		Headers:   headersToAdd,
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.org", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Ensure the header is not originally set.
	if req.Header.Get("X-Test-Header") != "" {
		t.Fatalf("the test header should not be pre-set in the request")
	}

	// Issue the request.
	_, err = atr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}

	// Verify that our header was added.
	if dummy.lastRequest == nil {
		t.Fatalf("dummy transport did not receive any request")
	}

	if got := dummy.lastRequest.Header.Get("X-Test-Header"); got != "TestValue" {
		t.Errorf("expected header X-Test-Header to have value 'TestValue', but got '%s'", got)
	}
}
