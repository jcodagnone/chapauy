// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// filename where SearchResultEntry objects are stored.
	notificationsFile = "documents.json"
)

// Combines multiple closers to ensure all resources are released.
type multiReadCloser struct {
	io.ReadCloser
	underlying io.Closer
}

// Implements io.Closer and ensures all resources are properly released.
func (r *multiReadCloser) Close() error {
	return errors.Join(
		r.ReadCloser.Close(),
		r.underlying.Close(),
	)
}

type FileStore struct {
	root  string
	dbRef *DbReference // Reference to use id2file conversion
}

// Creates a new file store instance. The provided path is the root
// directory where all database subdirectories will be created.
func NewFileStore(root string, dbRef *DbReference) *FileStore {
	return &FileStore{
		root:  filepath.Join(root, fmt.Sprintf("%02d", dbRef.ID)),
		dbRef: dbRef,
	}
}

// Ensures that the directory for the given database ID exists.
func (s *FileStore) dbDirMustExists() error {
	err := os.MkdirAll(s.root, 0o700)
	if err != nil {
		return fmt.Errorf("setting up file store: %w", err)
	}

	return nil
}

// Returns the full path to the notifications file.
func (s *FileStore) dbpath() string {
	return filepath.Join(s.root, notificationsFile)
}

// Reads and parses the entries from the database file.
func (s *FileStore) load(dbpath string) (map[string]SearchResultEntry, error) {
	// Load the ret map, if any.
	ret := make(map[string]SearchResultEntry)

	data, err := os.ReadFile(filepath.Clean(dbpath))
	if err != nil {
		// If the file does not exist, that's OK; we will create it.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading notifications file: %w", err)
		}
	} else if len(data) != 0 {
		if err = json.Unmarshal(data, &ret); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	return ret, nil
}

// Upsert loads the existing map of SearchResultEntry objects from notifications.json,
// inserts only the new entries, and returns the number of entries inserted.
func (s *FileStore) Upsert(entries []SearchResultEntry, dryRun bool) (int, error) {
	if err := s.dbDirMustExists(); err != nil {
		return 0, err
	}

	// Build the full path for the notifications file.
	entriesPath := s.dbpath()

	db, err := s.load(entriesPath)
	if err != nil {
		return 0, err
	}

	var n int

	for _, entry := range entries {
		// If the entry's key already exists, do nothing.
		if _, ok := db[entry.Href]; !ok {
			db[entry.Href] = entry
			n++
		}
	}

	if !dryRun {
		// Marshal the updated map into pretty-printed JSON.
		output, err := json.MarshalIndent(db, "", "  ")
		if err != nil {
			return 0, fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Write the output back to the file.
		if err = os.WriteFile(entriesPath, output, 0o600); err != nil {
			return 0, fmt.Errorf("failed to write notifications file: %w", err)
		}
	}

	return n, nil
}

// Converts a document ID to a filesystem path.
func (s *FileStore) pathFor(id string, createParent bool) (string, error) {
	if len(s.dbRef.id2file) == 0 {
		return "", fmt.Errorf("database %s doesn't support id2file conversion", s.dbRef.Name)
	}

	var path []string

	var err error

	// Try each extraction function until one succeeds
	for _, extractFunc := range s.dbRef.id2file {
		path, err = extractFunc(id)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	if len(path) == 0 {
		return "", fmt.Errorf("id2file returned an empty path for %q", id)
	}

	var ret string

	if len(path) == 1 {
		if createParent {
			if err := s.dbDirMustExists(); err != nil {
				return "", fmt.Errorf("creating parent directory: %w", err)
			}
		}

		ret = filepath.Join(
			s.root,
			path[0]+".html.gz",
		)
	} else {
		if createParent {
			parentDir := filepath.Join(s.root, filepath.Join(path[:len(path)-1]...))
			if err := os.MkdirAll(parentDir, 0o700); err != nil {
				return "", fmt.Errorf("creating parent directory: %w", err)
			}
		}

		last := len(path) - 1
		ret = filepath.Join(
			s.root,
			filepath.Join(path[:last]...),
			path[last]+".html.gz",
		)
	}

	return ret, nil
}

// Checks if a document exists in the file system.
func (s *FileStore) exists(url string) (bool, error) {
	path, err := s.pathFor(url, false)
	if err != nil {
		return false, fmt.Errorf("converting url to internal path: %s: %w", url, err)
	}

	_, err = os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// checkDocuments returns a list of document URLs based on their existence status.
// If wantExists is true, returns existing documents, otherwise returns missing documents.
func (s *FileStore) checkDocuments(wantExists bool) ([]string, error) {
	db, err := s.load(s.dbpath())
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0, len(db))

	for url := range db {
		exists, err := s.exists(url)
		if err != nil {
			return nil, err
		}

		// Add URL if existence matches what we want
		if exists == wantExists {
			ret = append(ret, url)
		}
	}

	return ret, nil
}

// Returns a list of document URLs that don't have local copies.
func (s *FileStore) MissingDocuments() ([]string, error) {
	return s.checkDocuments(false)
}

// Returns a list of document URLs that have local copies.
func (s *FileStore) ExistingDocuments() ([]string, error) {
	return s.checkDocuments(true)
}

// Stores a document of the specified type from an io.Reader.
// It compresses the content using gzip with best compression level.
func (s *FileStore) SaveDocument(id string, content io.Reader) error {
	path, err := s.pathFor(id, true)
	if err != nil {
		return fmt.Errorf("converting url to internal path: %s: %w", id, err)
	}

	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("creating html file: %w", err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", cerr))
		}
	}()

	gw, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}

	defer func() {
		if cerr := gw.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("closing gzip writer: %w", cerr))
		}
	}()

	if _, err := io.Copy(gw, content); err != nil {
		return fmt.Errorf("writing html file: %w", err)
	}

	return err
}

// GetDocument retrieves a document of the specified type as an io.ReadCloser.
func (s *FileStore) GetDocument(id string) (io.ReadCloser, error) {
	path, err := s.pathFor(id, false)
	if err != nil {
		return nil, fmt.Errorf("converting url to internal path: %s: %w", id, err)
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading html file: %w", err)
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		err1 := f.Close()

		return nil, errors.Join(fmt.Errorf("creating gzip reader: %w", err), err1)
	}

	return &multiReadCloser{gr, f}, nil
}
