# gce.LimitedLengthName(input, max_length): long names are truncated and get a
# 6-char fnv32a/base32hex hash appended; short names are returned unchanged.
data "kopsutils_limited_length_name" "ig" {
  input      = "control-plane-us-west1-a-cluster-e2b-dev"
  max_length = 32
}

output "name_prefix" {
  value = data.kopsutils_limited_length_name.ig.result
}
