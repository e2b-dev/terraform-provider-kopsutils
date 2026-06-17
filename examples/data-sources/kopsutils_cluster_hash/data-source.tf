# Cluster name hash: lower(base32hex(fnv32a(SafeClusterName(cluster_name))))[:length].
data "kopsutils_cluster_hash" "this" {
  cluster_name = "k8s1.tomas-virgl.e2b-test.dev"
}

output "cluster_name_hash" {
  value = data.kopsutils_cluster_hash.this.result # -> g73ca7
}
