package csrapprover

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"reflect"
	"testing"

	certapi "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	csrapprover "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
)

func TestNewControllerOptions(t *testing.T) {
	tests := map[string]struct {
		inConfig          *csrapprover.CSRApproverConfig
		expectedOutConfig *controllerConfig
		expectedError     string
	}{
		"empty": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: nil,
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{},
			},
			expectedError: "",
		},
		"skipped-profile-name": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: []csrapprover.CSRApprovalProfile{
					{
						Name:         "",
						AllowedNames: []string{"foobar"},
					},
				},
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{},
			},
			expectedError: "",
		},
		"multiple-profile-skipped-name": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: []csrapprover.CSRApprovalProfile{
					{
						Name:         "",
						AllowedNames: []string{"foobar"},
					},
					{
						Name: "foo",
					},
				},
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{
					"foo": {
						allowedNames:    nil,
						allowedUsages:   nil,
						allowedUsers:    nil,
						allowedGroups:   nil,
						allowedSubjects: nil,
					},
				},
			},
			expectedError: "",
		},
		"duplicate-profile-names": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: []csrapprover.CSRApprovalProfile{
					{
						Name:         "foo",
						AllowedNames: []string{"foobar"},
					},
					{
						Name:         "foo",
						AllowedNames: []string{"foobar"},
					},
				},
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{},
			},
			expectedError: "Duplicate allow profiles configured: \"foo\"",
		},
		"bad-keyusage": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: []csrapprover.CSRApprovalProfile{
					{
						Name: "foo",
						AllowedUsages: []string{
							"skeleton key",
						},
					},
				},
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{},
			},
			expectedError: "Not a supported certificate Usage: \"skeleton key\"",
		},
		"good-profiles": {
			inConfig: &csrapprover.CSRApproverConfig{
				TypeMeta: v1.TypeMeta{},
				Profiles: []csrapprover.CSRApprovalProfile{
					{
						Name: "foo",
						AllowedNames: []string{
							"foo1.example.com",
							"foo2.example.com",
						},
						AllowedSubjects: []string{
							"CN=foo1",
							"CN=foo2",
						},
						AllowedUsages: []string{
							"client auth",
							"digital signature",
						},
						AllowedUsers: []string{
							"foo1",
							"foo2",
						},
						AllowedGroups: []string{
							"foo1",
							"foo2",
						},
					},
					{
						Name: "bar",
						AllowedNames: []string{
							"bar1.example.com",
							"bar2.example.com",
						},
						AllowedSubjects: []string{
							"CN=bar1",
							"CN=bar2",
						},
						AllowedUsages: []string{
							"server auth",
							"data encipherment",
						},
						AllowedUsers: []string{
							"bar1",
							"bar2",
						},
						AllowedGroups: []string{
							"bar1",
							"bar2",
						},
					},
				},
			},
			expectedOutConfig: &controllerConfig{
				profiles: map[string]permissionProfile{
					"foo": {
						allowedNames: []string{
							"foo1.example.com",
							"foo2.example.com",
						},
						allowedUsages: []certapi.KeyUsage{
							certapi.KeyUsage("client auth"),
							certapi.KeyUsage("digital signature"),
						},
						allowedUsers: []string{
							"foo1",
							"foo2",
						},
						allowedGroups: []string{
							"foo1",
							"foo2",
						},
						allowedSubjects: []string{
							"CN=foo1",
							"CN=foo2",
						},
					},
					"bar": {
						allowedNames: []string{
							"bar1.example.com",
							"bar2.example.com",
						},
						allowedUsages: []certapi.KeyUsage{
							certapi.KeyUsage("server auth"),
							certapi.KeyUsage("data encipherment"),
						},
						allowedUsers: []string{
							"bar1",
							"bar2",
						},
						allowedGroups: []string{
							"bar1",
							"bar2",
						},
						allowedSubjects: []string{
							"CN=bar1",
							"CN=bar2",
						},
					},
				},
			},
			expectedError: "",
		},
	}

	for test, tc := range tests {
		opts, err := NewControllerOptions(tc.inConfig)
		if err != nil {
			if tc.expectedError == "" {
				t.Errorf("%s: unexpected error %v", test, err)
			} else {
				if err.Error() != tc.expectedError {
					t.Errorf("%s: expected error %v, got %v", test, tc.expectedError, err)
				}
			}
		} else {
			if !reflect.DeepEqual(opts, tc.expectedOutConfig) {
				t.Errorf("%s: expected output %#v, got %#v", test, tc.expectedOutConfig, opts)
			}
		}
	}
}

