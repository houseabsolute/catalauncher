// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

// EmptySHA defines empty git SHA
const EmptySHA = "0000000000000000000000000000000000000000"

// SHA1 a git commit name
type SHA1 [20]byte

// Equal returns true if s has the same SHA1 as caller.
// Support 40-length-string, []byte, SHA1.
func (id SHA1) Equal(s2 interface{}) bool {
	switch v := s2.(type) {
	case string:
		if len(v) != 40 {
			return false
		}
		return v == id.String()
	case []byte:
		return bytes.Equal(v, id[:])
	case SHA1:
		return v == id
	default:
		return false
	}
}

// String returns string (hex) representation of the Oid.
func (id SHA1) String() string {
	return hex.EncodeToString(id[:])
}

// MustID always creates a new SHA1 from a [20]byte array with no validation of input.
func MustID(b []byte) SHA1 {
	var id SHA1
	copy(id[:], b)
	return id
}

// NewID creates a new SHA1 from a [20]byte array.
func NewID(b []byte) (SHA1, error) {
	if len(b) != 20 {
		return SHA1{}, fmt.Errorf("Length must be 20: %v", b)
	}
	return MustID(b), nil
}

// MustIDFromString always creates a new sha from a ID with no validation of input.
func MustIDFromString(s string) SHA1 {
	b, _ := hex.DecodeString(s)
	return MustID(b)
}

// NewIDFromString creates a new SHA1 from a ID string of length 40.
func NewIDFromString(s string) (SHA1, error) {
	var id SHA1
	s = strings.TrimSpace(s)
	if len(s) != 40 {
		return id, fmt.Errorf("Length must be 40: %s", s)
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}
	return NewID(b)
}
