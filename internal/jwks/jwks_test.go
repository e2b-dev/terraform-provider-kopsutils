// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

package jwks

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
)

// saPublicKeyPEM is the deterministic test cluster's service-account public key
// (tests/example4/shared/golden-pki.json), in the exact form kОps writes it: a
// PKIX body under an "RSA PUBLIC KEY" PEM header.
const saPublicKeyPEM = `-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuLFG2byony9UgUzcsJUl
il70lcDoEeTbRxmjBtlbBHacepKWqNDVnfjcgOLAIiekFxyBnhQ/LxuoUVXSwy1H
30TWz5RwSC0ZUf6eZAlqgwjV00y9nMfYSkH0IxEKsyBMJamSPbogNDeT6QOGe+t9
i/4yo662LP2KcbRwEpK5Sah033ZDuv++xI4gOnFovxMrM7ELz+37N59MZ3jP4Dbz
ODXLWhdwiVPY0GYj74FrLFChA98Ov6CvVyTwqQsSl7e2No2z0z/E54E9mycwHK0u
wmPEIKj2NZN47b+I8eCqWKE0wUEo/ZYlbL0ysguCQdP1iocC+TqEJoKu1WwJY97o
RQIDAQAB
-----END RSA PUBLIC KEY-----
`

// Expected values for saPublicKeyPEM, derived the same way kОps does
// (kops/pkg/model/issuerdiscovery.go) and pinned here as a regression anchor.
const (
	wantKid = "pj8NeK48bBnWNDUdRiplGLUs36kx1h6yD3afS8O1wis"
	wantN   = "uLFG2byony9UgUzcsJUlil70lcDoEeTbRxmjBtlbBHacepKWqNDVnfjcgOLAIiekFxyBnhQ_LxuoUVXSwy1H30TWz5RwSC0ZUf6eZAlqgwjV00y9nMfYSkH0IxEKsyBMJamSPbogNDeT6QOGe-t9i_4yo662LP2KcbRwEpK5Sah033ZDuv--xI4gOnFovxMrM7ELz-37N59MZ3jP4DbzODXLWhdwiVPY0GYj74FrLFChA98Ov6CvVyTwqQsSl7e2No2z0z_E54E9mycwHK0uwmPEIKj2NZN47b-I8eCqWKE0wUEo_ZYlbL0ysguCQdP1iocC-TqEJoKu1WwJY97oRQ"
	wantE   = "AQAB"
)

// wantJSON is the full JWKS document, byte-identical to kОps' OIDCKeys.Open
// output (json.MarshalIndent with empty prefix and indent) and to what the API
// server serves at /openid/v1/jwks.
const wantJSON = `{
"keys": [
{
"use": "sig",
"kty": "RSA",
"kid": "pj8NeK48bBnWNDUdRiplGLUs36kx1h6yD3afS8O1wis",
"alg": "RS256",
"n": "uLFG2byony9UgUzcsJUlil70lcDoEeTbRxmjBtlbBHacepKWqNDVnfjcgOLAIiekFxyBnhQ_LxuoUVXSwy1H30TWz5RwSC0ZUf6eZAlqgwjV00y9nMfYSkH0IxEKsyBMJamSPbogNDeT6QOGe-t9i_4yo662LP2KcbRwEpK5Sah033ZDuv--xI4gOnFovxMrM7ELz-37N59MZ3jP4DbzODXLWhdwiVPY0GYj74FrLFChA98Ov6CvVyTwqQsSl7e2No2z0z_E54E9mycwHK0uwmPEIKj2NZN47b-I8eCqWKE0wUEo_ZYlbL0ysguCQdP1iocC-TqEJoKu1WwJY97oRQ",
"e": "AQAB"
}
]
}`

func TestBuildJWKS(t *testing.T) {
	doc, keys, err := BuildJWKS(saPublicKeyPEM)
	if err != nil {
		t.Fatalf("BuildJWKS: %v", err)
	}

	if doc != wantJSON {
		t.Errorf("JWKS JSON mismatch:\n got: %q\nwant: %q", doc, wantJSON)
	}

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	k := keys[0]
	if k.Kid != wantKid {
		t.Errorf("kid = %q, want %q", k.Kid, wantKid)
	}
	if k.N != wantN {
		t.Errorf("n = %q, want %q", k.N, wantN)
	}
	if k.E != wantE {
		t.Errorf("e = %q, want %q", k.E, wantE)
	}
	if k.Kty != "RSA" || k.Use != "sig" || k.Alg != "RS256" {
		t.Errorf("kty/use/alg = %q/%q/%q, want RSA/sig/RS256", k.Kty, k.Use, k.Alg)
	}
}

// TestKidMatchesIndependentComputation re-derives the kid straight from the PEM
// using the documented algorithm (base64url-raw of SHA256 of the PKIX DER) and
// checks BuildJWKS agrees, guarding against the encoding drifting.
func TestKidMatchesIndependentComputation(t *testing.T) {
	block, _ := pem.Decode([]byte(saPublicKeyPEM))
	if block == nil {
		t.Fatal("failed to decode PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("ParsePKIXPublicKey: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	h := crypto.SHA256.New()
	h.Write(der)
	want := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	_, keys, err := BuildJWKS(saPublicKeyPEM)
	if err != nil {
		t.Fatalf("BuildJWKS: %v", err)
	}
	if keys[0].Kid != want {
		t.Errorf("kid = %q, want %q", keys[0].Kid, want)
	}
}

// TestAcceptsConventionalPublicKeyHeader ensures a "PUBLIC KEY"-labeled block
// (not kОps' "RSA PUBLIC KEY") parses to the same result.
func TestAcceptsConventionalPublicKeyHeader(t *testing.T) {
	relabeled := strings.ReplaceAll(saPublicKeyPEM, "RSA PUBLIC KEY", "PUBLIC KEY")

	_, keys, err := BuildJWKS(relabeled)
	if err != nil {
		t.Fatalf("BuildJWKS (PUBLIC KEY header): %v", err)
	}
	if keys[0].Kid != wantKid {
		t.Errorf("kid = %q, want %q", keys[0].Kid, wantKid)
	}
}

// TestEmptyInput returns an error rather than an empty document.
func TestEmptyInput(t *testing.T) {
	if _, _, err := BuildJWKS(""); err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

// TestJSONIsValidAndWellFormed sanity-checks that the emitted document parses
// back into the expected structure.
func TestJSONIsValidAndWellFormed(t *testing.T) {
	doc, _, err := BuildJWKS(saPublicKeyPEM)
	if err != nil {
		t.Fatalf("BuildJWKS: %v", err)
	}
	var parsed struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal([]byte(doc), &parsed); err != nil {
		t.Fatalf("emitted JWKS is not valid JSON: %v", err)
	}
	if len(parsed.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(parsed.Keys))
	}
	for _, field := range []string{"use", "kty", "kid", "alg", "n", "e"} {
		if _, ok := parsed.Keys[0][field]; !ok {
			t.Errorf("JWK missing field %q", field)
		}
	}
}
