// Copyright 2026 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestClient_downloadMissing_NotYetPublished(t *testing.T) {
	// Create a mock server
	htmlContent := `<html><title>Test</title><body><h4>Este contenido se publicará en la edición del Diario Oficial del día 23/02/2026</h4></body></html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(htmlContent))
	}))
	defer ts.Close()

	// Create temp directory for db
	tempDir, err := os.MkdirTemp("", "impo-test")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbRef := &DbReference{
		ID:   48,
		Name: "Test DB",
		id2file: []func(string) ([]string, error){
			func(_ string) ([]string, error) {
				return []string{"test_doc"}, nil
			},
		},
	}

	opts := &ClientOptions{
		DbPath: tempDir,
	}

	client := NewImpoClient(opts, dbRef, nil)

	store := client.store

	entries := map[string]SearchResultEntry{
		ts.URL: {Href: ts.URL},
	}

	err = store.dbDirMustExists()
	if err != nil {
		t.Fatalf("dbDirMustExists failed: %v", err)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	err = os.WriteFile(store.dbpath(), data, 0o644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	client.options.DryRun = false

	err = client.downloadMissing()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if client.Metrics.DownloadsErr != 1 {
		t.Errorf("Expected 1 download error, got %d", client.Metrics.DownloadsErr)
	}

	if client.Metrics.DownloadsOk != 0 {
		t.Errorf("Expected 0 downloads ok, got %d", client.Metrics.DownloadsOk)
	}
}
