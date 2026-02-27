package tunnel_test

import (
	"devserve/tunnel"
	"fmt"
	"testing"
)

func TestGetTailscaleInfo(t *testing.T) {
	fakeJSON := `{
		"TailscaleIPs": ["100.98.150.14", "fd7a:115c:a1e0::1"],
		"Self": {
			"DNSName": "myhost.example.ts.net."
		}
	}`

	info, err := tunnel.GetTailscaleInfo(func() ([]byte, error) {
		return []byte(fakeJSON), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Hostname != "myhost.example.ts.net" {
		t.Errorf("expected hostname %q, got %q", "myhost.example.ts.net", info.Hostname)
	}
	if info.IP != "100.98.150.14" {
		t.Errorf("expected IP %q, got %q", "100.98.150.14", info.IP)
	}
}

func TestGetTailscaleInfoError(t *testing.T) {
	_, err := tunnel.GetTailscaleInfo(func() ([]byte, error) {
		return nil, fmt.Errorf("tailscale not running")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
