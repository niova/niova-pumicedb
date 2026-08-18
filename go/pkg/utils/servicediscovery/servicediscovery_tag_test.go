package servicediscovery

import (
	"testing"

	client "github.com/hashicorp/serf/client"
)

// isValidNodeData is the single gate every client passes through to reach a
// server, so the accepted "Type" tag is worth pinning down. Services must not be
// interchangeable: pickServer evicts non-matching members from its table as it
// scans, so a client that accepts the wrong tag does not just pick a bad server,
// it can exhaust its member list and fail hard.
func TestIsValidNodeDataMatchesExpectedServiceType(t *testing.T) {
	alive := func(serviceType string) client.Member {
		return client.Member{
			Status: "alive",
			Tags:   map[string]string{"Type": serviceType, "Hport": "8081"},
		}
	}

	tests := []struct {
		name string
		// clientTag is ServiceDiscoveryHandler.ServiceTypeTag; empty exercises
		// the control-plane default.
		clientTag string
		member    client.Member
		want      bool
	}{
		{
			name:   "default client accepts the control plane",
			member: alive(ServiceTypeNiovaMdsvc),
			want:   true,
		},
		{
			name:      "explicit control-plane client accepts the control plane",
			clientTag: ServiceTypeNiovaMdsvc,
			member:    alive(ServiceTypeNiovaMdsvc),
			want:      true,
		},
		{
			name:      "niovaKV client accepts NKV_proxy",
			clientTag: ServiceTypeNiovaKV,
			member:    alive(ServiceTypeNiovaKV),
			want:      true,
		},
		// The two services must not be interchangeable — this is what the shared
		// "PROXY" tag used to break.
		{
			name:   "default client rejects NKV_proxy",
			member: alive(ServiceTypeNiovaKV),
			want:   false,
		},
		{
			name:      "niovaKV client rejects the control plane",
			clientTag: ServiceTypeNiovaKV,
			member:    alive(ServiceTypeNiovaMdsvc),
			want:      false,
		},
		// The retired tag must not be honoured by anyone any more.
		{
			name:   "default client rejects the retired PROXY tag",
			member: alive("PROXY"),
			want:   false,
		},
		{
			name:      "niovaKV client rejects the retired PROXY tag",
			clientTag: ServiceTypeNiovaKV,
			member:    alive("PROXY"),
			want:      false,
		},
		{
			name:   "unrelated service type rejected",
			member: alive("PMDB_SERVER"),
			want:   false,
		},
		{
			name:   "empty type rejected",
			member: client.Member{Status: "alive", Tags: map[string]string{"Hport": "8081"}},
			want:   false,
		},
		{
			// Hport is how the caller reaches the endpoint; without it the member
			// is useless even with a matching tag.
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
			handler := &ServiceDiscoveryHandler{ServiceTypeTag: tt.clientTag}
			if got := handler.isValidNodeData(tt.member); got != tt.want {
				t.Errorf("isValidNodeData(%+v) with ServiceTypeTag=%q = %v, want %v",
					tt.member, tt.clientTag, got, tt.want)
			}
		})
	}
}

// An unset ServiceTypeTag must mean the control plane, since every existing
// control-plane caller relies on the zero value.
func TestExpectedServiceTypeDefaultsToControlPlane(t *testing.T) {
	if got := (&ServiceDiscoveryHandler{}).expectedServiceType(); got != ServiceTypeNiovaMdsvc {
		t.Errorf("expectedServiceType() = %q, want %q", got, ServiceTypeNiovaMdsvc)
	}
	if got := (&ServiceDiscoveryHandler{ServiceTypeTag: ServiceTypeNiovaKV}).expectedServiceType(); got != ServiceTypeNiovaKV {
		t.Errorf("expectedServiceType() = %q, want %q", got, ServiceTypeNiovaKV)
	}
}
