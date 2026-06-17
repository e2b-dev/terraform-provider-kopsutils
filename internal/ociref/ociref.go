// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package ociref resolves OCI image references to immutable, digest-pinned
// references. It mirrors how kОps pins images: crane.Digest with the default
// keychain (see kops/pkg/assets/builder.go), which by default (no platform)
// returns the top-level manifest-list/index digest. When a platform is given,
// crane.Digest resolves to the digest of the matching per-platform child
// manifest from a multi-arch index.
package ociref

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ParsePlatform parses an "os/arch[/variant]" string (e.g. "linux/arm64/v8")
// into a v1.Platform.
func ParsePlatform(platform string) (*v1.Platform, error) {
	return v1.ParsePlatform(platform)
}

// Reference builds the canonical, digest-pinned reference string, e.g.
// "registry.k8s.io/kops/kops-controller:1.35.1@sha256:434b...".
func Reference(image, tag, digest string) string {
	return image + ":" + tag + "@" + digest
}

// ResolveDigest queries the registry for the digest of image:tag, exactly as
// kОps does (crane.Digest + DefaultKeychain). It returns the "sha256:..."
// digest string.
//
// When platform is empty, no platform is applied and the returned digest is the
// top-level manifest digest (the multi-arch index/manifest-list digest for
// multi-arch images). This matches kОps' image pinning in
// kops/pkg/assets/builder.go, which calls crane.Digest without a platform.
//
// When platform is non-empty (e.g. "linux/arm64/v8") and the image is a
// multi-arch index, the returned digest is that of the platform-specific child
// manifest selected from the index.
func ResolveDigest(image, tag, platform string) (string, error) {
	if strings.TrimSpace(image) == "" {
		return "", fmt.Errorf("image must not be empty")
	}
	if strings.TrimSpace(tag) == "" {
		return "", fmt.Errorf("tag must not be empty")
	}

	ref := image + ":" + tag

	opts := []crane.Option{crane.WithAuthFromKeychain(authn.DefaultKeychain)}

	if strings.TrimSpace(platform) != "" {
		plat, err := ParsePlatform(platform)
		if err != nil {
			return "", fmt.Errorf("invalid platform %q: %w", platform, err)
		}
		opts = append(opts, crane.WithPlatform(plat))
	}

	digest, err := crane.Digest(ref, opts...)
	if err != nil {
		return "", fmt.Errorf("resolving digest for %q (platform %q): %w", ref, platform, err)
	}

	return digest, nil
}
