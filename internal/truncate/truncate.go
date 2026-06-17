// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package truncate is a byte-for-byte port of the kOps name-truncation helpers
// that cannot be expressed in pure HCL:
//
//   - HashString        (kOps pkg/truncate.HashString)
//   - SafeClusterName   (kOps upup/pkg/fi/cloudup/gce.SafeClusterName)
//   - LimitedLengthName (kOps upup/pkg/fi/cloudup/gce.LimitedLengthName)
//
// These are exercised by the kopsutils Terraform provider's data sources so
// that the gce-cluster module can compute kОps-equivalent resource names
// (instance-template name prefixes, service-account hashes) natively in HCL
// instead of relying on precomputed inputs.
//
// The implementations are verified against the golden tests/example1
// values (see truncate_test.go).
package truncate

import (
	"encoding/base32"
	"hash/fnv"
	"strings"
)

// HashString reproduces kОps pkg/truncate.HashString(s, length):
//
//	lower(base32hex(fnv32a(s)))[:length]
func HashString(s string, length int) string {
	h := fnv.New32a()
	// fnv hashing never returns an error from Write.
	_, _ = h.Write([]byte(s))
	out := strings.ToLower(base32.HexEncoding.EncodeToString(h.Sum(nil)))
	if length >= 0 && len(out) > length {
		out = out[:length]
	}
	return out
}

// SafeClusterName reproduces gce.SafeClusterName: "." -> "-".
func SafeClusterName(clusterName string) string {
	return strings.ReplaceAll(clusterName, ".", "-")
}

// LimitedLengthName reproduces gce.LimitedLengthName(s, n): the
// truncate-with-hash used for GCE instance-template name prefixes.
//
// If s already fits within n, it is returned unchanged. Otherwise a 6-char
// fnv32a/base32hex hash of s is appended and the base is truncated so the total
// length is at most n: s[:n-len(hash)-1] + "-" + hash.
func LimitedLengthName(s string, n int) string {
	if len(s) <= n {
		return s
	}

	hashString := HashString(s, 6)

	maxBaseLength := n - len(hashString) - 1
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
	}
	return s + "-" + hashString
}
