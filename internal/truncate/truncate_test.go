// Copyright (c) FoundryLabs, Inc.
// SPDX-License-Identifier: Apache-2.0

package truncate

import "testing"

// cluster is the example1 golden cluster name.
const cluster = "cluster.e2b.dev"

func TestHashString_ClusterHash(t *testing.T) {
	// cluster_name_hash from tests/example1/module/main.tf.
	got := HashString(SafeClusterName(cluster), 6)
	if want := "i7tm67"; got != want {
		t.Fatalf("HashString = %q, want %q", got, want)
	}
}

func TestLimitedLengthName_Passthrough(t *testing.T) {
	// Short inputs (<= n) are returned unchanged.
	if got := LimitedLengthName("nodes", 32); got != "nodes" {
		t.Fatalf("LimitedLengthName(nodes,32) = %q, want %q", got, "nodes")
	}
	if got := LimitedLengthName("", 32); got != "" {
		t.Fatalf("LimitedLengthName(\"\",32) = %q, want \"\"", got)
	}
}

func TestLimitedLengthName_GoldenIGPrefixes(t *testing.T) {
	// name_prefix values from tests/example1/golden/kubernetes.tf
	// (without the trailing dash kОps appends after LimitedLengthName).
	// kОps derives the input name as SafeClusterName(ig + "-" + cluster) and
	// then LimitedLengthName(name, 32).
	cases := []struct {
		ig   string
		want string
	}{
		{"control-plane-us-west1-a", "control-plane-us-west1-a--do4lm3"},
		{"control-plane-us-west1-b", "control-plane-us-west1-b--gk1uv6"},
		{"control-plane-us-west1-c", "control-plane-us-west1-c--83strm"},
		{"nodes-us-west1-a", "nodes-us-west1-a-cluster-e2b-dev"},
		{"nodes-us-west1-b", "nodes-us-west1-b-cluster-e2b-dev"},
		{"nodes-us-west1-c", "nodes-us-west1-c-cluster-e2b-dev"},
	}
	for _, tc := range cases {
		t.Run(tc.ig, func(t *testing.T) {
			name := SafeClusterName(tc.ig + "-" + cluster)
			got := LimitedLengthName(name, 32)
			if got != tc.want {
				t.Fatalf("LimitedLengthName(%q,32) = %q, want %q", name, got, tc.want)
			}
			if len(got) > 32 {
				t.Fatalf("result %q exceeds max length 32 (len=%d)", got, len(got))
			}
		})
	}
}
