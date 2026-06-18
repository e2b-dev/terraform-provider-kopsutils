// Copyright FoundryLabs, Inc. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/e2b-dev/terraform-provider-kopsutils/internal/ociref"
)

var _ datasource.DataSource = &ociReferenceDataSource{}

// NewOCIReferenceDataSource constructs the data source.
func NewOCIReferenceDataSource() datasource.DataSource {
	return &ociReferenceDataSource{}
}

type ociReferenceDataSource struct{}

type ociReferenceModel struct {
	Image     types.String `tfsdk:"image"`
	Tag       types.String `tfsdk:"tag"`
	Platform  types.String `tfsdk:"platform"`
	Digest    types.String `tfsdk:"digest"`
	Reference types.String `tfsdk:"reference"`
}

func (d *ociReferenceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oci_reference"
}

func (d *ociReferenceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resolves an OCI image reference to an immutable, digest-pinned reference by querying the registry (the same way kОps pins images: crane.Digest with the default keychain). For multi-arch images, the digest is that of the per-platform child manifest selected from the index.",
		Attributes: map[string]schema.Attribute{
			"image": schema.StringAttribute{
				Description: "The image repository without tag or digest (e.g. registry.k8s.io/kops/kops-controller).",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "The image tag (e.g. 1.35.1).",
				Required:    true,
			},
			"platform": schema.StringAttribute{
				Description: "Target platform as os/arch[/variant] (e.g. linux/arm64/v8). If omitted, the top-level manifest digest is returned (the multi-arch index digest for multi-arch images), matching kОps' image pinning.",
				Optional:    true,
			},
			"digest": schema.StringAttribute{
				Description: "The resolved manifest digest (e.g. sha256:fb5c...). With no platform this is the top-level/index digest; with a platform it is the platform-specific child manifest digest.",
				Computed:    true,
			},
			"reference": schema.StringAttribute{
				Description: "The fully-pinned reference: image:tag@digest (e.g. registry.k8s.io/kops/kops-controller:1.35.1@sha256:434b...).",
				Computed:    true,
			},
		},
	}
}

func (d *ociReferenceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ociReferenceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var platform string
	if !data.Platform.IsNull() && !data.Platform.IsUnknown() {
		platform = data.Platform.ValueString()
	}

	digest, err := ociref.ResolveDigest(data.Image.ValueString(), data.Tag.ValueString(), platform)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve OCI image digest", err.Error())
		return
	}

	data.Digest = types.StringValue(digest)
	data.Reference = types.StringValue(ociref.Reference(data.Image.ValueString(), data.Tag.ValueString(), digest))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
