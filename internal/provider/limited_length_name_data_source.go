// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/e2b-dev/terraform-provider-kopsutils/internal/truncate"
)

var _ datasource.DataSource = &limitedLengthNameDataSource{}

// NewLimitedLengthNameDataSource constructs the data source.
func NewLimitedLengthNameDataSource() datasource.DataSource {
	return &limitedLengthNameDataSource{}
}

type limitedLengthNameDataSource struct{}

type limitedLengthNameModel struct {
	Input     types.String `tfsdk:"input"`
	MaxLength types.Int64  `tfsdk:"max_length"`
	Result    types.String `tfsdk:"result"`
}

func (d *limitedLengthNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_limited_length_name"
}

func (d *limitedLengthNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Computes kОps' gce.LimitedLengthName(input, max_length): returns input unchanged when it fits within max_length, otherwise appends a 6-char fnv32a/base32hex hash and truncates so the result is at most max_length characters.",
		Attributes: map[string]schema.Attribute{
			"input": schema.StringAttribute{
				Description: "The name to (possibly) truncate.",
				Required:    true,
			},
			"max_length": schema.Int64Attribute{
				Description: "Maximum allowed length of the result.",
				Required:    true,
			},
			"result": schema.StringAttribute{
				Description: "The length-limited name.",
				Computed:    true,
			},
		},
	}
}

func (d *limitedLengthNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data limitedLengthNameModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Result = types.StringValue(
		truncate.LimitedLengthName(data.Input.ValueString(), int(data.MaxLength.ValueInt64())),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
