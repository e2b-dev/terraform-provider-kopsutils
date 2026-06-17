// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package jwks builds the OIDC JSON Web Key Set (the /openid/v1/jwks document)
// from a cluster's service-account public key(s), byte-identically to kОps.
//
// kОps derives the JWKS purely from the service-account signing keypair's public
// key (kops/pkg/model/issuerdiscovery.go, OIDCKeys.Open):
//
//   - kid = base64url-raw( SHA256( PKIX-DER(pubkey) ) )
//   - n, e are the RSA modulus/exponent base64url-encoded by go-jose
//   - kty=RSA, alg=RS256, use=sig
//   - keys are sorted by kid and marshaled with json.MarshalIndent(_, "", "")
//
// This package reproduces that exactly (same go-jose library, same hashing and
// encoding), so the result matches `kubectl get --raw /openid/v1/jwks`.
package jwks

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"

	jose "github.com/go-jose/go-jose/v4"
)

// KeyFields is the per-key structured view of a JWK, exposing the individual
// fields a caller may want without re-parsing the JSON.
type KeyFields struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// BuildJWKS parses one or more PEM blocks containing the service-account public
// key(s) and returns:
//
//   - jsonDoc: the full {"keys":[...]} JWKS document, byte-identical to kОps'
//     output (json.MarshalIndent with empty prefix and indent).
//   - keys: the structured per-key fields (kid, kty, use, alg, n, e).
//
// kОps writes the service-account public key with a "RSA PUBLIC KEY" PEM header
// even though the body is standard PKIX DER (upup/pkg/fi/ca.go ToPublicKeys), so
// this accepts both "RSA PUBLIC KEY" and "PUBLIC KEY" blocks; the body is always
// parsed as PKIX.
func BuildJWKS(serviceAccountPublicKeysPEM string) (string, []KeyFields, error) {
	joseKeys, err := parseKeys(serviceAccountPublicKeysPEM)
	if err != nil {
		return "", nil, err
	}
	if len(joseKeys) == 0 {
		return "", nil, fmt.Errorf("no public key PEM blocks found in input")
	}

	// kОps sorts the keys by kid before marshaling.
	sort.Slice(joseKeys, func(i, j int) bool {
		return joseKeys[i].KeyID < joseKeys[j].KeyID
	})

	keyResponse := struct {
		Keys []jose.JSONWebKey `json:"keys"`
	}{Keys: joseKeys}

	jsonBytes, err := json.MarshalIndent(keyResponse, "", "")
	if err != nil {
		return "", nil, fmt.Errorf("marshaling JWKS: %w", err)
	}

	fields, err := toFields(joseKeys)
	if err != nil {
		return "", nil, err
	}

	return string(jsonBytes), fields, nil
}

// parseKeys decodes every PEM block in the input into a jose.JSONWebKey with the
// kОps-computed kid and the RS256/sig metadata.
func parseKeys(pemData string) ([]jose.JSONWebKey, error) {
	var keys []jose.JSONWebKey
	rest := []byte(pemData)

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		// kОps labels the SA public key "RSA PUBLIC KEY" but the body is PKIX
		// DER; also accept the conventional "PUBLIC KEY" label.
		if block.Type != "RSA PUBLIC KEY" && block.Type != "PUBLIC KEY" {
			continue
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parsing PKIX public key (PEM type %q): %w", block.Type, err)
		}

		kid, err := keyID(pub)
		if err != nil {
			return nil, err
		}

		keys = append(keys, jose.JSONWebKey{
			Key:       pub,
			KeyID:     kid,
			Algorithm: string(jose.RS256),
			Use:       "sig",
		})
	}

	return keys, nil
}

// keyID computes kОps' kid: base64url-raw( SHA256( PKIX-DER(pubkey) ) ).
func keyID(pub any) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("serializing public key to DER: %w", err)
	}
	hasher := crypto.SHA256.New()
	hasher.Write(der)
	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil)), nil
}

// toFields reads each jose key's JSON back into KeyFields so the n/e values use
// the identical base64url encoding go-jose produced in the JWKS document.
func toFields(keys []jose.JSONWebKey) ([]KeyFields, error) {
	fields := make([]KeyFields, 0, len(keys))
	for _, k := range keys {
		b, err := k.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("marshaling JWK %q: %w", k.KeyID, err)
		}
		var f KeyFields
		if err := json.Unmarshal(b, &f); err != nil {
			return nil, fmt.Errorf("unmarshaling JWK %q: %w", k.KeyID, err)
		}
		fields = append(fields, f)
	}
	return fields, nil
}