func TestAllowedByProfile(t *testing.T) {
	tests := map[string]struct {
		profiles       map[string]permissionProfile
		spec           certapi.CertificateSigningRequestSpec
		csr            *x509.CertificateRequest
		expectedResult bool
	}{
		"blank": {
			profiles:       map[string]permissionProfile{},
			spec:           certapi.CertificateSigningRequestSpec{},
			csr:            nil,
			expectedResult: false,
		},
		"blank-deny": {
			profiles: map[string]permissionProfile{},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("server auth"),
				},
			},
			csr: &x509.CertificateRequest{
				DNSNames: []string{"foo"},
			},
			expectedResult: false,
		},
		"insecure-auto-approve": {
			profiles: map[string]permissionProfile{
				InsecureProfileName: {},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("server auth"),
				},
			},
			csr: &x509.CertificateRequest{
				DNSNames: []string{"foo"},
			},
			expectedResult: true,
		},
		"restrict-name-only-ok": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedNames: []string{
						"foo", "bar",
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{},
			csr: &x509.CertificateRequest{
				DNSNames: []string{
					"foo",
				},
			},
			expectedResult: true,
		},
		"restrict-name-only-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedNames: []string{
						"foo", "bar",
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{},
			csr: &x509.CertificateRequest{
				DNSNames: []string{
					"baz",
				},
			},
			expectedResult: false,
		},
		"restrict-name-only-extra-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedNames: []string{
						"foo", "bar",
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{},
			csr: &x509.CertificateRequest{
				DNSNames: []string{
					"foo", "bar", "baz",
				},
			},
			expectedResult: false,
		},
		"restrict-usage-only-ok": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("server auth"),
				},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: true,
		},
		"restrict-usage-only-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("client auth"),
				},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: false,
		},
		"restrict-usage-only-extra-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
					},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("server auth"),
					certapi.KeyUsage("client auth"),
				},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: false,
		},
		"restrict-groups-only-ok": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedGroups: []string{"foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Groups: []string{"foo"},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: true,
		},
		"restrict-groups-only-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedGroups: []string{"foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Groups: []string{"bar"},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: false,
		},
		"restrict-groups-only-extra-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedGroups: []string{"foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Groups: []string{"foo", "bar"},
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: false,
		},
		"restrict-user-only-ok": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedUsers: []string{"foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Username: "foo",
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: true,
		},
		"restrict-user-only-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedUsers: []string{"foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Username: "bar",
			},
			csr:            &x509.CertificateRequest{},
			expectedResult: false,
		},
		"restrict-subject-only-ok": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedSubjects: []string{"CN=foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{},
			csr: &x509.CertificateRequest{
				Subject: pkix.Name{CommonName: "foo"},
			},
			expectedResult: true,
		},
		"restrict-subject-only-no": {
			profiles: map[string]permissionProfile{
				"one": {
					allowedSubjects: []string{"CN=foo"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{},
			csr: &x509.CertificateRequest{
				Subject: pkix.Name{CommonName: "bar"},
			},
			expectedResult: false,
		},
		"restrict-multiple-match-one": {
			profiles: map[string]permissionProfile{
				"notmatch": {
					allowedNames: []string{"foo", "baz"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
					},
					allowedUsers:    []string{"bar"},
					allowedGroups:   []string{"foogroup"},
					allowedSubjects: []string{"CN=foo"},
				},
				"match": {
					allowedNames: []string{"bar", "bar2"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("client auth"),
					},
					allowedUsers:    []string{"bar", "bar2"},
					allowedGroups:   []string{"bargroup", "bazgroup"},
					allowedSubjects: []string{"CN=bar"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("client auth"),
				},
				Username: "bar",
				Groups:   []string{"bargroup"},
			},
			csr: &x509.CertificateRequest{
				DNSNames: []string{"bar"},
				Subject:  pkix.Name{CommonName: "bar"},
			},
			expectedResult: true,
		},
		"restrict-multiple-match-none": {
			profiles: map[string]permissionProfile{
				"notmatch": {
					allowedNames: []string{"foo", "baz"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
					},
					allowedUsers:    []string{"bar"},
					allowedGroups:   []string{"foogroup"},
					allowedSubjects: []string{"CN=foo"},
				},
				"notmatchalso": {
					allowedNames: []string{"bar", "bar2"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("client auth"),
					},
					allowedUsers:    []string{"bar", "bar2"},
					allowedGroups:   []string{"bargroup", "bazgroup"},
					allowedSubjects: []string{"CN=bar"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("client auth"),
				},
				Username: "bar",
				Groups:   []string{"bargroup"},
			},
			csr: &x509.CertificateRequest{
				DNSNames: []string{"bar"},
				Subject:  pkix.Name{CommonName: "far"},
			},
			expectedResult: false,
		},
		"restrict-multiple-match-either": {
			profiles: map[string]permissionProfile{
				"match": {
					allowedNames: []string{"foo", "baz"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("server auth"),
						certapi.KeyUsage("client auth"),
					},
					allowedUsers:  []string{"bar"},
					allowedGroups: []string{"foogroup"},
				},
				"matchalso": {
					allowedNames: []string{"bar", "foo"},
					allowedUsages: []certapi.KeyUsage{
						certapi.KeyUsage("client auth"),
						certapi.KeyUsage("data encipherment"),
					},
					allowedUsers:  []string{"bar", "bar2"},
					allowedGroups: []string{"bargroup", "bazgroup", "foogroup"},
				},
			},
			spec: certapi.CertificateSigningRequestSpec{
				Usages: []certapi.KeyUsage{
					certapi.KeyUsage("client auth"),
				},
				Username: "bar",
				Groups:   []string{"foogroup"},
			},
			csr: &x509.CertificateRequest{
				DNSNames: []string{"foo"},
				Subject:  pkix.Name{CommonName: "far"},
			},
			expectedResult: true,
		},
	}

	for test, tc := range tests {
		if allowedByProfiles(tc.profiles, tc.spec, tc.csr) != tc.expectedResult {
			t.Errorf("%s: result expected %v, got %v", test, tc.expectedResult, !tc.expectedResult)
		}
	}
}
