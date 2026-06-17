// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

package ociref

import "testing"

func TestParsePlatform_Variant(t *testing.T) {
	plat, err := ParsePlatform("linux/arm64/v8")
	if err != nil {
		t.Fatalf("ParsePlatform error: %v", err)
	}
	if plat.OS != "linux" || plat.Architecture != "arm64" || plat.Variant != "v8" {
		t.Fatalf("ParsePlatform = %+v, want linux/arm64/v8", plat)
	}
}

func TestParsePlatform_OSArch(t *testing.T) {
	plat, err := ParsePlatform("linux/amd64")
	if err != nil {
		t.Fatalf("ParsePlatform error: %v", err)
	}
	if plat.OS != "linux" || plat.Architecture != "amd64" {
		t.Fatalf("ParsePlatform = %+v, want linux/amd64", plat)
	}
}

func TestReference(t *testing.T) {
	got := Reference(
		"registry.k8s.io/kops/kops-controller",
		"1.35.1",
		"sha256:434b3dd3f9bffce98a1cfb148feb395c672b8bf5280a8a921ee1ca69bee1e239",
	)
	want := "registry.k8s.io/kops/kops-controller:1.35.1@sha256:434b3dd3f9bffce98a1cfb148feb395c672b8bf5280a8a921ee1ca69bee1e239"
	if got != want {
		t.Fatalf("Reference = %q, want %q", got, want)
	}
}

func TestResolveDigest_EmptyInputs(t *testing.T) {
	if _, err := ResolveDigest("", "1.35.1", ""); err == nil {
		t.Fatal("ResolveDigest with empty image: expected error, got nil")
	}
	if _, err := ResolveDigest("registry.k8s.io/kops/kops-controller", "", ""); err == nil {
		t.Fatal("ResolveDigest with empty tag: expected error, got nil")
	}
}

func TestResolveDigest_InvalidPlatform(t *testing.T) {
	if _, err := ResolveDigest("registry.k8s.io/kops/kops-controller", "1.35.1", "not a platform!!"); err == nil {
		t.Fatal("ResolveDigest with invalid platform: expected error, got nil")
	}
}
