package servicediscovery

import (
	"testing"

	client "github.com/hashicorp/serf/client"
)

// isValidNodeData is the single gate every Go consumer passes through to reach an
// app server, so the set of accepted "Type" tags is worth pinning: widening it by
// accident exposes unrelated services, narrowing it makes the fleet undiscoverable.
func TestIsValidNodeDataAcceptsCurrentAndLegacyTags(t *testing.T) {
	tests := []struct {
		name   string
		member client.Member
		want   bool
	}{
		{
			name:   "current tag",
			member: client.Member{Status: "alive", Tags: map[string]string{"Type": ServiceTypeNiovaMdsvc, "Hport": "8081"}},
			want:   true,
		},
		{
			// Keeps a not-yet-migrated publisher (and niovaKV's NKV_proxy) reachable.
			name:   "legacy tag still accepted",
			member: client.Member{Status: "alive", Tags: map[string]string{"Type": ServiceTypeLegacyProxy, "Hport": "8081"}},
			want:   true,
		},
		{
			name:   "unrelated service type rejected",
			member: client.Member{Status: "alive", Tags: map[string]string{"Type": "PMDB_SERVER", "Hport": "8081"}},
			want:   false,
		},
		{
			name:   "empty type rejected",
			member: client.Member{Status: "alive", Tags: map[string]string{"Hport": "8081"}},
			want:   false,
		},
		{
			// Hport is how the caller reaches the REST endpoint; without it the
			// member is useless even with a matching tag.
			name:   "missing Hport rejected",
			member: client.Member{Status: "alive", Tags: map[string]string{"Type": ServiceTypeNiovaMdsvc}},
			want:   false,
		},
		{
			name:   "dead member rejected",
			member: client.Member{Status: "failed", Tags: map[string]string{"Type": ServiceTypeNiovaMdsvc, "Hport": "8081"}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidNodeData(tt.member); got != tt.want {
				t.Errorf("isValidNodeData(%+v) = %v, want %v", tt.member, got, tt.want)
			}
		})
	}
}
