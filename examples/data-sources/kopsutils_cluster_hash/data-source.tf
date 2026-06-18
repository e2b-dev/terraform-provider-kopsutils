# Cluster name hash: lower(base32hex(fnv32a(SafeClusterName(cluster_name))))[:length].
data "kopsutils_cluster_hash" "this" {
  cluster_name = "cluster.e2b.dev"
}

output "cluster_name_hash" {
  value = data.kopsutils_cluster_hash.this.result # -> i7tm67
}
