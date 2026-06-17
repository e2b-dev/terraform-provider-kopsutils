# terraform-provider-kopsutils

A small Terraform/OpenTofu provider that exposes kОps' name-truncation helpers,
OCI image digest resolution, and OIDC service-account JWKS generation as data
sources, so HCL can compute kОps-equivalent values natively instead of relying
on precomputed inputs.

The truncation kОps applies (an `fnv32a` hash, `base32hex`-encoded and
truncated) cannot be expressed in pure HCL. The `gce-cluster` module uses this
provider's `kopsutils_limited_length_name` data source to compute the
instance-template `name_prefix` directly and the `kopsutils_cluster_hash` data
source to compute the service-account truncation hash.

The logic in `internal/truncate`, `internal/jwks`, and `internal/ociref` is a
byte-for-byte port of the corresponding kОps helpers, verified against golden
values in the package tests.

## Data sources

### `kopsutils_limited_length_name`

Computes `gce.LimitedLengthName(input, max_length)`.

| field        | type   | role     | description                                              |
| ------------ | ------ | -------- | -------------------------------------------------------- |
| `input`      | string | required | name to (possibly) truncate                              |
| `max_length` | number | required | maximum length of the result                             |
| `result`     | string | computed | `input` unchanged if it fits, else `base[:max-7]-<hash6>` |

### `kopsutils_cluster_hash`

Computes `lower(base32hex(fnv32a(SafeClusterName(cluster_name))))[:length]`.

| field          | type   | role     | description                                  |
| -------------- | ------ | -------- | -------------------------------------------- |
| `cluster_name` | string | required | dotted cluster name, e.g. `k8s1.example.dev` |
| `length`       | number | optional | hash chars to keep (default `6`)             |
| `result`       | string | computed | the cluster name hash                        |

### `kopsutils_oci_reference`

Resolves an OCI image to an immutable, digest-pinned reference by querying the
registry (the same way kОps pins images: `crane.Digest` with the default
keychain).

| field       | type   | role     | description                                       |
| ----------- | ------ | -------- | ------------------------------------------------- |
| `image`     | string | required | repository without tag/digest                     |
| `tag`       | string | required | image tag                                         |
| `platform`  | string | optional | `os/arch[/variant]`; omit for the index digest    |
| `digest`    | string | computed | resolved manifest digest (`sha256:...`)           |
| `reference` | string | computed | fully-pinned `image:tag@digest`                   |

### `kopsutils_service_account_jwks`

Builds the cluster's OIDC JSON Web Key Set (the `/openid/v1/jwks` document) from
the service-account public key(s), byte-identically to kОps.

| field                         | type   | role     | description                              |
| ----------------------------- | ------ | -------- | ---------------------------------------- |
| `service_account_public_keys` | string | required | service-account public key PEM(s)        |
| `json`                        | string | computed | the full JWKS document                   |
| `keys`                        | list   | computed | structured JWK entries (kid/kty/use/alg/n/e) |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

> The provider depends on `terraform-plugin-framework`, whose module graph
> requires `go >= 1.25`.

## Building the Provider

```shell
go install
```

This builds the provider and puts the binary in the `$GOPATH/bin` directory.

## Developing the Provider

To generate or update documentation, run `make generate`.

To run the unit tests:

```shell
go test ./...
```

In order to run the full suite of acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.
