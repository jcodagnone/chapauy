// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package htmlutils

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func asHTMLNode(resp *http.Response) (*html.Node, error) {
	r, err := AsReader(resp)
	if err != nil {
		return nil, err
	}

	return AsNode(r)
}

func TestNode2string(t *testing.T) {
	tests := []struct {
		fail     bool
		expected string
		input    string
	}{
		{false, "foo bar", "<div><pre>foo</pre><span>bar</span>"},
		{true, "", "<span>a\uFFFDo</span>"},
	}

	for _, test := range tests {
		n, err := html.Parse(strings.NewReader(test.input))
		if err != nil {
			t.Fatalf("parsing HTML `%s': %s", test.input, err)
		}

		sb := strings.Builder{}

		err = Node2string(n, &sb)
		if !test.fail && err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if test.fail && err == nil {
			t.Errorf("didn't fail: %s", test.input)
		}

		if got := sb.String(); got != test.expected {
			t.Errorf("`%s': expected `%v' but got `%v'", test.input, test.expected, got)
		}
	}
}

func TestAsHTMLReader_WithNonOKStatus(t *testing.T) {
	const msg = "status 404"

	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	r, err := asHTMLNode(resp)
	if r != nil {
		t.Errorf("Expected nil reader")
	} else if err == nil || !strings.Contains(err.Error(), msg) {
		t.Errorf("Expected error containing '%s', got %v", msg, err)
	}
}

func TestAsHTMLReader_WithWrongMediaType(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("plain text")),
	}
	resp.Header.Set("Content-Type", "text/plain")

	r, err := asHTMLNode(resp)
	if r != nil {
		t.Errorf("Expected nil reader")
	} else if err == nil || !strings.Contains(err.Error(), "text/plain") {
		t.Errorf("Expected error mentioning media type, got %v", err)
	}
}

func TestAsHTMLReader_HappyPathTranscoding(t *testing.T) {
	htmlData := "<html>hola</html>"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(htmlData)),
	}
	// Include charset information to test that the reader is correctly created.
	resp.Header.Set("Content-Type", "text/html; charset=iso-8859-1")

	reader, err := asHTMLNode(resp)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	sb := strings.Builder{}
	if err = Node2string(reader, &sb); err != nil {
		t.Fatal(err)
	}

	if sb.String() != "hola" {
		t.Errorf("Expected content")
	}
}

func TestHasHtmlContentType(t *testing.T) {
	tests := []struct {
		expected bool
		input    string
	}{
		{false, ""},
		{false, "text/plain"},
		{true, "text/html"},
		{true, "text/HTml"},
		{true, "text/html; charset=ISO-8859-1"},
	}

	for _, test := range tests {
		if got := hasHTMLContentType(test.input); got != test.expected {
			t.Errorf("`%s': expected %v but got %v", test.input, test.expected, got)
		}
	}
}

func TestFailIfLogin_Fails(t *testing.T) {
	htmlData := `<!DOCTYPE html>
<html lang="es">
  <head>
    <title>Ingreso - IMPO</title>
  </head>
</html>`

	n, err := html.Parse(strings.NewReader(htmlData))
	if nil != err {
		t.Error(err)
	}

	err = failIfLogin(n)
	if err == nil {
		t.Fatal("was expecting an error")
	} else if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("expect %s got %s", ErrSessionExpired, err)
	}
}

func TestFailIfLogin_Nil(t *testing.T) {
	htmlData := `<!DOCTYPE html>
<html lang="es">
  <head>
    <title>Foo bar</title>
  </head>
</html>`

	n, err := html.Parse(strings.NewReader(htmlData))
	if nil != err {
		t.Error(err)
	}

	err = failIfLogin(n)
	if err != nil {
		t.Fatal("wasn't  expecting an error")
	}
}
