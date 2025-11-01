// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

// Package htmlutils provides utility functions for working with HTML.
package htmlutils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const htmlAllowedRunes = "(?i)[\u0020-\u007Fáéíóöúñº°ªq]*"

var (
	htmlAllowedRunesRegex     = regexp.MustCompile("^" + htmlAllowedRunes + "$")
	htmlAllowedRunesShowRegex = regexp.MustCompile("(?i)" + htmlAllowedRunes)
)

// to be able to get correct records.
func Node2string(n *html.Node, sb *strings.Builder) (err error) {
	if n.Type == html.TextNode {
		tmp := strings.TrimSpace(n.Data)

		// original documents are coded with ISO-8859-1, so if we
		// find a  REPLACEMENT CHARACTER (U+FFFD) it means that we
		// are reading it in the incorrect charset
		if idx := strings.IndexRune(tmp, utf8.RuneError); idx != -1 {
			err = fmt.Errorf("charset missmatch found: `%s'", tmp)
		}

		for _, r := range [][]string{
			{"Ã³", "ó"},
			{"Ã¡", "á"},
			{"Ãƒ?", "Ñ"},
			{"mÃnimo", "mínimo"},
			{"E´", "É"},
			{"X\nPERS", "XPERS"}, // moto URs for Colonia
			{"\n", " "},
		} {
			// somehow some lines are bad coded
			tmp = strings.ReplaceAll(tmp, r[0], r[1])
		}

		if !htmlAllowedRunesRegex.MatchString(tmp) {
			a := htmlAllowedRunesShowRegex.FindStringIndex(tmp)[1]
			err = fmt.Errorf("invalid character: %+q in %q", tmp[a], tmp)
		}

		if err == nil && len(tmp) > 0 {
			if sb.Len() != 0 {
				sb.WriteByte(' ')
			}

			sb.WriteString(tmp)
		}
	} else {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			err = Node2string(child, sb)
			if err != nil {
				break
			}
		}
	}

	return err
}

// Validates that response seems to be an HTML response.
func hasHTMLContentType(media string) bool {
	const expectedMedia = "text/html"

	return strings.EqualFold(
		expectedMedia,
		media[0:min(len(media), len(expectedMedia))],
	)
}

// AsReader converts an HTTP response body to an io.Reader with the correct charset.
func AsReader(resp *http.Response) (io.Reader, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	media := resp.Header.Get("Content-Type")
	if !hasHTMLContentType(media) {
		return nil, fmt.Errorf("media type is %s", media)
	}

	rr, err := charset.NewReader(resp.Body, media)
	if err != nil {
		return nil, err
	}

	return rr, nil
}

// AsNode parses an io.Reader as an HTML node.
func AsNode(r io.Reader) (*html.Node, error) {
	n, err := html.Parse(r)
	if nil != err {
		return nil, fmt.Errorf("parsing body as HTML: %w", err)
	}

	if err := failIfLogin(n); err != nil {
		return nil, err
	}

	return n, nil
}

// ErrSessionExpired is returned when the session has expired.
var ErrSessionExpired = errors.New("session expired")

func failIfLogin(n *html.Node) (err error) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && strings.EqualFold("title", child.Data) {
			sb := strings.Builder{}

			err = Node2string(child, &sb)
			if err != nil {
				break
			}

			if sb.String() == "Ingreso - IMPO" {
				err = ErrSessionExpired

				break
			}
		} else if child.Type == html.ElementNode && strings.EqualFold("body", child.Data) {
			// we're done
			break
		} else {
			err = failIfLogin(child)
			if err != nil {
				break
			}
		}
	}

	return err
}
