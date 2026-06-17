terraform {
  required_providers {
    kopsutils = {
      source = "e2b-dev/kopsutils"
    }
  }
}

# The provider takes no configuration; its data sources are pure computations
# (except kopsutils_oci_reference, which queries container registries).
provider "kopsutils" {}
