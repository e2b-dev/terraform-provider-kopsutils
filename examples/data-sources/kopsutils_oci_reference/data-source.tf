# Resolve an OCI image to a digest-pinned reference (queries the registry, the
# same way kОps pins images). For multi-arch images the digest is the
# platform-specific child manifest when a platform is given.
data "kopsutils_oci_reference" "kops_controller" {
  image    = "registry.k8s.io/kops/kops-controller"
  tag      = "1.35.1"
  platform = "linux/arm64/v8"
}

output "kops_controller_reference" {
  # -> registry.k8s.io/kops/kops-controller:1.35.1@sha256:...
  value = data.kopsutils_oci_reference.kops_controller.reference
}
