// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"bytes"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// fnPath builds a dot-separated path for nested fields.
func fnPath(base, add string) string {
	if base == "" {
		return add
	}
	return base + "." + add
}

// diffMessages recursively compares two protobuf messages and returns a slice of human‑readable differences.
func diffMessages(desired, actual protoreflect.Message, path string) []string {
	var diffs []string

	// Iterate over fields set in desired.
	desired.Range(func(fd protoreflect.FieldDescriptor, vDesired protoreflect.Value) bool {
		vActual := actual.Get(fd)
		fieldName := string(fd.Name())
		currPath := fnPath(path, fieldName)

		if fd.IsList() {
			if vDesired.List().Len() != vActual.List().Len() {
				diffs = append(diffs, fmt.Sprintf("%s len mismatch", currPath))
			}
			return true
		}

		if fd.IsMap() {
			mDesired := vDesired.Map()
			mActual := vActual.Map()
			if mDesired.Len() != mActual.Len() {
				diffs = append(diffs, fmt.Sprintf("%s count mismatch (%d vs %d)", currPath, mDesired.Len(), mActual.Len()))
			}
			mDesired.Range(func(mk protoreflect.MapKey, mvDesired protoreflect.Value) bool {
				if !mActual.Has(mk) {
					diffs = append(diffs, fmt.Sprintf("%s key missing: %v", currPath, mk))
					return true
				}
				mvActual := mActual.Get(mk)
				if fd.MapValue().Kind() == protoreflect.MessageKind {
					diffs = append(diffs, diffMessages(mvDesired.Message(), mvActual.Message(), fmt.Sprintf("%s[%v]", currPath, mk))...)
				} else {
					if mvDesired.Interface() != mvActual.Interface() {
						diffs = append(diffs, fmt.Sprintf("%s[%v] mismatch", currPath, mk))
					}
				}
				return true
			})
			return true
		}

		if fd.Kind() == protoreflect.MessageKind {
			// Recurse into nested
			diffs = append(diffs, diffMessages(vDesired.Message(), vActual.Message(), currPath)...)
			return true
		}

		// Bytes need special handling because they are slices (uncomparable).
		if fd.Kind() == protoreflect.BytesKind {
			if !bytes.Equal(vDesired.Bytes(), vActual.Bytes()) {
				diffs = append(diffs, fmt.Sprintf("%s bytes mismatch", currPath))
			}
			return true
		}

		// Basic scalar types.
		if vDesired.Interface() != vActual.Interface() {
			diffs = append(diffs, fmt.Sprintf("%s mismatch (Current: %v, Desired: %v)", currPath, vActual.Interface(), vDesired.Interface()))
		}
		return true
	})
	return diffs
}
