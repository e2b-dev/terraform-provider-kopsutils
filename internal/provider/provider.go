// Copyright FoundryLabs, Inc. 2026
// SPDX-License-Identifier: Apache-2.0

// Package provider implements the kopsutils Terraform provider.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &KopsUtilsProvider{}

// KopsUtilsProvider defines the provider implementation.
type KopsUtilsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name. This is the prefix for every data
// source (e.g. "kopsutils_limited_length_name") and matches the source-address
// type "e2b-dev/kopsutils".
func (p *KopsUtilsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kopsutils"
	resp.Version = p.version
}

// Schema defines the provider-level configuration schema. The provider takes no
// configuration; its data sources are pure computations (except oci_reference,
// which queries registries directly).
func (p *KopsUtilsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Exposes kОps name-truncation helpers (LimitedLengthName, cluster HashString), OCI image digest resolution, and OIDC service-account JWKS generation as data sources for use in HCL.",
	}
}

// Configure is a no-op; the provider needs no client.
func (p *KopsUtilsProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

// DataSources defines the data sources implemented in the provider.
func (p *KopsUtilsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLimitedLengthNameDataSource,
		NewClusterHashDataSource,
		NewOCIReferenceDataSource,
		NewServiceAccountJWKSDataSource,
	}
}

// Resources defines the resources implemented in the provider (none).
func (p *KopsUtilsProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

// New is a helper for the provider server and acceptance tests.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KopsUtilsProvider{
			version: version,
		}
	}
}
