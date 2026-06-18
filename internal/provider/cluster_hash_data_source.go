// Copyright FoundryLabs, Inc. 2026
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/e2b-dev/terraform-provider-kopsutils/internal/truncate"
)

var _ datasource.DataSource = &clusterHashDataSource{}

// NewClusterHashDataSource constructs the data source.
func NewClusterHashDataSource() datasource.DataSource {
	return &clusterHashDataSource{}
}

type clusterHashDataSource struct{}

type clusterHashModel struct {
	ClusterName types.String `tfsdk:"cluster_name"`
	Length      types.Int64  `tfsdk:"length"`
	Result      types.String `tfsdk:"result"`
}

func (d *clusterHashDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_hash"
}

func (d *clusterHashDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Computes the kОps cluster name hash: lower(base32hex(fnv32a(SafeClusterName(cluster_name))))[:length], where SafeClusterName replaces '.' with '-'. This is the value kОps appends to truncated names (e.g. service-account account_id).",
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Description: "The cluster name (dotted form, e.g. k8s1.example.dev).",
				Required:    true,
			},
			"length": schema.Int64Attribute{
				Description: "Number of hash characters to keep (kОps uses 6). Defaults to 6.",
				Optional:    true,
			},
			"result": schema.StringAttribute{
				Description: "The cluster name hash.",
				Computed:    true,
			},
		},
	}
}

func (d *clusterHashDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clusterHashModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	length := int64(6)
	if !data.Length.IsNull() && !data.Length.IsUnknown() {
		length = data.Length.ValueInt64()
	}
	data.Length = types.Int64Value(length)

	data.Result = types.StringValue(
		truncate.HashString(truncate.SafeClusterName(data.ClusterName.ValueString()), int(length)),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
