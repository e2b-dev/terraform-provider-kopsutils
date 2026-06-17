// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/e2b-dev/terraform-provider-kopsutils/internal/jwks"
)

var _ datasource.DataSource = &serviceAccountJWKSDataSource{}

// NewServiceAccountJWKSDataSource constructs the data source.
func NewServiceAccountJWKSDataSource() datasource.DataSource {
	return &serviceAccountJWKSDataSource{}
}

type serviceAccountJWKSDataSource struct{}

type serviceAccountJWKSModel struct {
	ServiceAccountPublicKeys types.String   `tfsdk:"service_account_public_keys"`
	JSON                     types.String   `tfsdk:"json"`
	Keys                     []jwksKeyModel `tfsdk:"keys"`
}

type jwksKeyModel struct {
	Kid types.String `tfsdk:"kid"`
	Kty types.String `tfsdk:"kty"`
	Use types.String `tfsdk:"use"`
	Alg types.String `tfsdk:"alg"`
	N   types.String `tfsdk:"n"`
	E   types.String `tfsdk:"e"`
}

func (d *serviceAccountJWKSDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_jwks"
}

func (d *serviceAccountJWKSDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Builds the cluster's OIDC JSON Web Key Set (the /openid/v1/jwks document) from the service-account public key(s), byte-identically to kОps: kid = base64url(SHA256(PKIX-DER(pubkey))), kty=RSA, alg=RS256, use=sig, with n/e from the RSA key. Matches `kubectl get --raw /openid/v1/jwks`.",
		Attributes: map[string]schema.Attribute{
			"service_account_public_keys": schema.StringAttribute{
				Description: "The service-account public key PEM(s). Accepts kОps' \"RSA PUBLIC KEY\"-labeled PKIX blocks as well as conventional \"PUBLIC KEY\" blocks; multiple blocks may be concatenated.",
				Required:    true,
			},
			"json": schema.StringAttribute{
				Description: "The full JWKS document ({\"keys\":[...]}), byte-identical to what the API server serves at /openid/v1/jwks.",
				Computed:    true,
			},
			"keys": schema.ListNestedAttribute{
				Description: "The structured JWK entries, sorted by kid.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kid": schema.StringAttribute{
							Description: "Key ID: base64url(SHA256(PKIX-DER(pubkey))).",
							Computed:    true,
						},
						"kty": schema.StringAttribute{
							Description: "Key type (RSA).",
							Computed:    true,
						},
						"use": schema.StringAttribute{
							Description: "Public key use (sig).",
							Computed:    true,
						},
						"alg": schema.StringAttribute{
							Description: "Signing algorithm (RS256).",
							Computed:    true,
						},
						"n": schema.StringAttribute{
							Description: "RSA modulus (base64url).",
							Computed:    true,
						},
						"e": schema.StringAttribute{
							Description: "RSA public exponent (base64url).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *serviceAccountJWKSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceAccountJWKSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jsonDoc, keys, err := jwks.BuildJWKS(data.ServiceAccountPublicKeys.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to build service-account JWKS", err.Error())
		return
	}

	data.JSON = types.StringValue(jsonDoc)
	data.Keys = make([]jwksKeyModel, 0, len(keys))
	for _, k := range keys {
		data.Keys = append(data.Keys, jwksKeyModel{
			Kid: types.StringValue(k.Kid),
			Kty: types.StringValue(k.Kty),
			Use: types.StringValue(k.Use),
			Alg: types.StringValue(k.Alg),
			N:   types.StringValue(k.N),
			E:   types.StringValue(k.E),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
