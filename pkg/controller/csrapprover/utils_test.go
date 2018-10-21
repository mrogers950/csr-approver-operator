package csrapprover

import (
	"testing"

	certapi "k8s.io/api/certificates/v1beta1"
)

func TestIsSubset(t *testing.T) {
	tests := map[string]struct {
		subset   []string
		set      []string
		expected bool
	}{
		"same": {
			subset:   []string{"foo"},
			set:      []string{"foo"},
			expected: true,
		},
		"samemulti": {
			subset:   []string{"foo", "bar"},
			set:      []string{"foo", "bar"},
			expected: true,
		},
		"none": {
			subset:   []string{},
			set:      []string{"foo"},
			expected: true,
		},
		"single": {
			subset:   []string{"foo"},
			set:      []string{},
			expected: false,
		},
		"less": {
			subset:   []string{"foo"},
			set:      []string{"foo", "bar"},
			expected: true,
		},
		"more": {
			subset:   []string{"foo", "bar", "baz"},
			set:      []string{"foo", "bar"},
			expected: false,
		},
		"differ1": {
			subset:   []string{"bar", "baz"},
			set:      []string{"foo", "bar"},
			expected: false,
		},
		"differ2": {
			subset:   []string{"foo", "baz"},
			set:      []string{"foo", "bar"},
			expected: false,
		},
	}
	for name, tc := range tests {
		if result := subset(tc.subset, tc.set); result != tc.expected {
			t.Fatalf("test %s failed: expected %s, got %s", name, tc.expected, result)
		}
	}
}

func TestUsagesAreAllowed(t *testing.T) {
	tests := map[string]struct {
		given    []certapi.KeyUsage
		allowed  []certapi.KeyUsage
		expected bool
	}{
		"nil": {
			allowed:  nil,
			given:    nil,
			expected: true,
		},
		"empty": {
			given:    []certapi.KeyUsage{},
			allowed:  []certapi.KeyUsage{},
			expected: true,
		},
		"none given": {
			given: []certapi.KeyUsage{},
			allowed: []certapi.KeyUsage{
				certapi.UsageSigning,
			},
			expected: true,
		},
		"none allowed": {
			given: []certapi.KeyUsage{
				certapi.UsageSigning,
			},
			allowed:  []certapi.KeyUsage{},
			expected: true,
		},
		"subset given": {
			given: []certapi.KeyUsage{
				certapi.UsageSigning,
			},
			allowed: []certapi.KeyUsage{
				certapi.UsageSigning,
				certapi.UsageDigitalSignature,
			},
			expected: true,
		},
		"given exact": {
			given: []certapi.KeyUsage{
				certapi.UsageSigning,
				certapi.UsageDigitalSignature,
			},
			allowed: []certapi.KeyUsage{
				certapi.UsageSigning,
				certapi.UsageDigitalSignature,
			},
			expected: true,
		},
		"given more": {
			given: []certapi.KeyUsage{
				certapi.UsageSigning,
				certapi.UsageDigitalSignature,
				certapi.UsageCertSign,
			},
			allowed: []certapi.KeyUsage{
				certapi.UsageSigning,
				certapi.UsageDigitalSignature,
			},
			expected: false,
		},
	}
	for name, tc := range tests {
		if name != "given more" {
			continue
		}
		result := usageSubset(tc.given, tc.allowed)
		if result != tc.expected {
			t.Fatalf("test %s failed: Expected %v, got %v", name, tc.expected, result)
		}
	}
}
