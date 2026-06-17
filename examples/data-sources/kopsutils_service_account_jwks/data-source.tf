# Build the cluster's OIDC JSON Web Key Set (the /openid/v1/jwks document) from
# the service-account public key(s), byte-identically to kОps.
data "kopsutils_service_account_jwks" "this" {
  service_account_public_keys = file("${path.module}/service-account.pub")
}

output "jwks_json" {
  value = data.kopsutils_service_account_jwks.this.json
}

output "jwks_kids" {
  value = [for k in data.kopsutils_service_account_jwks.this.keys : k.kid]
}
